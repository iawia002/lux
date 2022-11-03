package gojq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type compiler struct {
	moduleLoader  ModuleLoader
	environLoader func() []string
	variables     []string
	customFuncs   map[string]function
	inputIter     Iter
	codes         []*code
	codeinfos     []codeinfo
	scopes        []*scopeinfo
	scopecnt      int
}

// Code is a compiled jq query.
type Code struct {
	variables []string
	codes     []*code
	codeinfos []codeinfo
}

// Run runs the code with the variable values (which should be in the
// same order as the given variables using WithVariables) and returns
// a result iterator.
//
// It is safe to call this method of a *Code in multiple goroutines.
func (c *Code) Run(v interface{}, values ...interface{}) Iter {
	return c.RunWithContext(context.Background(), v, values...)
}

// RunWithContext runs the code with context.
func (c *Code) RunWithContext(ctx context.Context, v interface{}, values ...interface{}) Iter {
	if len(values) > len(c.variables) {
		return NewIter(&tooManyVariableValuesError{})
	} else if len(values) < len(c.variables) {
		return NewIter(&expectedVariableError{c.variables[len(values)]})
	}
	for i, v := range values {
		values[i] = normalizeNumbers(v)
	}
	return newEnv(ctx).execute(c, normalizeNumbers(v), values...)
}

// ModuleLoader is an interface for loading modules.
//
// Implement following optional methods. Use NewModuleLoader to load local modules.
//  LoadModule(string) (*Query, error)
//  LoadModuleWithMeta(string, map[string]interface{}) (*Query, error)
//  LoadInitModules() ([]*Query, error)
//  LoadJSON(string) (interface{}, error)
//  LoadJSONWithMeta(string, map[string]interface{}) (interface{}, error)
type ModuleLoader interface{}

type scopeinfo struct {
	variables   []*varinfo
	funcs       []*funcinfo
	id          int
	depth       int
	variablecnt int
}

type varinfo struct {
	name  string
	index [2]int
	depth int
}

type funcinfo struct {
	name   string
	pc     int
	argcnt int
}

