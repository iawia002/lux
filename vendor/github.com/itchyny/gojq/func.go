package gojq

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/itchyny/timefmt-go"
)

//go:generate go run -modfile=go.dev.mod _tools/gen_builtin.go -i builtin.jq -o builtin.go
var builtinFuncDefs map[string][]*FuncDef

const (
	argcount0 = 1 << iota
	argcount1
	argcount2
	argcount3
)

type function struct {
	argcount int
	iter     bool
	callback func(interface{}, []interface{}) interface{}
}

func (fn function) accept(cnt int) bool {
	return fn.argcount&(1<<cnt) != 0
}

var internalFuncs map[string]function

func init() {
	internalFuncs = map[string]function{
		"empty":          argFunc0(nil),
		"path":           argFunc1(nil),
		"env":            argFunc0(nil),
		"builtins":       argFunc0(nil),
		"input":          argFunc0(nil),
		"modulemeta":     argFunc0(nil),
		"length":         argFunc0(funcLength),
		"utf8bytelength": argFunc0(funcUtf8ByteLength),
		"keys":           argFunc0(funcKeys),
		"has":            argFunc1(funcHas),
		"add":            argFunc0(funcAdd),
		"tonumber":       argFunc0(funcToNumber),
		"tostring":       argFunc0(funcToString),
		"type":           argFunc0(funcType),
		"reverse":        argFunc0(funcReverse),
		"contains":       argFunc1(funcContains),
		"explode":        argFunc0(funcExplode),
		"implode":        argFunc0(funcImplode),
		"split":          {argcount1 | argcount2, false, funcSplit},
		"tojson":         argFunc0(funcToJSON),
		"fromjson":       argFunc0(funcFromJSON),
		"format":         argFunc1(funcFormat),
		"_tohtml":        argFunc0(funcToHTML),
		"_touri":         argFunc0(funcToURI),
		"_tocsv":         argFunc0(funcToCSV),
		"_totsv":         argFunc0(funcToTSV),
		"_tosh":          argFunc0(funcToSh),
		"_tobase64":      argFunc0(funcToBase64),
		"_tobase64d":     argFunc0(funcToBase64d),
		"_index":         argFunc2(funcIndex),
		"_slice":         argFunc3(funcSlice),
		"_indices":       argFunc1(funcIndices),
		"_lindex":        argFunc1(funcLindex),
		"_rindex":        argFunc1(funcRindex),
		"_plus":          argFunc0(funcOpPlus),
		"_negate":        argFunc0(funcOpNegate),
		"_add":           argFunc2(funcOpAdd),
		"_subtract":      argFunc2(funcOpSub),
		"_multiply":      argFunc2(funcOpMul),
		"_divide":        argFunc2(funcOpDiv),
		"_modulo":        argFunc2(funcOpMod),
		"_alternative":   argFunc2(funcOpAlt),
		"_equal":         argFunc2(funcOpEq),
		"_notequal":      argFunc2(funcOpNe),
		"_greater":       argFunc2(funcOpGt),
		"_less":          argFunc2(funcOpLt),
		"_greatereq":     argFunc2(funcOpGe),
		"_lesseq":        argFunc2(funcOpLe),
		"_range":         {argcount3, true, funcRange},
		"_min_by":        argFunc1(funcMinBy),
		"_max_by":        argFunc1(funcMaxBy),
		"_sort_by":       argFunc1(funcSortBy),
		"_group_by":      argFunc1(funcGroupBy),
		"_unique_by":     argFunc1(funcUniqueBy),
		"_join":          argFunc1(funcJoin),
		"sin":            mathFunc("sin", math.Sin),
		"cos":            mathFunc("cos", math.Cos),
		"tan":            mathFunc("tan", math.Tan),
		"asin":           mathFunc("asin", math.Asin),
		"acos":           mathFunc("acos", math.Acos),
		"atan":           mathFunc("atan", math.Atan),
		"sinh":           mathFunc("sinh", math.Sinh),
		"cosh":           mathFunc("cosh", math.Cosh),
		"tanh":           mathFunc("tanh", math.Tanh),
		"asinh":          mathFunc("asinh", math.Asinh),
		"acosh":          mathFunc("acosh", math.Acosh),
		"atanh":          mathFunc("atanh", math.Atanh),
		"floor":          mathFunc("floor", math.Floor),
		"round":          mathFunc("round", math.Round),
		"nearbyint":      mathFunc("nearbyint", math.Round),
		"rint":           mathFunc("rint", math.Round),
		"ceil":           mathFunc("ceil", math.Ceil),
		"trunc":          mathFunc("trunc", math.Trunc),
		"significand":    mathFunc("significand", funcSignificand),
		"fabs":           mathFunc("fabs", math.Abs),
		"sqrt":           mathFunc("sqrt", math.Sqrt),
		"cbrt":           mathFunc("cbrt", math.Cbrt),
		"exp":            mathFunc("exp", math.Exp),
		"exp10":          mathFunc("exp10", funcExp10),
		"exp2":           mathFunc("exp2", math.Exp2),
		"expm1":          mathFunc("expm1", math.Expm1),
		"frexp":          argFunc0(funcFrexp),
		"modf":           argFunc0(funcModf),
		"log":            mathFunc("log", math.Log),
		"log10":          mathFunc("log10", math.Log10),
		"log1p":          mathFunc("log1p", math.Log1p),
		"log2":           mathFunc("log2", math.Log2),
		"logb":           mathFunc("logb", math.Logb),
		"gamma":          mathFunc("gamma", math.Gamma),
		"tgamma":         mathFunc("tgamma", math.Gamma),
		"lgamma":         mathFunc("lgamma", funcLgamma),
		"erf":            mathFunc("erf", math.Erf),
		"erfc":           mathFunc("erfc", math.Erfc),
		"j0":             mathFunc("j0", math.J0),
		"j1":             mathFunc("j1", math.J1),
		"y0":             mathFunc("y0", math.Y0),
		"y1":             mathFunc("y1", math.Y1),
		"atan2":          mathFunc2("atan2", math.Atan2),
		"copysign":       mathFunc2("copysign", math.Copysign),
		"drem":           mathFunc2("drem", funcDrem),
		"fdim":           mathFunc2("fdim", math.Dim),
		"fmax":           mathFunc2("fmax", math.Max),
		"fmin":           mathFunc2("fmin", math.Min),
		"fmod":           mathFunc2("fmod", math.Mod),
		"hypot":          mathFunc2("hypot", math.Hypot),
		"jn":             mathFunc2("jn", funcJn),
		"ldexp":          mathFunc2("ldexp", funcLdexp),
		"nextafter":      mathFunc2("nextafter", math.Nextafter),
		"nexttoward":     mathFunc2("nexttoward", math.Nextafter),
		"remainder":      mathFunc2("remainder", math.Remainder),
		"scalb":          mathFunc2("scalb", funcScalb),
		"scalbln":        mathFunc2("scalbln", funcScalbln),
		"yn":             mathFunc2("yn", funcYn),
		"pow":            mathFunc2("pow", math.Pow),
		"pow10":          mathFunc("pow10", funcExp10),
		"fma":            mathFunc3("fma", math.FMA),
		"infinite":       argFunc0(funcInfinite),
		"isfinite":       argFunc0(funcIsfinite),
		"isinfinite":     argFunc0(funcIsinfinite),
		"nan":            argFunc0(funcNan),
		"isnan":          argFunc0(funcIsnan),
		"isnormal":       argFunc0(funcIsnormal),
		"setpath":        argFunc2(funcSetpath),
		"delpaths":       argFunc1(funcDelpaths),
		"getpath":        argFunc1(funcGetpath),
		"transpose":      argFunc0(funcTranspose),
		"bsearch":        argFunc1(funcBsearch),
		"gmtime":         argFunc0(funcGmtime),
		"localtime":      argFunc0(funcLocaltime),
		"mktime":         argFunc0(funcMktime),
		"strftime":       argFunc1(funcStrftime),
		"strflocaltime":  argFunc1(funcStrflocaltime),
		"strptime":       argFunc1(funcStrptime),
		"now":            argFunc0(funcNow),
		"_match":         argFunc3(funcMatch),
		"error":          {argcount0 | argcount1, false, funcError},
		"halt":           argFunc0(funcHalt),
		"halt_error":     {argcount0 | argcount1, false, funcHaltError},
		"_type_error":    argFunc1(internalfuncTypeError),
	}
}

