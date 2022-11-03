//go:build debug
// +build debug

package gojq

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	debug    bool
	debugOut io.Writer
)

func init() {
	if out := os.Getenv("GOJQ_DEBUG"); out != "" {
		debug = true
		if out == "stdout" {
			debugOut = os.Stdout
		} else {
			debugOut = os.Stderr
		}
	}
}

type codeinfo struct {
	name string
	pc   int
}

func (c *compiler) appendCodeInfo(x interface{}) {
	if !debug {
		return
	}
	var name string
	switch x := x.(type) {
	case string:
		name = x
	default:
		name = fmt.Sprint(x)
	}
	var diff int
	if c.codes[len(c.codes)-1] != nil && c.codes[len(c.codes)-1].op == opret && strings.HasPrefix(name, "end of ") {
		diff = -1
	}
	c.codeinfos = append(c.codeinfos, codeinfo{name, c.pc() + diff})
}

func (c *compiler) deleteCodeInfo(name string) {
	for i := 0; i < len(c.codeinfos); i++ {
		if strings.HasSuffix(c.codeinfos[i].name, name) {
			copy(c.codeinfos[i:], c.codeinfos[i+1:])
			c.codeinfos = c.codeinfos[:len(c.codeinfos)-1]
			i--
		}
	}
}

func (env *env) lookupInfoName(pc int) string {
	var name string
	for _, ci := range env.codeinfos {
		if ci.pc == pc {
			if name != "" {
				name += ", "
			}
			name += ci.name
		}
	}
	return name
}

func (env *env) debugCodes() {
	if !debug {
		return
	}
	for i, c := range env.codes {
		pc := i
		switch c.op {
		case opcall, opcallrec:
			if x, ok := c.v.(int); ok {
				pc = x
			}
		case opjump:
			x := c.v.(int)
			if x > 0 && env.codes[x-1].op == opscope {
				pc = x - 1
			}
		}
		var s string
		if name := env.lookupInfoName(pc); name != "" {
			switch c.op {
			case opcall, opcallrec, opjump:
				if !strings.HasPrefix(name, "module ") {
					s = "\t## call " + name
					break
				}
				fallthrough
			default:
				s = "\t## " + name
			}
		}
		fmt.Fprintf(debugOut, "\t%d\t%s%s%s\n", i, formatOp(c.op, false), debugOperand(c), s)
	}
	fmt.Fprintln(debugOut, "\t"+strings.Repeat("-", 40)+"+")
}

func (env *env) debugState(pc int, backtrack bool) {
	if !debug {
		return
	}
	var sb strings.Builder
	c := env.codes[pc]
	fmt.Fprintf(&sb, "\t%d\t%s%s\t|", pc, formatOp(c.op, backtrack), debugOperand(c))
	var xs []int
	for i := env.stack.index; i >= 0; i = env.stack.data[i].next {
		xs = append(xs, i)
	}
	for i := len(xs) - 1; i >= 0; i-- {
		sb.WriteString("\t")
		sb.WriteString(debugValue(env.stack.data[xs[i]].value))
	}
	switch c.op {
	case opcall, opcallrec:
		if x, ok := c.v.(int); ok {
			pc = x
		}
	case opjump:
		x := c.v.(int)
		if x > 0 && env.codes[x-1].op == opscope {
			pc = x - 1
		}
	}
	if name := env.lookupInfoName(pc); name != "" {
		switch c.op {
		case opcall, opcallrec, opjump:
			if !strings.HasPrefix(name, "module ") {
				sb.WriteString("\t\t\t## call " + name)
				break
			}
			fallthrough
		default:
			sb.WriteString("\t\t\t## " + name)
		}
	}
	fmt.Fprintln(debugOut, sb.String())
}

func formatOp(c opcode, backtrack bool) string {
	if backtrack {
		return c.String() + " <backtrack>" + strings.Repeat(" ", 13-len(c.String()))
	}
	return c.String() + strings.Repeat(" ", 25-len(c.String()))
}

func (env *env) debugForks(pc int, op string) {
	if !debug {
		return
	}
	var sb strings.Builder
	for i, v := range env.forks {
		if i > 0 {
			sb.WriteByte('\t')
		}
		if i == len(env.forks)-1 {
			sb.WriteByte('<')
		}
		fmt.Fprintf(&sb, "%d, %s", v.pc, debugValue(env.stack.data[v.stackindex].value))
		if i == len(env.forks)-1 {
			sb.WriteByte('>')
		}
	}
	fmt.Fprintf(debugOut, "\t-\t%s%s%d\t|\t%s\n", op, strings.Repeat(" ", 22), pc, sb.String())
}

func debugOperand(c *code) string {
	switch c.op {
	case opcall, opcallrec:
		switch v := c.v.(type) {
		case int:
			return strconv.Itoa(v)
		case [3]interface{}:
			return fmt.Sprintf("%s/%d", v[2], v[1])
		default:
			panic(c)
		}
	default:
		return debugValue(c.v)
	}
}

func debugValue(v interface{}) string {
	switch v := v.(type) {
	case Iter:
		return fmt.Sprintf("gojq.Iter(%#v)", v)
	case [2]int:
		return fmt.Sprintf("[%d,%d]", v[0], v[1])
	case [3]int:
		return fmt.Sprintf("[%d,%d,%d]", v[0], v[1], v[2])
	case [3]interface{}:
		return fmt.Sprintf("[%v,%v,%v]", v[0], v[1], v[2])
	default:
		return previewValue(v)
	}
}