// Compile compiles a query.
func Compile(q *Query, options ...CompilerOption) (*Code, error) {
	c := &compiler{}
	for _, opt := range options {
		opt(c)
	}
	scope := c.newScope()
	c.scopes = []*scopeinfo{scope}
	setscope := c.lazy(func() *code {
		return &code{op: opscope, v: [3]int{scope.id, scope.variablecnt, 0}}
	})
	if c.moduleLoader != nil {
		if moduleLoader, ok := c.moduleLoader.(interface {
			LoadInitModules() ([]*Query, error)
		}); ok {
			qs, err := moduleLoader.LoadInitModules()
			if err != nil {
				return nil, err
			}
			for _, q := range qs {
				if err := c.compileModule(q, ""); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := c.compile(q); err != nil {
		return nil, err
	}
	setscope()
	c.optimizeTailRec()
	c.optimizeCodeOps()
	return &Code{
		variables: c.variables,
		codes:     c.codes,
		codeinfos: c.codeinfos,
	}, nil
}

func (c *compiler) compile(q *Query) error {
	for _, name := range c.variables {
		if !newLexer(name).validVarName() {
			return &variableNameError{name}
		}
		c.appendCodeInfo(name)
		c.append(&code{op: opstore, v: c.pushVariable(name)})
	}
	for _, i := range q.Imports {
		if err := c.compileImport(i); err != nil {
			return err
		}
	}
	if err := c.compileQuery(q); err != nil {
		return err
	}
	c.append(&code{op: opret})
	return nil
}

func (c *compiler) compileImport(i *Import) error {
	var path, alias string
	var err error
	if i.ImportPath != "" {
		path, alias = i.ImportPath, i.ImportAlias
	} else {
		path = i.IncludePath
	}
	if c.moduleLoader == nil {
		return fmt.Errorf("cannot load module: %q", path)
	}
	if strings.HasPrefix(alias, "$") {
		var vals interface{}
		if moduleLoader, ok := c.moduleLoader.(interface {
			LoadJSONWithMeta(string, map[string]interface{}) (interface{}, error)
		}); ok {
			if vals, err = moduleLoader.LoadJSONWithMeta(path, i.Meta.ToValue()); err != nil {
				return err
			}
		} else if moduleLoader, ok := c.moduleLoader.(interface {
			LoadJSON(string) (interface{}, error)
		}); ok {
			if vals, err = moduleLoader.LoadJSON(path); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("module not found: %q", path)
		}
		vals = normalizeNumbers(vals)
		c.append(&code{op: oppush, v: vals})
		c.append(&code{op: opstore, v: c.pushVariable(alias)})
		c.append(&code{op: oppush, v: vals})
		c.append(&code{op: opstore, v: c.pushVariable(alias + "::" + alias[1:])})
		return nil
	}
	var q *Query
	if moduleLoader, ok := c.moduleLoader.(interface {
		LoadModuleWithMeta(string, map[string]interface{}) (*Query, error)
	}); ok {
		if q, err = moduleLoader.LoadModuleWithMeta(path, i.Meta.ToValue()); err != nil {
			return err
		}
	} else if moduleLoader, ok := c.moduleLoader.(interface {
		LoadModule(string) (*Query, error)
	}); ok {
		if q, err = moduleLoader.LoadModule(path); err != nil {
			return err
		}
	}
	c.appendCodeInfo("module " + path)
	defer c.appendCodeInfo("end of module " + path)
	return c.compileModule(q, alias)
}

func (c *compiler) compileModule(q *Query, alias string) error {
	scope := c.scopes[len(c.scopes)-1]
	scope.depth++
	defer func(l int) {
		scope.depth--
		scope.variables = scope.variables[:l]
	}(len(scope.variables))
	if alias != "" {
		defer func(l int) {
			for _, f := range scope.funcs[l:] {
				f.name = alias + "::" + f.name
			}
		}(len(scope.funcs))
	}
	for _, i := range q.Imports {
		if err := c.compileImport(i); err != nil {
			return err
		}
	}
	for _, fd := range q.FuncDefs {
		if err := c.compileFuncDef(fd, false); err != nil {
			return err
		}
	}
	return nil
}

func (c *compiler) newVariable() [2]int {
	return c.createVariable("")
}

func (c *compiler) pushVariable(name string) [2]int {
	s := c.scopes[len(c.scopes)-1]
	for _, v := range s.variables {
		if v.name == name && v.depth == s.depth {
			return v.index
		}
	}
	return c.createVariable(name)
}

func (c *compiler) createVariable(name string) [2]int {
	s := c.scopes[len(c.scopes)-1]
	v := [2]int{s.id, s.variablecnt}
	s.variablecnt++
	s.variables = append(s.variables, &varinfo{name, v, s.depth})
	return v
}

func (c *compiler) lookupVariable(name string) ([2]int, error) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		s := c.scopes[i]
		for j := len(s.variables) - 1; j >= 0; j-- {
			if w := s.variables[j]; w.name == name {
				return w.index, nil
			}
		}
	}
	return [2]int{}, &variableNotFoundError{name}
}

func (c *compiler) lookupFuncOrVariable(name string) (*funcinfo, *varinfo) {
	for i, isFunc := len(c.scopes)-1, name[0] != '$'; i >= 0; i-- {
		s := c.scopes[i]
		if isFunc {
			for j := len(s.funcs) - 1; j >= 0; j-- {
				if f := s.funcs[j]; f.name == name && f.argcnt == 0 {
					return f, nil
				}
			}
		}
		for j := len(s.variables) - 1; j >= 0; j-- {
			if v := s.variables[j]; v.name == name {
				return nil, v
			}
		}
	}
	return nil, nil
}

func (c *compiler) newScope() *scopeinfo {
	i := c.scopecnt // do not use len(c.scopes) because it pops
	c.scopecnt++
	return &scopeinfo{id: i}
}

func (c *compiler) newScopeDepth() func() {
	scope := c.scopes[len(c.scopes)-1]
	l, m := len(scope.variables), len(scope.funcs)
	scope.depth++
	return func() {
		scope.depth--
		scope.variables = scope.variables[:l]
		scope.funcs = scope.funcs[:m]
	}
}

func (c *compiler) compileFuncDef(e *FuncDef, builtin bool) error {
	var scope *scopeinfo
	if builtin {
		scope = c.scopes[0]
		for i := len(scope.funcs) - 1; i >= 0; i-- {
			if f := scope.funcs[i]; f.name == e.Name && f.argcnt == len(e.Args) {
				return nil
			}
		}
	} else {
		scope = c.scopes[len(c.scopes)-1]
	}
	defer c.lazy(func() *code {
		return &code{op: opjump, v: c.pc()}
	})()
	c.appendCodeInfo(e.Name)
	defer c.appendCodeInfo("end of " + e.Name)
	pc := c.pc()
	scope.funcs = append(scope.funcs, &funcinfo{e.Name, pc, len(e.Args)})
	defer func(scopes []*scopeinfo, variables []string) {
		c.scopes, c.variables = scopes, variables
	}(c.scopes, c.variables)
	c.variables = c.variables[len(c.variables):]
	scope = c.newScope()
	if builtin {
		c.scopes = []*scopeinfo{c.scopes[0], scope}
	} else {
		c.scopes = append(c.scopes, scope)
	}
	defer c.lazy(func() *code {
		return &code{op: opscope, v: [3]int{scope.id, scope.variablecnt, len(e.Args)}}
	})()
	if len(e.Args) > 0 {
		type varIndex struct {
			name  string
			index [2]int
		}
		vis := make([]varIndex, 0, len(e.Args))
		v := c.newVariable()
		c.append(&code{op: opstore, v: v})
		for _, arg := range e.Args {
			if arg[0] == '$' {
				c.appendCodeInfo(arg[1:])
				w := c.createVariable(arg[1:])
				c.append(&code{op: opstore, v: w})
				vis = append(vis, varIndex{arg, w})
			} else {
				c.appendCodeInfo(arg)
				c.append(&code{op: opstore, v: c.createVariable(arg)})
			}
		}
		for _, w := range vis {
			c.append(&code{op: opload, v: v})
			c.append(&code{op: opload, v: w.index})
			c.append(&code{op: opcallpc})
			c.appendCodeInfo(w.name)
			c.append(&code{op: opstore, v: c.pushVariable(w.name)})
		}
		c.append(&code{op: opload, v: v})
	}
	return c.compile(e.Body)
}

func (c *compiler) compileQuery(e *Query) error {
	for _, fd := range e.FuncDefs {
		if err := c.compileFuncDef(fd, false); err != nil {
			return err
		}
	}
	if e.Func != "" {
		switch e.Func {
		case ".":
			return c.compileTerm(&Term{Type: TermTypeIdentity})
		case "..":
			return c.compileTerm(&Term{Type: TermTypeRecurse})
		case "null":
			return c.compileTerm(&Term{Type: TermTypeNull})
		case "true":
			return c.compileTerm(&Term{Type: TermTypeTrue})
		case "false":
			return c.compileTerm(&Term{Type: TermTypeFalse})
		default:
			return c.compileFunc(&Func{Name: e.Func})
		}
	} else if e.Term != nil {
		return c.compileTerm(e.Term)
	}
	switch e.Op {
	case OpPipe:
		if err := c.compileQuery(e.Left); err != nil {
			return err
		}
		return c.compileQuery(e.Right)
	case OpComma:
		return c.compileComma(e.Left, e.Right)
	case OpAlt:
		return c.compileAlt(e.Left, e.Right)
	case OpAssign, OpModify, OpUpdateAdd, OpUpdateSub,
		OpUpdateMul, OpUpdateDiv, OpUpdateMod, OpUpdateAlt:
		return c.compileQueryUpdate(e.Left, e.Right, e.Op)
	case OpOr:
		return c.compileIf(
			&If{
				Cond: e.Left,
				Then: &Query{Term: &Term{Type: TermTypeTrue}},
				Else: &Query{Term: &Term{Type: TermTypeIf, If: &If{
					Cond: e.Right,
					Then: &Query{Term: &Term{Type: TermTypeTrue}},
					Else: &Query{Term: &Term{Type: TermTypeFalse}},
				}}},
			},
		)
	case OpAnd:
		return c.compileIf(
			&If{
				Cond: e.Left,
				Then: &Query{Term: &Term{Type: TermTypeIf, If: &If{
					Cond: e.Right,
					Then: &Query{Term: &Term{Type: TermTypeTrue}},
					Else: &Query{Term: &Term{Type: TermTypeFalse}},
				}}},
				Else: &Query{Term: &Term{Type: TermTypeFalse}},
			},
		)
	default:
		return c.compileCall(
			e.Op.getFunc(),
			[]*Query{e.Left, e.Right},
		)
	}
}

func (c *compiler) compileComma(l, r *Query) error {
	setfork := c.lazy(func() *code {
		return &code{op: opfork, v: c.pc() + 1}
	})
	if err := c.compileQuery(l); err != nil {
		return err
	}
	setfork()
	defer c.lazy(func() *code {
		return &code{op: opjump, v: c.pc()}
	})()
	return c.compileQuery(r)
}

func (c *compiler) compileAlt(l, r *Query) error {
	c.append(&code{op: oppush, v: false})
	found := c.newVariable()
	c.append(&code{op: opstore, v: found})
	setfork := c.lazy(func() *code {
		return &code{op: opfork, v: c.pc()} // opload found
	})
	if err := c.compileQuery(l); err != nil {
		return err
	}
	c.append(&code{op: opdup})
	c.append(&code{op: opjumpifnot, v: c.pc() + 4}) // oppop
	c.append(&code{op: oppush, v: true})            // found some value
	c.append(&code{op: opstore, v: found})
	defer c.lazy(func() *code {
		return &code{op: opjump, v: c.pc()} // ret
	})()
	c.append(&code{op: oppop})
	c.append(&code{op: opbacktrack})
	setfork()
	c.append(&code{op: opload, v: found})
	c.append(&code{op: opjumpifnot, v: c.pc() + 3})
	c.append(&code{op: opbacktrack}) // if found, backtrack
	c.append(&code{op: oppop})
	return c.compileQuery(r)
}

func (c *compiler) compileQueryUpdate(l, r *Query, op Operator) error {
	switch op {
	case OpAssign:
		// .foo.bar = f => setpath(["foo", "bar"]; f)
		if xs := l.toIndices(); xs != nil {
			// ref: compileCall
			v := c.newVariable()
			c.append(&code{op: opstore, v: v})
			c.append(&code{op: opload, v: v})
			if err := c.compileQuery(r); err != nil {
				return err
			}
			c.append(&code{op: oppush, v: xs})
			c.append(&code{op: opload, v: v})
			c.append(&code{op: opcall, v: [3]interface{}{internalFuncs["setpath"].callback, 2, "setpath"}})
			return nil
		}
		fallthrough
	case OpModify:
		return c.compileFunc(
			&Func{
				Name: op.getFunc(),
				Args: []*Query{l, r},
			},
		)
	default:
		name := "$%0"
		c.append(&code{op: opdup})
		if err := c.compileQuery(r); err != nil {
			return err
		}
		c.append(&code{op: opstore, v: c.pushVariable(name)})
		return c.compileFunc(
			&Func{
				Name: "_modify",
				Args: []*Query{
					l,
					{Term: &Term{
						Type: TermTypeFunc,
						Func: &Func{
							Name: op.getFunc(),
							Args: []*Query{
								{Term: &Term{Type: TermTypeIdentity}},
								{Func: name},
							},
						},
					}},
				},
			},
		)
	}
}

func (c *compiler) compileBind(b *Bind) error {
	var pc int
	var vs [][2]int
	for i, p := range b.Patterns {
		var pcc int
		var err error
		if i < len(b.Patterns)-1 {
			defer c.lazy(func() *code {
				return &code{op: opforkalt, v: pcc}
			})()
		}
		if 0 < i {
			for _, v := range vs {
				c.append(&code{op: oppush, v: nil})
				c.append(&code{op: opstore, v: v})
			}
		}
		vs, err = c.compilePattern(p)
		if err != nil {
			return err
		}
		if i < len(b.Patterns)-1 {
			defer c.lazy(func() *code {
				return &code{op: opjump, v: pc}
			})()
			pcc = c.pc()
		}
	}
	if len(b.Patterns) > 1 {
		pc = c.pc()
	}
	if len(b.Patterns) == 1 && c.codes[len(c.codes)-2].op == opexpbegin {
		c.codes[len(c.codes)-2].op = opnop
	} else {
		c.append(&code{op: opexpend}) // ref: compileTermSuffix
	}
	return c.compileQuery(b.Body)
}

func (c *compiler) compilePattern(p *Pattern) ([][2]int, error) {
	c.appendCodeInfo(p)
	if p.Name != "" {
		v := c.pushVariable(p.Name)
		c.append(&code{op: opstore, v: v})
		return [][2]int{v}, nil
	} else if len(p.Array) > 0 {
		var vs [][2]int
		v := c.newVariable()
		c.append(&code{op: opstore, v: v})
		for i, p := range p.Array {
			c.append(&code{op: oppush, v: i})
			c.append(&code{op: opload, v: v})
			c.append(&code{op: opload, v: v})
			// ref: compileCall
			c.append(&code{op: opcall, v: [3]interface{}{internalFuncs["_index"].callback, 2, "_index"}})
			ns, err := c.compilePattern(p)
			if err != nil {
				return nil, err
			}
			vs = append(vs, ns...)
		}
		return vs, nil
	} else if len(p.Object) > 0 {
		var vs [][2]int
		v := c.newVariable()
		c.append(&code{op: opstore, v: v})
		for _, kv := range p.Object {
			var key, name string
			if kv.KeyOnly != "" {
				key, name = kv.KeyOnly[1:], kv.KeyOnly
				c.append(&code{op: oppush, v: key})
			} else if kv.Key != "" {
				key = kv.Key
				if key != "" && key[0] == '$' {
					key, name = key[1:], key
				}
				c.append(&code{op: oppush, v: key})
			} else if kv.KeyString != nil {
				c.append(&code{op: opload, v: v})
				if err := c.compileString(kv.KeyString, nil); err != nil {
					return nil, err
				}
			} else if kv.KeyQuery != nil {
				c.append(&code{op: opload, v: v})
				if err := c.compileQuery(kv.KeyQuery); err != nil {
					return nil, err
				}
			}
			c.append(&code{op: opload, v: v})
			c.append(&code{op: opload, v: v})
			// ref: compileCall
			c.append(&code{op: opcall, v: [3]interface{}{internalFuncs["_index"].callback, 2, "_index"}})
			if name != "" {
				if kv.Val != nil {
					c.append(&code{op: opdup})
				}
				ns, err := c.compilePattern(&Pattern{Name: name})
				if err != nil {
					return nil, err
				}
				vs = append(vs, ns...)
			}
			if kv.Val != nil {
				ns, err := c.compilePattern(kv.Val)
				if err != nil {
					return nil, err
				}
				vs = append(vs, ns...)
			}
		}
		return vs, nil
	} else {
		return nil, fmt.Errorf("invalid pattern: %s", p)
	}
}

func (c *compiler) compileIf(e *If) error {
	c.appendCodeInfo(e)
	c.append(&code{op: opdup}) // duplicate the value for then or else clause
	c.append(&code{op: opexpbegin})
	pc := len(c.codes)
	f := c.newScopeDepth()
	if err := c.compileQuery(e.Cond); err != nil {
		return err
	}
	f()
	if pc == len(c.codes) {
		c.codes = c.codes[:pc-1]
	} else {
		c.append(&code{op: opexpend})
	}
	pcc := len(c.codes)
	setjumpifnot := c.lazy(func() *code {
		return &code{op: opjumpifnot, v: c.pc() + 1} // if falsy, skip then clause
	})
	f = c.newScopeDepth()
	if err := c.compileQuery(e.Then); err != nil {
		return err
	}
	f()
	setjumpifnot()
	defer c.lazy(func() *code {
		return &code{op: opjump, v: c.pc()} // jump to ret after else clause
	})()
	if len(e.Elif) > 0 {
		return c.compileIf(&If{e.Elif[0].Cond, e.Elif[0].Then, e.Elif[1:], e.Else})
	}
	if e.Else != nil {
		defer c.newScopeDepth()()
		defer func() {
			// optimize constant results
			//    opdup, ..., opjumpifnot, opconst, opjump, opconst
			// => opnop, ..., opjumpifnot, oppush,  opjump, oppush
			if pcc+4 == len(c.codes) &&
				c.codes[pcc+1] != nil && c.codes[pcc+1].op == opconst &&
				c.codes[pcc+3] != nil && c.codes[pcc+3].op == opconst {
				c.codes[pc-2].op = opnop
				c.codes[pcc+1].op = oppush
				c.codes[pcc+3].op = oppush
			}
		}()
		return c.compileQuery(e.Else)
	}
	return nil
}

func (c *compiler) compileTry(e *Try) error {
	c.appendCodeInfo(e)
	setforktrybegin := c.lazy(func() *code {
		return &code{op: opforktrybegin, v: c.pc()}
	})
	f := c.newScopeDepth()
	if err := c.compileQuery(e.Body); err != nil {
		return err
	}
	f()
	c.append(&code{op: opforktryend})
	defer c.lazy(func() *code {
		return &code{op: opjump, v: c.pc()}
	})()
	setforktrybegin()
	if e.Catch != nil {
		defer c.newScopeDepth()()
		return c.compileQuery(e.Catch)
	}
	c.append(&code{op: opbacktrack})
	return nil
}

func (c *compiler) compileReduce(e *Reduce) error {
	c.appendCodeInfo(e)
	defer c.newScopeDepth()()
	defer c.lazy(func() *code {
		return &code{op: opfork, v: c.pc() - 2}
	})()
	c.append(&code{op: opdup})
	v := c.newVariable()
	f := c.newScopeDepth()
	if err := c.compileQuery(e.Start); err != nil {
		return err
	}
	f()
	c.append(&code{op: opstore, v: v})
	if err := c.compileTerm(e.Term); err != nil {
		return err
	}
	if _, err := c.compilePattern(e.Pattern); err != nil {
		return err
	}
	c.append(&code{op: opload, v: v})
	f = c.newScopeDepth()
	if err := c.compileQuery(e.Update); err != nil {
		return err
	}
	f()
	c.append(&code{op: opstore, v: v})
	c.append(&code{op: opbacktrack})
	c.append(&code{op: oppop})
	c.append(&code{op: opload, v: v})
	return nil
}

func (c *compiler) compileForeach(e *Foreach) error {
	c.appendCodeInfo(e)
	defer c.newScopeDepth()()
	c.append(&code{op: opdup})
	v := c.newVariable()
	f := c.newScopeDepth()
	if err := c.compileQuery(e.Start); err != nil {
		return err
	}
	f()
	c.append(&code{op: opstore, v: v})
	if err := c.compileTerm(e.Term); err != nil {
		return err
	}
	if _, err := c.compilePattern(e.Pattern); err != nil {
		return err
	}
	c.append(&code{op: opload, v: v})
	f = c.newScopeDepth()
	if err := c.compileQuery(e.Update); err != nil {
		return err
	}
	f()
	c.append(&code{op: opdup})
	c.append(&code{op: opstore, v: v})
	if e.Extract != nil {
		defer c.newScopeDepth()()
		return c.compileQuery(e.Extract)
	}
	return nil
}

func (c *compiler) compileLabel(e *Label) error {
	c.appendCodeInfo(e)
	v := c.pushVariable("$%" + e.Ident[1:])
	defer c.lazy(func() *code {
		return &code{op: opforklabel, v: v}
	})()
	return c.compileQuery(e.Body)
}

func (c *compiler) compileBreak(label string) error {
	v, err := c.lookupVariable("$%" + label[1:])
	if err != nil {
		return &breakError{label, nil}
	}
	c.append(&code{op: oppop})
	c.append(&code{op: opload, v: v})
	c.append(&code{op: opcall, v: [3]interface{}{
		func(v interface{}, _ []interface{}) interface{} {
			return &breakError{label, v}
		},
		0,
		"_break",
	}})
	return nil
}

func (c *compiler) compileTerm(e *Term) error {
	if len(e.SuffixList) > 0 {
		s := e.SuffixList[len(e.SuffixList)-1]
		t := *e // clone without changing e
		(&t).SuffixList = t.SuffixList[:len(e.SuffixList)-1]
		return c.compileTermSuffix(&t, s)
	}
	switch e.Type {
	case TermTypeIdentity:
		return nil
	case TermTypeRecurse:
		return c.compileFunc(&Func{Name: "recurse"})
	case TermTypeNull:
		c.append(&code{op: opconst, v: nil})
		return nil
	case TermTypeTrue:
		c.append(&code{op: opconst, v: true})
		return nil
	case TermTypeFalse:
		c.append(&code{op: opconst, v: false})
		return nil
	case TermTypeIndex:
		return c.compileIndex(&Term{Type: TermTypeIdentity}, e.Index)
	case TermTypeFunc:
		return c.compileFunc(e.Func)
	case TermTypeObject:
		return c.compileObject(e.Object)
	case TermTypeArray:
		return c.compileArray(e.Array)
	case TermTypeNumber:
		v := normalizeNumber(json.Number(e.Number))
		if err, ok := v.(error); ok {
			return err
		}
		c.append(&code{op: opconst, v: v})
		return nil
	case TermTypeUnary:
		return c.compileUnary(e.Unary)
	case TermTypeFormat:
		return c.compileFormat(e.Format, e.Str)
	case TermTypeString:
		return c.compileString(e.Str, nil)
	case TermTypeIf:
		return c.compileIf(e.If)
	case TermTypeTry:
		return c.compileTry(e.Try)
	case TermTypeReduce:
		return c.compileReduce(e.Reduce)
	case TermTypeForeach:
		return c.compileForeach(e.Foreach)
	case TermTypeLabel:
		return c.compileLabel(e.Label)
	case TermTypeBreak:
		return c.compileBreak(e.Break)
	case TermTypeQuery:
		defer c.newScopeDepth()()
		return c.compileQuery(e.Query)
	default:
		panic("invalid term: " + e.String())
	}
}

func (c *compiler) compileIndex(e *Term, x *Index) error {
	c.appendCodeInfo(x)
	if x.Name != "" {
		return c.compileCall("_index", []*Query{{Term: e}, {Term: &Term{Type: TermTypeString, Str: &String{Str: x.Name}}}})
	}
	if x.Str != nil {
		return c.compileCall("_index", []*Query{{Term: e}, {Term: &Term{Type: TermTypeString, Str: x.Str}}})
	}
	if !x.IsSlice {
		return c.compileCall("_index", []*Query{{Term: e}, x.Start})
	}
	if x.Start == nil {
		return c.compileCall("_slice", []*Query{{Term: e}, x.End, {Term: &Term{Type: TermTypeNull}}})
	}
	if x.End == nil {
		return c.compileCall("_slice", []*Query{{Term: e}, {Term: &Term{Type: TermTypeNull}}, x.Start})
	}
	return c.compileCall("_slice", []*Query{{Term: e}, x.End, x.Start})
}

func (c *compiler) compileFunc(e *Func) error {
	name := e.Name
	if len(e.Args) == 0 {
		if f, v := c.lookupFuncOrVariable(name); f != nil {
			return c.compileCallPc(f, e.Args)
		} else if v != nil {
			if name[0] == '$' {
				c.append(&code{op: oppop})
				c.append(&code{op: opload, v: v.index})
			} else {
				c.append(&code{op: opload, v: v.index})
				c.append(&code{op: opcallpc})
			}
			return nil
		} else if name == "$ENV" || name == "env" {
			env := make(map[string]interface{})
			if c.environLoader != nil {
				for _, kv := range c.environLoader() {
					if i := strings.IndexByte(kv, '='); i > 0 {
						env[kv[:i]] = kv[i+1:]
					}
				}
			}
			c.append(&code{op: opconst, v: env})
			return nil
		} else if name[0] == '$' {
			return &variableNotFoundError{name}
		}
	} else {
		for i := len(c.scopes) - 1; i >= 0; i-- {
			s := c.scopes[i]
			for j := len(s.funcs) - 1; j >= 0; j-- {
				if f := s.funcs[j]; f.name == name && f.argcnt == len(e.Args) {
					return c.compileCallPc(f, e.Args)
				}
			}
		}
	}
	if name[0] == '_' {
		name = name[1:]
	}
	if fds, ok := builtinFuncDefs[name]; ok {
		for _, fd := range fds {
			if len(fd.Args) == len(e.Args) {
				if err := c.compileFuncDef(fd, true); err != nil {
					return err
				}
			}
		}
		s := c.scopes[0]
		for i := len(s.funcs) - 1; i >= 0; i-- {
			if f := s.funcs[i]; f.name == e.Name && f.argcnt == len(e.Args) {
				return c.compileCallPc(f, e.Args)
			}
		}
	}
	if fn, ok := internalFuncs[e.Name]; ok && fn.accept(len(e.Args)) {
		switch e.Name {
		case "empty":
			c.append(&code{op: opbacktrack})
			return nil
		case "path":
			c.append(&code{op: oppathbegin})
			if err := c.compileCall(e.Name, e.Args); err != nil {
				return err
			}
			c.codes[len(c.codes)-1] = &code{op: oppathend}
			return nil
		case "builtins":
			return c.compileCallInternal(
				[3]interface{}{c.funcBuiltins, 0, e.Name},
				e.Args,
				true,
				false,
			)
		case "input":
			if c.inputIter == nil {
				return &inputNotAllowedError{}
			}
			return c.compileCallInternal(
				[3]interface{}{c.funcInput, 0, e.Name},
				e.Args,
				true,
				false,
			)
		case "modulemeta":
			return c.compileCallInternal(
				[3]interface{}{c.funcModulemeta, 0, e.Name},
				e.Args,
				true,
				false,
			)
		default:
			return c.compileCall(e.Name, e.Args)
		}
	}
	if fn, ok := c.customFuncs[e.Name]; ok && fn.accept(len(e.Args)) {
		if err := c.compileCallInternal(
			[3]interface{}{fn.callback, len(e.Args), e.Name},
			e.Args,
			true,
			false,
		); err != nil {
			return err
		}
		if fn.iter {
			c.append(&code{op: opeach})
		}
		return nil
	}
	return &funcNotFoundError{e}
}

func (c *compiler) funcBuiltins(interface{}, []interface{}) interface{} {
	type funcNameArity struct {
		name  string
		arity int
	}
	var xs []*funcNameArity
	for _, fds := range builtinFuncDefs {
		for _, fd := range fds {
			if fd.Name[0] != '_' {
				xs = append(xs, &funcNameArity{fd.Name, len(fd.Args)})
			}
		}
	}
	for name, fn := range internalFuncs {
		if name[0] != '_' {
			for i, cnt := 0, fn.argcount; cnt > 0; i, cnt = i+1, cnt>>1 {
				if cnt&1 > 0 {
					xs = append(xs, &funcNameArity{name, i})
				}
			}
		}
	}
	for name, fn := range c.customFuncs {
		if name[0] != '_' {
			for i, cnt := 0, fn.argcount; cnt > 0; i, cnt = i+1, cnt>>1 {
				if cnt&1 > 0 {
					xs = append(xs, &funcNameArity{name, i})
				}
			}
		}
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i].name < xs[j].name ||
			xs[i].name == xs[j].name && xs[i].arity < xs[j].arity
	})
	ys := make([]interface{}, len(xs))
	for i, x := range xs {
		ys[i] = x.name + "/" + strconv.Itoa(x.arity)
	}
	return ys
}