func argFunc0(fn func(interface{}) interface{}) function {
	return function{
		argcount0, false, func(v interface{}, _ []interface{}) interface{} {
			return fn(v)
		},
	}
}

func argFunc1(fn func(_, _ interface{}) interface{}) function {
	return function{
		argcount1, false, func(v interface{}, args []interface{}) interface{} {
			return fn(v, args[0])
		},
	}
}

func argFunc2(fn func(_, _, _ interface{}) interface{}) function {
	return function{
		argcount2, false, func(v interface{}, args []interface{}) interface{} {
			return fn(v, args[0], args[1])
		},
	}
}

func argFunc3(fn func(_, _, _, _ interface{}) interface{}) function {
	return function{
		argcount3, false, func(v interface{}, args []interface{}) interface{} {
			return fn(v, args[0], args[1], args[2])
		},
	}
}

func mathFunc(name string, f func(float64) float64) function {
	return argFunc0(func(v interface{}) interface{} {
		x, ok := toFloat(v)
		if !ok {
			return &funcTypeError{name, v}
		}
		return f(x)
	})
}

func mathFunc2(name string, f func(_, _ float64) float64) function {
	return argFunc2(func(_, x, y interface{}) interface{} {
		l, ok := toFloat(x)
		if !ok {
			return &funcTypeError{name, x}
		}
		r, ok := toFloat(y)
		if !ok {
			return &funcTypeError{name, y}
		}
		return f(l, r)
	})
}

func mathFunc3(name string, f func(_, _, _ float64) float64) function {
	return argFunc3(func(_, a, b, c interface{}) interface{} {
		x, ok := toFloat(a)
		if !ok {
			return &funcTypeError{name, a}
		}
		y, ok := toFloat(b)
		if !ok {
			return &funcTypeError{name, b}
		}
		z, ok := toFloat(c)
		if !ok {
			return &funcTypeError{name, c}
		}
		return f(x, y, z)
	})
}

func funcLength(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return len(v)
	case map[string]interface{}:
		return len(v)
	case string:
		return len([]rune(v))
	case int:
		if v >= 0 {
			return v
		}
		return -v
	case float64:
		return math.Abs(v)
	case *big.Int:
		if v.Sign() >= 0 {
			return v
		}
		return new(big.Int).Abs(v)
	case nil:
		return 0
	default:
		return &funcTypeError{"length", v}
	}
}

func funcUtf8ByteLength(v interface{}) interface{} {
	switch v := v.(type) {
	case string:
		return len(v)
	default:
		return &funcTypeError{"utf8bytelength", v}
	}
}

func funcKeys(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		w := make([]interface{}, len(v))
		for i := range v {
			w[i] = i
		}
		return w
	case map[string]interface{}:
		w := make([]string, len(v))
		var i int
		for k := range v {
			w[i] = k
			i++
		}
		sort.Strings(w)
		u := make([]interface{}, len(v))
		for i, x := range w {
			u[i] = x
		}
		return u
	default:
		return &funcTypeError{"keys", v}
	}
}

func funcHas(v, x interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		if x, ok := toInt(x); ok {
			return 0 <= x && x < len(v)
		}
		return &hasKeyTypeError{v, x}
	case map[string]interface{}:
		switch x := x.(type) {
		case string:
			_, ok := v[x]
			return ok
		default:
			return &hasKeyTypeError{v, x}
		}
	case nil:
		return false
	default:
		return &hasKeyTypeError{v, x}
	}
}