func (c *compiler) funcInput(interface{}, []interface{}) interface{} {
	v, ok := c.inputIter.Next()
	if !ok {
		return errors.New("break")
	}
	return normalizeNumbers(v)
}

func (c *compiler) funcModulemeta(v interface{}, _ []interface{}) interface{} {
	s, ok := v.(string)
	if !ok {
		return &funcTypeError{"modulemeta", v}
	}
	if c.moduleLoader == nil {
		return fmt.Errorf("cannot load module: %q", s)
	}
	var q *Query
	var err error
	if moduleLoader, ok := c.moduleLoader.(interface {
		LoadModuleWithMeta(string, map[string]interface{}) (*Query, error)
	}); ok {
		if q, err = moduleLoader.LoadModuleWithMeta(s, nil); err != nil {
			return err
		}
	} else if moduleLoader, ok := c.moduleLoader.(interface {
		LoadModule(string) (*Query, error)
	}); ok {
		if q, err = moduleLoader.LoadModule(s); err != nil {
			return err
		}
	}
	meta := q.Meta.ToValue()
	if meta == nil {
		meta = make(map[string]interface{})
	}
	var deps []interface{}
	for _, i := range q.Imports {
		v := i.Meta.ToValue()
		if v == nil {
			v = make(map[string]interface{})
		} else {
			for k := range v {
				// dirty hack to remove the internal fields
				if strings.HasPrefix(k, "$$") {
					delete(v, k)
				}
			}
		}
		if i.ImportPath == "" {
			v["relpath"] = i.IncludePath
		} else {
			v["relpath"] = i.ImportPath
		}
		if err != nil {
			return err
		}
		if i.ImportAlias != "" {
			v["as"] = strings.TrimPrefix(i.ImportAlias, "$")
		}
		v["is_data"] = strings.HasPrefix(i.ImportAlias, "$")
		deps = append(deps, v)
	}
	meta["deps"] = deps
	return meta
}

func (c *compiler) compileObject(e *Object) error {
	c.appendCodeInfo(e)
	if len(e.KeyVals) == 0 {
		c.append(&code{op: opconst, v: map[string]interface{}{}})
		return nil
	}
	defer c.newScopeDepth()()
	v := c.newVariable()
	c.append(&code{op: opstore, v: v})
	pc := len(c.codes)
	for _, kv := range e.KeyVals {
		if err := c.compileObjectKeyVal(v, kv); err != nil {
			return err
		}
	}
	c.append(&code{op: opobject, v: len(e.KeyVals)})
	// optimize constant objects
	l := len(e.KeyVals)
	if pc+l*3+1 != len(c.codes) {
		return nil
	}
	for i := 0; i < l; i++ {
		if c.codes[pc+i*3].op != oppush ||
			c.codes[pc+i*3+1].op != opload ||
			c.codes[pc+i*3+2].op != opconst {
			return nil
		}
	}
	w := make(map[string]interface{}, l)
	for i := 0; i < l; i++ {
		w[c.codes[pc+i*3].v.(string)] = c.codes[pc+i*3+2].v
	}
	c.codes[pc-1] = &code{op: opconst, v: w}
	c.codes = c.codes[:pc]
	return nil
}

func (c *compiler) compileObjectKeyVal(v [2]int, kv *ObjectKeyVal) error {
	if kv.KeyOnly != "" {
		if kv.KeyOnly[0] == '$' {
			c.append(&code{op: oppush, v: kv.KeyOnly[1:]})
			c.append(&code{op: opload, v: v})
			return c.compileFunc(&Func{Name: kv.KeyOnly})
		}
		c.append(&code{op: oppush, v: kv.KeyOnly})
		c.append(&code{op: opload, v: v})
		return c.compileIndex(&Term{Type: TermTypeIdentity}, &Index{Name: kv.KeyOnly})
	} else if kv.KeyOnlyString != nil {
		c.append(&code{op: opload, v: v})
		if err := c.compileString(kv.KeyOnlyString, nil); err != nil {
			return err
		}
		c.append(&code{op: opdup})
		c.append(&code{op: opload, v: v})
		c.append(&code{op: opload, v: v})
		// ref: compileCall
		c.append(&code{op: opcall, v: [3]interface{}{internalFuncs["_index"].callback, 2, "_index"}})
		return nil
	} else {
		if kv.KeyQuery != nil {
			c.append(&code{op: opload, v: v})
			f := c.newScopeDepth()
			if err := c.compileQuery(kv.KeyQuery); err != nil {
				return err
			}
			f()
		} else if kv.KeyString != nil {
			c.append(&code{op: opload, v: v})
			if err := c.compileString(kv.KeyString, nil); err != nil {
				return err
			}
			if d := c.codes[len(c.codes)-1]; d.op == opconst {
				c.codes[len(c.codes)-2] = &code{op: oppush, v: d.v}
				c.codes = c.codes[:len(c.codes)-1]
			}
		} else if kv.Key[0] == '$' {
			c.append(&code{op: opload, v: v})
			if err := c.compileFunc(&Func{Name: kv.Key}); err != nil {
				return err
			}
		} else {
			c.append(&code{op: oppush, v: kv.Key})
		}
		c.append(&code{op: opload, v: v})
		return c.compileObjectVal(kv.Val)
	}
}