func funcAdd(v interface{}) interface{} {
	if vs, ok := v.(map[string]interface{}); ok {
		xs := make([]string, len(vs))
		var i int
		for k := range vs {
			xs[i] = k
			i++
		}
		sort.Strings(xs)
		us := make([]interface{}, len(vs))
		for i, x := range xs {
			us[i] = vs[x]
		}
		v = us
	}
	vs, ok := v.([]interface{})
	if !ok {
		return &funcTypeError{"add", v}
	}
	v = nil
	for _, x := range vs {
		switch y := x.(type) {
		case map[string]interface{}:
			switch w := v.(type) {
			case nil:
				m := make(map[string]interface{}, len(y))
				for k, e := range y {
					m[k] = e
				}
				v = m
				continue
			case map[string]interface{}:
				for k, e := range y {
					w[k] = e
				}
				continue
			}
		case []interface{}:
			switch w := v.(type) {
			case nil:
				s := make([]interface{}, len(y))
				copy(s, y)
				v = s
				continue
			case []interface{}:
				v = append(w, y...)
				continue
			}
		}
		v = funcOpAdd(nil, v, x)
		if err, ok := v.(error); ok {
			return err
		}
	}
	return v
}

func funcToNumber(v interface{}) interface{} {
	switch v := v.(type) {
	case int, float64, *big.Int:
		return v
	case string:
		if !newLexer(v).validNumber() {
			return fmt.Errorf("invalid number: %q", v)
		}
		return normalizeNumber(json.Number(v))
	default:
		return &funcTypeError{"tonumber", v}
	}
}

func funcToString(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		return s
	}
	return funcToJSON(v)
}

func funcType(v interface{}) interface{} {
	return typeof(v)
}

func funcReverse(v interface{}) interface{} {
	vs, ok := v.([]interface{})
	if !ok {
		return &expectedArrayError{v}
	}
	ws := make([]interface{}, len(vs))
	for i, v := range vs {
		ws[len(ws)-i-1] = v
	}
	return ws
}