func (c *compiler) compileObjectVal(e *ObjectVal) error {
	for _, e := range e.Queries {
		if err := c.compileQuery(e); err != nil {
			return err
		}
	}
	return nil
}

func (c *compiler) compileArray(e *Array) error {
	c.appendCodeInfo(e)
	if e.Query == nil {
		c.append(&code{op: opconst, v: []interface{}{}})
		return nil
	}
	c.append(&code{op: oppush, v: []interface{}{}})
	arr := c.newVariable()
	c.append(&code{op: opstore, v: arr})
	pc := len(c.codes)
	c.append(&code{op: opfork})
	defer func() {
		if pc < len(c.codes) {
			c.codes[pc].v = c.pc() - 2
		}
	}()
	defer c.newScopeDepth()()
	if err := c.compileQuery(e.Query); err != nil {
		return err
	}
	c.append(&code{op: opappend, v: arr})
	c.append(&code{op: opbacktrack})
	c.append(&code{op: oppop})
	c.append(&code{op: opload, v: arr})
	if e.Query.Op == OpPipe {
		return nil
	}
	// optimize constant arrays
	if (len(c.codes)-pc)%3 != 0 {
		return nil
	}
	l := (len(c.codes) - pc - 3) / 3
	for i := 0; i < l; i++ {
		if c.codes[pc+i].op != opfork ||
			c.codes[pc+i*2+l].op != opconst ||
			(i < l-1 && c.codes[pc+i*2+l+1].op != opjump) {
			return nil
		}
	}
	v := make([]interface{}, l)
	for i := 0; i < l; i++ {
		v[i] = c.codes[pc+i*2+l].v
	}
	c.codes[pc-2] = &code{op: opconst, v: v}
	c.codes = c.codes[:pc-1]
	return nil
}

func (c *compiler) compileUnary(e *Unary) error {
	c.appendCodeInfo(e)
	if err := c.compileTerm(e.Term); err != nil {
		return err
	}
	switch e.Op {
	case OpAdd:
		return c.compileCall("_plus", nil)
	case OpSub:
		return c.compileCall("_negate", nil)
	default:
		return fmt.Errorf("unexpected operator in Unary: %s", e.Op)
	}
}

func (c *compiler) compileFormat(fmt string, str *String) error {
	f := formatToFunc(fmt)
	if f == nil {
		f = &Func{
			Name: "format",
			Args: []*Query{{Term: &Term{Type: TermTypeString, Str: &String{Str: fmt[1:]}}}},
		}
	}
	if str == nil {
		return c.compileFunc(f)
	}
	return c.compileString(str, f)
}

func formatToFunc(fmt string) *Func {
	switch fmt {
	case "@text":
		return &Func{Name: "tostring"}
	case "@json":
		return &Func{Name: "tojson"}
	case "@html":
		return &Func{Name: "_tohtml"}
	case "@uri":
		return &Func{Name: "_touri"}
	case "@csv":
		return &Func{Name: "_tocsv"}
	case "@tsv":
		return &Func{Name: "_totsv"}
	case "@sh":
		return &Func{Name: "_tosh"}
	case "@base64":
		return &Func{Name: "_tobase64"}
	case "@base64d":
		return &Func{Name: "_tobase64d"}
	default:
		return nil
	}
}