func funcContains(v, x interface{}) interface{} {
	switch v := v.(type) {
	case nil:
		if x == nil {
			return true
		}
	case bool:
		switch x := x.(type) {
		case bool:
			if v == x {
				return true
			}
		}
	}
	return binopTypeSwitch(v, x,
		func(l, r int) interface{} { return l == r },
		func(l, r float64) interface{} { return l == r },
		func(l, r *big.Int) interface{} { return l.Cmp(r) == 0 },
		func(l, r string) interface{} { return strings.Contains(l, r) },
		func(l, r []interface{}) interface{} {
			for _, x := range r {
				var found bool
				for _, y := range l {
					if funcContains(y, x) == true {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
			return true
		},
		func(l, r map[string]interface{}) interface{} {
			for k, rk := range r {
				lk, ok := l[k]
				if !ok {
					return false
				}
				c := funcContains(lk, rk)
				if _, ok := c.(error); ok {
					return false
				}
				if c == false {
					return false
				}
			}
			return true
		},
		func(l, r interface{}) interface{} { return &funcContainsError{l, r} },
	)
}

func funcExplode(v interface{}) interface{} {
	switch v := v.(type) {
	case string:
		return explode(v)
	default:
		return &funcTypeError{"explode", v}
	}
}

func explode(s string) []interface{} {
	rs := []int32(s)
	xs := make([]interface{}, len(rs))
	for i, r := range rs {
		xs[i] = int(r)
	}
	return xs
}

func funcImplode(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return implode(v)
	default:
		return &funcTypeError{"implode", v}
	}
}

func implode(v []interface{}) interface{} {
	var sb strings.Builder
	sb.Grow(len(v))
	for _, r := range v {
		if r, ok := toInt(r); ok && 0 <= r && r <= utf8.MaxRune {
			sb.WriteRune(rune(r))
		} else {
			return &funcTypeError{"implode", v}
		}
	}
	return sb.String()
}

func funcSplit(v interface{}, args []interface{}) interface{} {
	s, ok := v.(string)
	if !ok {
		return &funcTypeError{"split", v}
	}
	x, ok := args[0].(string)
	if !ok {
		return &funcTypeError{"split", x}
	}
	var ss []string
	if len(args) == 1 {
		ss = strings.Split(s, x)
	} else {
		var flags string
		if args[1] != nil {
			v, ok := args[1].(string)
			if !ok {
				return &funcTypeError{"split", args[1]}
			}
			flags = v
		}
		r, err := compileRegexp(x, flags)
		if err != nil {
			return err
		}
		ss = r.Split(s, -1)
	}
	xs := make([]interface{}, len(ss))
	for i, s := range ss {
		xs[i] = s
	}
	return xs
}

func funcToJSON(v interface{}) interface{} {
	return jsonMarshal(v)
}

func funcFromJSON(v interface{}) interface{} {
	switch v := v.(type) {
	case string:
		var w interface{}
		err := json.Unmarshal([]byte(v), &w)
		if err != nil {
			return err
		}
		return w
	default:
		return &funcTypeError{"fromjson", v}
	}
}

func funcFormat(v, x interface{}) interface{} {
	switch x := x.(type) {
	case string:
		fmt := "@" + x
		f := formatToFunc(fmt)
		if f == nil {
			return &formatNotFoundError{fmt}
		}
		return internalFuncs[f.Name].callback(v, nil)
	default:
		return &funcTypeError{"format", x}
	}
}

var htmlEscaper = strings.NewReplacer(
	`<`, "&lt;",
	`>`, "&gt;",
	`&`, "&amp;",
	`'`, "&apos;",
	`"`, "&quot;",
)

func funcToHTML(v interface{}) interface{} {
	switch x := funcToString(v).(type) {
	case string:
		return htmlEscaper.Replace(x)
	default:
		return x
	}
}

func funcToURI(v interface{}) interface{} {
	switch x := funcToString(v).(type) {
	case string:
		return url.QueryEscape(x)
	default:
		return x
	}
}

func funcToCSV(v interface{}) interface{} {
	return funcToCSVTSV("csv", v, ",", func(s string) string {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	})
}

var tsvEscaper = strings.NewReplacer(
	"\t", `\t`,
	"\r", `\r`,
	"\n", `\n`,
	"\\", `\\`,
)

func funcToTSV(v interface{}) interface{} {
	return funcToCSVTSV("tsv", v, "\t", func(s string) string {
		return tsvEscaper.Replace(s)
	})
}

func funcToCSVTSV(typ string, v interface{}, sep string, escape func(string) string) interface{} {
	switch xs := v.(type) {
	case []interface{}:
		ys := make([]string, len(xs))
		for i, x := range xs {
			y, err := toCSVTSV(typ, x, escape)
			if err != nil {
				return err
			}
			ys[i] = y
		}
		return strings.Join(ys, sep)
	default:
		return &expectedArrayError{v}
	}
}

func toCSVTSV(typ string, v interface{}, escape func(string) string) (string, error) {
	switch v := v.(type) {
	case map[string]interface{}, []interface{}:
		return "", &formatCsvTsvRowError{typ, v}
	case string:
		return escape(v), nil
	default:
		if s := jsonMarshal(v); s != "null" {
			return s, nil
		}
		return "", nil
	}
}

func funcToSh(v interface{}) interface{} {
	var xs []interface{}
	if w, ok := v.([]interface{}); ok {
		xs = w
	} else {
		xs = []interface{}{v}
	}
	var s strings.Builder
	for i, x := range xs {
		if i > 0 {
			s.WriteByte(' ')
		}
		switch x := x.(type) {
		case map[string]interface{}, []interface{}:
			return &formatShError{x}
		case string:
			s.WriteByte('\'')
			s.WriteString(strings.ReplaceAll(x, "'", `'\''`))
			s.WriteByte('\'')
		default:
			s.WriteString(jsonMarshal(x))
		}
	}
	return s.String()
}

func funcToBase64(v interface{}) interface{} {
	switch x := funcToString(v).(type) {
	case string:
		return base64.StdEncoding.EncodeToString([]byte(x))
	default:
		return x
	}
}

func funcToBase64d(v interface{}) interface{} {
	switch x := funcToString(v).(type) {
	case string:
		if i := strings.IndexRune(x, base64.StdPadding); i >= 0 {
			x = x[:i]
		}
		y, err := base64.RawStdEncoding.DecodeString(x)
		if err != nil {
			return err
		}
		return string(y)
	default:
		return x
	}
}

func funcIndex(_, v, x interface{}) interface{} {
	switch x := x.(type) {
	case string:
		switch v := v.(type) {
		case nil:
			return nil
		case map[string]interface{}:
			return v[x]
		default:
			return &expectedObjectError{v}
		}
	case int, float64, *big.Int:
		idx, _ := toInt(x)
		switch v := v.(type) {
		case nil:
			return nil
		case []interface{}:
			return funcIndexSlice(nil, nil, &idx, v)
		case string:
			switch v := funcIndexSlice(nil, nil, &idx, explode(v)).(type) {
			case []interface{}:
				return implode(v)
			case int:
				return implode([]interface{}{v})
			case nil:
				return ""
			default:
				panic(v)
			}
		default:
			return &expectedArrayError{v}
		}
	case []interface{}:
		switch v := v.(type) {
		case nil:
			return nil
		case []interface{}:
			return indices(v, x)
		default:
			return &expectedArrayError{v}
		}
	case map[string]interface{}:
		if v == nil {
			return nil
		}
		start, ok := x["start"]
		if !ok {
			return &expectedStartEndError{x}
		}
		end, ok := x["end"]
		if !ok {
			return &expectedStartEndError{x}
		}
		return funcSlice(nil, v, end, start)
	default:
		return &objectKeyNotStringError{x}
	}
}

func indices(vs, xs []interface{}) interface{} {
	var rs []interface{}
	if len(xs) == 0 {
		return rs
	}
	for i := 0; i < len(vs) && i < len(vs)-len(xs)+1; i++ {
		var neq bool
		for j, y := range xs {
			if neq = compare(vs[i+j], y) != 0; neq {
				break
			}
		}
		if !neq {
			rs = append(rs, i)
		}
	}
	return rs
}

func funcSlice(_, v, end, start interface{}) (r interface{}) {
	if w, ok := v.(string); ok {
		v = explode(w)
		defer func() {
			switch s := r.(type) {
			case []interface{}:
				r = implode(s)
			case int:
				r = implode([]interface{}{s})
			case nil:
				r = ""
			case error:
			default:
				panic(r)
			}
		}()
	}
	switch v := v.(type) {
	case nil:
		return nil
	case []interface{}:
		if start != nil {
			if start, ok := toInt(start); ok {
				if end != nil {
					if end, ok := toInt(end); ok {
						return funcIndexSlice(&start, &end, nil, v)
					}
					return &arrayIndexNotNumberError{end}
				}
				return funcIndexSlice(&start, nil, nil, v)
			}
			return &arrayIndexNotNumberError{start}
		}
		if end != nil {
			if end, ok := toInt(end); ok {
				return funcIndexSlice(nil, &end, nil, v)
			}
			return &arrayIndexNotNumberError{end}
		}
		return v
	default:
		return &expectedArrayError{v}
	}
}

func funcIndexSlice(start, end, index *int, a []interface{}) interface{} {
	aa := a
	if index != nil {
		i := toIndex(aa, *index)
		if i < 0 {
			return nil
		}
		return a[i]
	}
	if end != nil {
		i := toIndex(aa, *end)
		if i == -1 {
			i = len(a)
		} else if i == -2 {
			i = 0
		}
		a = a[:i]
	}
	if start != nil {
		i := toIndex(aa, *start)
		if i == -1 || len(a) < i {
			i = len(a)
		} else if i == -2 {
			i = 0
		}
		a = a[i:]
	}
	return a
}

func toIndex(a []interface{}, i int) int {
	l := len(a)
	switch {
	case i < -l:
		return -2
	case i < 0:
		return l + i
	case i < l:
		return i
	default:
		return -1
	}
}

func funcIndices(v, x interface{}) interface{} {
	return indexFunc(v, x, indices)
}

func funcLindex(v, x interface{}) interface{} {
	return indexFunc(v, x, func(vs, xs []interface{}) interface{} {
		if len(xs) == 0 {
			return nil
		}
		for i := 0; i < len(vs) && i < len(vs)-len(xs)+1; i++ {
			var neq bool
			for j, y := range xs {
				if neq = compare(vs[i+j], y) != 0; neq {
					break
				}
			}
			if !neq {
				return i
			}
		}
		return nil
	})
}

func funcRindex(v, x interface{}) interface{} {
	return indexFunc(v, x, func(vs, xs []interface{}) interface{} {
		if len(xs) == 0 {
			return nil
		}
		i := len(vs) - 1
		if j := len(vs) - len(xs); j < i {
			i = j
		}
		for ; i >= 0; i-- {
			var neq bool
			for j, y := range xs {
				if neq = compare(vs[i+j], y) != 0; neq {
					break
				}
			}
			if !neq {
				return i
			}
		}
		return nil
	})
}

func indexFunc(v, x interface{}, f func(_, _ []interface{}) interface{}) interface{} {
	switch v := v.(type) {
	case nil:
		return nil
	case []interface{}:
		switch x := x.(type) {
		case []interface{}:
			return f(v, x)
		default:
			return f(v, []interface{}{x})
		}
	case string:
		if x, ok := x.(string); ok {
			return f(explode(v), explode(x))
		}
		return &expectedStringError{x}
	default:
		return &expectedArrayError{v}
	}
}

type rangeIter struct {
	value, end, step interface{}
}

func (iter *rangeIter) Next() (interface{}, bool) {
	if compare(iter.step, 0)*compare(iter.value, iter.end) >= 0 {
		return nil, false
	}
	v := iter.value
	iter.value = funcOpAdd(nil, v, iter.step)
	return v, true
}

func funcRange(_ interface{}, xs []interface{}) interface{} {
	for _, x := range xs {
		switch x.(type) {
		case int, float64, *big.Int:
		default:
			return &funcTypeError{"range", x}
		}
	}
	return &rangeIter{xs[0], xs[1], xs[2]}
}

func funcMinBy(v, x interface{}) interface{} {
	vs, ok := v.([]interface{})
	if !ok {
		return &expectedArrayError{v}
	}
	xs, ok := x.([]interface{})
	if !ok {
		return &expectedArrayError{x}
	}
	if len(vs) != len(xs) {
		return &lengthMismatchError{"min_by", vs, xs}
	}
	return funcMinMaxBy(vs, xs, true)
}

func funcMaxBy(v, x interface{}) interface{} {
	vs, ok := v.([]interface{})
	if !ok {
		return &expectedArrayError{v}
	}
	xs, ok := x.([]interface{})
	if !ok {
		return &expectedArrayError{x}
	}
	if len(vs) != len(xs) {
		return &lengthMismatchError{"max_by", vs, xs}
	}
	return funcMinMaxBy(vs, xs, false)
}

func funcMinMaxBy(vs, xs []interface{}, isMin bool) interface{} {
	if len(vs) == 0 {
		return nil
	}
	i, j, x := 0, 0, xs[0]
	for i++; i < len(xs); i++ {
		if (compare(x, xs[i]) > 0) == isMin {
			j, x = i, xs[i]
		}
	}
	return vs[j]
}

type sortItem struct {
	value, key interface{}
}

func funcSortBy(v, x interface{}) interface{} {
	items, err := sortItems("sort_by", v, x)
	if err != nil {
		return err
	}
	rs := make([]interface{}, len(items))
	for i, x := range items {
		rs[i] = x.value
	}
	return rs
}

func funcGroupBy(v, x interface{}) interface{} {
	items, err := sortItems("group_by", v, x)
	if err != nil {
		return err
	}
	var rs []interface{}
	var last interface{}
	for i, r := range items {
		if i == 0 || compare(last, r.key) != 0 {
			rs, last = append(rs, []interface{}{r.value}), r.key
		} else {
			rs[len(rs)-1] = append(rs[len(rs)-1].([]interface{}), r.value)
		}
	}
	return rs
}

func funcUniqueBy(v, x interface{}) interface{} {
	items, err := sortItems("unique_by", v, x)
	if err != nil {
		return err
	}
	var rs []interface{}
	var last interface{}
	for i, r := range items {
		if i == 0 || compare(last, r.key) != 0 {
			rs, last = append(rs, r.value), r.key
		}
	}
	return rs
}

func funcJoin(v, x interface{}) interface{} {
	vs, ok := v.([]interface{})
	if !ok {
		return &expectedArrayError{v}
	}
	if len(vs) == 0 {
		return ""
	}
	sep, ok := x.(string)
	if len(vs) > 1 && !ok {
		return &funcTypeError{"join", x}
	}
	ss := make([]string, len(vs))
	for i, e := range vs {
		switch e := e.(type) {
		case nil:
		case string:
			ss[i] = e
		case bool:
			if e {
				ss[i] = "true"
			} else {
				ss[i] = "false"
			}
		case int, float64, *big.Int:
			ss[i] = jsonMarshal(e)
		default:
			return &unaryTypeError{"join", e}
		}
	}
	return strings.Join(ss, sep)
}

func sortItems(name string, v, x interface{}) ([]*sortItem, error) {
	vs, ok := v.([]interface{})
	if !ok {
		return nil, &expectedArrayError{v}
	}
	xs, ok := x.([]interface{})
	if !ok {
		return nil, &expectedArrayError{x}
	}
	if len(vs) != len(xs) {
		return nil, &lengthMismatchError{name, vs, xs}
	}
	items := make([]*sortItem, len(vs))
	for i, v := range vs {
		items[i] = &sortItem{v, xs[i]}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return compare(items[i].key, items[j].key) < 0
	})
	return items, nil
}

func funcSignificand(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v == 0.0 {
		return v
	}
	return math.Float64frombits((math.Float64bits(v) & 0x800fffffffffffff) | 0x3ff0000000000000)
}

func funcExp10(v float64) float64 {
	return math.Pow(10, v)
}

func funcFrexp(v interface{}) interface{} {
	x, ok := toFloat(v)
	if !ok {
		return &funcTypeError{"frexp", v}
	}
	f, e := math.Frexp(x)
	return []interface{}{f, e}
}

func funcModf(v interface{}) interface{} {
	x, ok := toFloat(v)
	if !ok {
		return &funcTypeError{"modf", v}
	}
	i, f := math.Modf(x)
	return []interface{}{f, i}
}

func funcLgamma(v float64) float64 {
	v, _ = math.Lgamma(v)
	return v
}

func funcDrem(l, r float64) float64 {
	x := math.Remainder(l, r)
	if x == 0.0 {
		return math.Copysign(x, l)
	}
	return x
}

func funcJn(l, r float64) float64 {
	return math.Jn(int(l), r)
}

func funcLdexp(l, r float64) float64 {
	return math.Ldexp(l, int(r))
}

func funcScalb(l, r float64) float64 {
	return l * math.Pow(2, r)
}

func funcScalbln(l, r float64) float64 {
	return l * math.Pow(2, r)
}

func funcYn(l, r float64) float64 {
	return math.Yn(int(l), r)
}

func funcInfinite(interface{}) interface{} {
	return math.Inf(1)
}

func funcIsfinite(v interface{}) interface{} {
	x, ok := toFloat(v)
	return ok && !math.IsInf(x, 0)
}

func funcIsinfinite(v interface{}) interface{} {
	x, ok := toFloat(v)
	return ok && math.IsInf(x, 0)
}

func funcNan(interface{}) interface{} {
	return math.NaN()
}

func funcIsnan(v interface{}) interface{} {
	x, ok := toFloat(v)
	if !ok {
		if v == nil {
			return false
		}
		return &funcTypeError{"isnan", v}
	}
	return math.IsNaN(x)
}

func funcIsnormal(v interface{}) interface{} {
	x, ok := toFloat(v)
	return ok && !math.IsNaN(x) && !math.IsInf(x, 0) && x != 0.0
}

func funcSetpath(v, p, w interface{}) interface{} {
	path, ok := p.([]interface{})
	if !ok {
		return &funcTypeError{"setpath", p}
	}
	var err error
	if v, err = updatePaths(v, path, w, false); err != nil {
		if err, ok := err.(*funcTypeError); ok {
			err.name = "setpath"
		}
		return err
	}
	return v
}

func funcDelpaths(v, p interface{}) interface{} {
	paths, ok := p.([]interface{})
	if !ok {
		return &funcTypeError{"delpaths", p}
	}
	// Fills the paths with an empty value and then delete them. We cannot delete
	// in each loop because array indices should not change. For example,
	//   jq -n "[0, 1, 2, 3] | delpaths([[1], [2]])" #=> [0, 3].
	var empty struct{}
	var err error
	for _, p := range paths {
		path, ok := p.([]interface{})
		if !ok {
			return &funcTypeError{"delpaths", p}
		}
		if v, err = updatePaths(v, path, empty, true); err != nil {
			return err
		}
	}
	return deleteEmpty(v)
}

func updatePaths(v interface{}, path []interface{}, w interface{}, delpaths bool) (interface{}, error) {
	if len(path) == 0 {
		return w, nil
	}
	switch x := path[0].(type) {
	case string:
		if v == nil {
			if delpaths {
				return v, nil
			}
			v = make(map[string]interface{})
		}
		switch uu := v.(type) {
		case map[string]interface{}:
			if _, ok := uu[x]; !ok && delpaths {
				return v, nil
			}
			u, err := updatePaths(uu[x], path[1:], w, delpaths)
			if err != nil {
				return nil, err
			}
			vs := make(map[string]interface{}, len(uu))
			for k, v := range uu {
				vs[k] = v
			}
			vs[x] = u
			return vs, nil
		case struct{}:
			return v, nil
		default:
			return nil, &expectedObjectError{v}
		}
	case int, float64, *big.Int:
		if v == nil {
			if delpaths {
				return v, nil
			}
			v = []interface{}{}
		}
		switch uu := v.(type) {
		case []interface{}:
			y, _ := toInt(x)
			l := len(uu)
			var copied bool
			if copied = y >= l; copied {
				if delpaths {
					return v, nil
				}
				if y > 0x3ffffff {
					return nil, &arrayIndexTooLargeError{y}
				}
				l = y + 1
				ys := make([]interface{}, l)
				copy(ys, uu)
				uu = ys
			} else if y < -l {
				if delpaths {
					return v, nil
				}
				return nil, &funcTypeError{v: y}
			} else if y < 0 {
				y += l
			}
			u, err := updatePaths(uu[y], path[1:], w, delpaths)
			if err != nil {
				return nil, err
			}
			if copied {
				uu[y] = u
				return uu, nil
			}
			vs := make([]interface{}, l)
			copy(vs, uu)
			vs[y] = u
			return vs, nil
		case struct{}:
			return v, nil
		default:
			return nil, &expectedArrayError{v}
		}
	case map[string]interface{}:
		if len(x) == 0 {
			switch v.(type) {
			case []interface{}:
				return nil, &arrayIndexNotNumberError{x}
			default:
				return nil, &objectKeyNotStringError{x}
			}
		}
		if v == nil {
			v = []interface{}{}
		}
		switch uu := v.(type) {
		case []interface{}:
			var start, end int
			if x, ok := toInt(x["start"]); ok {
				x := toIndex(uu, x)
				if x > len(uu) || x == -1 {
					start = len(uu)
				} else if x == -2 {
					start = 0
				} else {
					start = x
				}
			}
			if x, ok := toInt(x["end"]); ok {
				x := toIndex(uu, x)
				if x == -1 {
					end = len(uu)
				} else if x < start {
					end = start
				} else {
					end = x
				}
			} else {
				end = len(uu)
			}
			if delpaths {
				if start >= end {
					return uu, nil
				}
				if len(path) > 1 {
					u, err := updatePaths(uu[start:end], path[1:], w, delpaths)
					if err != nil {
						return nil, err
					}
					switch us := u.(type) {
					case []interface{}:
						vs := make([]interface{}, len(uu))
						copy(vs, uu)
						copy(vs[start:end], us)
						return vs, nil
					default:
						return nil, &expectedArrayError{u}
					}
				}
				vs := make([]interface{}, len(uu))
				copy(vs, uu)
				for y := start; y < end; y++ {
					vs[y] = w
				}
				return vs, nil
			}
			if len(path) > 1 {
				u, err := updatePaths(uu[start:end], path[1:], w, delpaths)
				if err != nil {
					return nil, err
				}
				w = u
			}
			switch v := w.(type) {
			case []interface{}:
				vs := make([]interface{}, start+len(v)+len(uu)-end)
				copy(vs, uu[:start])
				copy(vs[start:], v)
				copy(vs[start+len(v):], uu[end:])
				return vs, nil
			default:
				return nil, &expectedArrayError{v}
			}
		case struct{}:
			return v, nil
		default:
			return nil, &expectedArrayError{v}
		}
	default:
		switch v.(type) {
		case []interface{}:
			return nil, &arrayIndexNotNumberError{x}
		default:
			return nil, &objectKeyNotStringError{x}
		}
	}
}

func funcGetpath(v, p interface{}) interface{} {
	keys, ok := p.([]interface{})
	if !ok {
		return &funcTypeError{"getpath", p}
	}
	u := v
	for _, x := range keys {
		switch v.(type) {
		case map[string]interface{}:
		case []interface{}:
		case nil:
		default:
			return &getpathError{u, p}
		}
		v = funcIndex(nil, v, x)
		if _, ok := v.(error); ok {
			return &getpathError{u, p}
		}
	}
	return v
}

func funcTranspose(v interface{}) interface{} {
	vss, ok := v.([]interface{})
	if !ok {
		return &funcTypeError{"transpose", v}
	}
	if len(vss) == 0 {
		return []interface{}{}
	}
	var l int
	for _, vs := range vss {
		vs, ok := vs.([]interface{})
		if !ok {
			return &funcTypeError{"transpose", v}
		}
		if k := len(vs); l < k {
			l = k
		}
	}
	wss := make([][]interface{}, l)
	xs := make([]interface{}, l)
	for i, k := 0, len(vss); i < l; i++ {
		s := make([]interface{}, k)
		wss[i] = s
		xs[i] = s
	}
	for i, vs := range vss {
		for j, v := range vs.([]interface{}) {
			wss[j][i] = v
		}
	}
	return xs
}

func funcBsearch(v, t interface{}) interface{} {
	vs, ok := v.([]interface{})
	if !ok {
		return &funcTypeError{"bsearch", v}
	}
	i := sort.Search(len(vs), func(i int) bool {
		return compare(vs[i], t) >= 0
	})
	if i < len(vs) && compare(vs[i], t) == 0 {
		return i
	}
	return -i - 1
}

func funcGmtime(v interface{}) interface{} {
	if v, ok := toFloat(v); ok {
		return epochToArray(v, time.UTC)
	}
	return &funcTypeError{"gmtime", v}
}

func funcLocaltime(v interface{}) interface{} {
	if v, ok := toFloat(v); ok {
		return epochToArray(v, time.Local)
	}
	return &funcTypeError{"localtime", v}
}

func epochToArray(v float64, loc *time.Location) []interface{} {
	t := time.Unix(int64(v), int64((v-math.Floor(v))*1e9)).In(loc)
	return []interface{}{
		t.Year(),
		int(t.Month()) - 1,
		t.Day(),
		t.Hour(),
		t.Minute(),
		float64(t.Second()) + float64(t.Nanosecond())/1e9,
		int(t.Weekday()),
		t.YearDay() - 1,
	}
}

func funcMktime(v interface{}) interface{} {
	if a, ok := v.([]interface{}); ok {
		t, err := arrayToTime("mktime", a, time.UTC)
		if err != nil {
			return err
		}
		return float64(t.Unix())
	}
	return &funcTypeError{"mktime", v}
}

func funcStrftime(v, x interface{}) interface{} {
	if w, ok := toFloat(v); ok {
		v = epochToArray(w, time.UTC)
	}
	if a, ok := v.([]interface{}); ok {
		if format, ok := x.(string); ok {
			t, err := arrayToTime("strftime", a, time.UTC)
			if err != nil {
				return err
			}
			return timefmt.Format(t, format)
		}
		return &funcTypeError{"strftime", x}
	}
	return &funcTypeError{"strftime", v}
}

func funcStrflocaltime(v, x interface{}) interface{} {
	if w, ok := toFloat(v); ok {
		v = epochToArray(w, time.Local)
	}
	if a, ok := v.([]interface{}); ok {
		if format, ok := x.(string); ok {
			t, err := arrayToTime("strflocaltime", a, time.Local)
			if err != nil {
				return err
			}
			return timefmt.Format(t, format)
		}
		return &funcTypeError{"strflocaltime", x}
	}
	return &funcTypeError{"strflocaltime", v}
}

func funcStrptime(v, x interface{}) interface{} {
	if v, ok := v.(string); ok {
		if format, ok := x.(string); ok {
			t, err := timefmt.Parse(v, format)
			if err != nil {
				return err
			}
			var s time.Time
			if t == s {
				return &funcTypeError{"strptime", v}
			}
			return epochToArray(float64(t.Unix())+float64(t.Nanosecond())/1e9, time.UTC)
		}
		return &funcTypeError{"strptime", x}
	}
	return &funcTypeError{"strptime", v}
}

func arrayToTime(name string, a []interface{}, loc *time.Location) (time.Time, error) {
	var t time.Time
	if len(a) != 8 {
		return t, &funcTypeError{name, a}
	}
	var y, m, d, h, min, sec, nsec int
	if x, ok := toInt(a[0]); ok {
		y = x
	} else {
		return t, &funcTypeError{name, a}
	}
	if x, ok := toInt(a[1]); ok {
		m = x + 1
	} else {
		return t, &funcTypeError{name, a}
	}
	if x, ok := toInt(a[2]); ok {
		d = x
	} else {
		return t, &funcTypeError{name, a}
	}
	if x, ok := toInt(a[3]); ok {
		h = x
	} else {
		return t, &funcTypeError{name, a}
	}
	if x, ok := toInt(a[4]); ok {
		min = x
	} else {
		return t, &funcTypeError{name, a}
	}
	if x, ok := toFloat(a[5]); ok {
		sec = int(x)
		nsec = int((x - math.Floor(x)) * 1e9)
	} else {
		return t, &funcTypeError{name, a}
	}
	return time.Date(y, time.Month(m), d, h, min, sec, nsec, loc), nil
}

func funcNow(interface{}) interface{} {
	t := time.Now()
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}

func funcMatch(v, re, fs, testing interface{}) interface{} {
	var flags string
	if fs != nil {
		v, ok := fs.(string)
		if !ok {
			return &funcTypeError{"match", fs}
		}
		flags = v
	}
	s, ok := v.(string)
	if !ok {
		return &funcTypeError{"match", v}
	}
	restr, ok := re.(string)
	if !ok {
		return &funcTypeError{"match", v}
	}
	r, err := compileRegexp(restr, flags)
	if err != nil {
		return err
	}
	var xs [][]int
	if strings.ContainsRune(flags, 'g') && testing != true {
		xs = r.FindAllStringSubmatchIndex(s, -1)
	} else {
		got := r.FindStringSubmatchIndex(s)
		if testing == true {
			return got != nil
		}
		if got != nil {
			xs = [][]int{got}
		}
	}
	res, names := make([]interface{}, len(xs)), r.SubexpNames()
	for i, x := range xs {
		captures := make([]interface{}, (len(x)-2)/2)
		for j := 1; j < len(x)/2; j++ {
			var name interface{}
			if n := names[j]; n != "" {
				name = n
			}
			if x[j*2] < 0 {
				captures[j-1] = map[string]interface{}{
					"name":   name,
					"offset": -1,
					"length": 0,
					"string": nil,
				}
				continue
			}
			captures[j-1] = map[string]interface{}{
				"name":   name,
				"offset": len([]rune(s[:x[j*2]])),
				"length": len([]rune(s[:x[j*2+1]])) - len([]rune(s[:x[j*2]])),
				"string": s[x[j*2]:x[j*2+1]],
			}
		}
		res[i] = map[string]interface{}{
			"offset":   len([]rune(s[:x[0]])),
			"length":   len([]rune(s[:x[1]])) - len([]rune(s[:x[0]])),
			"string":   s[x[0]:x[1]],
			"captures": captures,
		}
	}
	return res
}

func compileRegexp(re, flags string) (*regexp.Regexp, error) {
	if strings.IndexFunc(flags, func(r rune) bool {
		return r != 'g' && r != 'i' && r != 'm'
	}) >= 0 {
		return nil, fmt.Errorf("unsupported regular expression flag: %q", flags)
	}
	re = strings.ReplaceAll(re, "(?<", "(?P<")
	if strings.ContainsRune(flags, 'i') {
		re = "(?i)" + re
	}
	if strings.ContainsRune(flags, 'm') {
		re = "(?s)" + re
	}
	r, err := regexp.Compile(re)
	if err != nil {
		return nil, fmt.Errorf("invalid regular expression %q: %s", re, err)
	}
	return r, nil
}

func funcError(v interface{}, args []interface{}) interface{} {
	if len(args) > 0 {
		v = args[0]
	}
	code := 5
	if v == nil {
		code = 0
	}
	return &exitCodeError{v, code, false}
}

func funcHalt(interface{}) interface{} {
	return &exitCodeError{nil, 0, true}
}

func funcHaltError(v interface{}, args []interface{}) interface{} {
	code := 5
	if len(args) > 0 {
		var ok bool
		if code, ok = toInt(args[0]); !ok {
			return &funcTypeError{"halt_error", args[0]}
		}
	}
	return &exitCodeError{v, code, true}
}

func internalfuncTypeError(v, x interface{}) interface{} {
	if x, ok := x.(string); ok {
		return &funcTypeError{x, v}
	}
	return &funcTypeError{"_type_error", v}
}

func toInt(x interface{}) (int, bool) {
	switch x := x.(type) {
	case int:
		return x, true
	case float64:
		return floatToInt(x), true
	case *big.Int:
		if x.IsInt64() {
			if i := x.Int64(); minInt <= i && i <= maxInt {
				return int(i), true
			}
		}
		if x.Sign() > 0 {
			return maxInt, true
		}
		return minInt, true
	default:
		return 0, false
	}
}

func floatToInt(x float64) int {
	if minInt <= x && x <= maxInt {
		return int(x)
	}
	if x > 0 {
		return maxInt
	}
	return minInt
}

func toFloat(x interface{}) (float64, bool) {
	switch x := x.(type) {
	case int:
		return float64(x), true
	case float64:
		return x, true
	case *big.Int:
		return bigToFloat(x), true
	default:
		return 0.0, false
	}
}

func bigToFloat(x *big.Int) float64 {
	if x.IsInt64() {
		return float64(x.Int64())
	}
	if f, err := strconv.ParseFloat(x.String(), 64); err == nil {
		return f
	}
	return math.Inf(x.Sign())
}