func (c *compiler) compileString(s *String, f *Func) error {
	if s.Queries == nil {
		c.append(&code{op: opconst, v: s.Str})
		return nil
	}
	if f == nil {
		f = &Func{Name: "tostring"}
	}
	var q *Query
	for _, e := range s.Queries {
		if e.Term.Str == nil {
			e = &Query{Left: e, Op: OpPipe, Right: &Query{Term: &Term{Type: TermTypeFunc, Func: f}}}
		}
		if q == nil {
			q = e
		} else {
			q = &Query{Left: q, Op: OpAdd, Right: e}
		}
	}
	return c.compileQuery(q)
}

func (c *compiler) compileTermSuffix(e *Term, s *Suffix) error {
	if s.Index != nil {
		return c.compileIndex(e, s.Index)
	} else if s.Iter {
		if err := c.compileTerm(e); err != nil {
			return err
		}
		c.append(&code{op: opeach})
		return nil
	} else if s.Optional {
		if len(e.SuffixList) > 0 {
			if u, ok := e.SuffixList[len(e.SuffixList)-1].toTerm(); ok {
				t := *e // clone without changing e
				(&t).SuffixList = t.SuffixList[:len(e.SuffixList)-1]
				if err := c.compileTerm(&t); err != nil {
					return err
				}
				e = u
			}
		}
		return c.compileTry(&Try{Body: &Query{Term: e}})
	} else if s.Bind != nil {
		c.append(&code{op: opdup})
		c.append(&code{op: opexpbegin})
		if err := c.compileTerm(e); err != nil {
			return err
		}
		return c.compileBind(s.Bind)
	} else {
		return fmt.Errorf("invalid suffix: %s", s)
	}
}

func (c *compiler) compileCall(name string, args []*Query) error {
	fn := internalFuncs[name]
	if err := c.compileCallInternal(
		[3]interface{}{fn.callback, len(args), name},
		args,
		true,
		name == "_index" || name == "_slice",
	); err != nil {
		return err
	}
	if fn.iter {
		c.append(&code{op: opeach})
	}
	return nil
}

func (c *compiler) compileCallPc(fn *funcinfo, args []*Query) error {
	return c.compileCallInternal(fn.pc, args, false, false)
}

func (c *compiler) compileCallInternal(
	fn interface{}, args []*Query, internal, indexing bool) error {
	if len(args) == 0 {
		c.append(&code{op: opcall, v: fn})
		return nil
	}
	idx := c.newVariable()
	c.append(&code{op: opstore, v: idx})
	if indexing && len(args) > 1 {
		c.append(&code{op: opexpbegin})
	}
	for i := len(args) - 1; i >= 0; i-- {
		pc := c.pc() + 1 // skip opjump (ref: compileFuncDef)
		name := "lambda:" + strconv.Itoa(pc)
		if err := c.compileFuncDef(&FuncDef{Name: name, Body: args[i]}, false); err != nil {
			return err
		}
		if internal {
			switch c.pc() - pc {
			case 2: // optimize identity argument (opscope, opret)
				j := len(c.codes) - 3
				c.codes[j] = &code{op: opload, v: idx}
				c.codes = c.codes[:j+1]
				s := c.scopes[len(c.scopes)-1]
				s.funcs = s.funcs[:len(s.funcs)-1]
				c.deleteCodeInfo(name)
			case 3: // optimize one instruction argument (opscope, opX, opret)
				j := len(c.codes) - 4
				if c.codes[j+2].op == opconst {
					c.codes[j] = &code{op: oppush, v: c.codes[j+2].v}
					c.codes = c.codes[:j+1]
				} else {
					c.codes[j] = &code{op: opload, v: idx}
					c.codes[j+1] = c.codes[j+2]
					c.codes = c.codes[:j+2]
				}
				s := c.scopes[len(c.scopes)-1]
				s.funcs = s.funcs[:len(s.funcs)-1]
				c.deleteCodeInfo(name)
			default:
				c.append(&code{op: opload, v: idx})
				c.append(&code{op: oppushpc, v: pc})
				c.append(&code{op: opcallpc})
			}
		} else {
			c.append(&code{op: oppushpc, v: pc})
		}
		if indexing && i == 1 {
			if c.codes[len(c.codes)-2].op == opexpbegin {
				c.codes[len(c.codes)-2] = c.codes[len(c.codes)-1]
				c.codes = c.codes[:len(c.codes)-1]
			} else {
				c.append(&code{op: opexpend})
			}
		}
	}
	c.append(&code{op: opload, v: idx})
	c.append(&code{op: opcall, v: fn})
	return nil
}

func (c *compiler) append(code *code) {
	c.codes = append(c.codes, code)
}

func (c *compiler) pc() int {
	return len(c.codes)
}

func (c *compiler) lazy(f func() *code) func() {
	i := len(c.codes)
	c.codes = append(c.codes, nil)
	return func() { c.codes[i] = f() }
}

func (c *compiler) optimizeTailRec() {
	var pcs []int
	scopes := map[int]bool{}
L:
	for i, l := 0, len(c.codes); i < l; i++ {
		switch c.codes[i].op {
		case opscope:
			pcs = append(pcs, i)
			if v := c.codes[i].v.([3]int); v[2] == 0 {
				scopes[i] = v[1] == 0
			}
		case opcall:
			var canjump bool
			if j, ok := c.codes[i].v.(int); !ok ||
				len(pcs) == 0 || pcs[len(pcs)-1] != j {
				break
			} else if canjump, ok = scopes[j]; !ok {
				break
			}
			for j := i + 1; j < l; {
				switch c.codes[j].op {
				case opjump:
					j = c.codes[j].v.(int)
				case opret:
					if canjump {
						c.codes[i].op = opjump
						c.codes[i].v = pcs[len(pcs)-1] + 1
					} else {
						c.codes[i].op = opcallrec
					}
					continue L
				default:
					continue L
				}
			}
		case opret:
			if len(pcs) == 0 {
				break L
			}
			pcs = pcs[:len(pcs)-1]
		}
	}
}

func (c *compiler) optimizeCodeOps() {
	for i, next := len(c.codes)-1, (*code)(nil); i >= 0; i-- {
		code := c.codes[i]
		switch code.op {
		case oppush, opdup, opload:
			switch next.op {
			case oppop:
				code.op = opnop
				next.op = opnop
			case opconst:
				code.op = opnop
				next.op = oppush
			}
		case opjump, opjumpifnot:
			if j := code.v.(int); j-1 == i {
				code.op = opnop
			} else if next = c.codes[j]; next.op == opjump {
				code.v = next.v
			}
		}
		next = code
	}
}
