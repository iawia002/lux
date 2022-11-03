package gojq

import (
	"math"
	"math/big"
	"strings"
)

// Operator ...
type Operator int

// Operators ...
const (
	OpPipe Operator = iota + 1
	OpComma
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEq
	OpNe
	OpGt
	OpLt
	OpGe
	OpLe
	OpAnd
	OpOr
	OpAlt
	OpAssign
	OpModify
	OpUpdateAdd
	OpUpdateSub
	OpUpdateMul
	OpUpdateDiv
	OpUpdateMod
	OpUpdateAlt
)

// String implements Stringer.
func (op Operator) String() string {
	switch op {
	case OpPipe:
		return "|"
	case OpComma:
		return ","
	case OpAdd:
		return "+"
	case OpSub:
		return "-"
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpMod:
		return "%"
	case OpEq:
		return "=="
	case OpNe:
		return "!="
	case OpGt:
		return ">"
	case OpLt:
		return "<"
	case OpGe:
		return ">="
	case OpLe:
		return "<="
	case OpAnd:
		return "and"
	case OpOr:
		return "or"
	case OpAlt:
		return "//"
	case OpAssign:
		return "="
	case OpModify:
		return "|="
	case OpUpdateAdd:
		return "+="
	case OpUpdateSub:
		return "-="
	case OpUpdateMul:
		return "*="
	case OpUpdateDiv:
		return "/="
	case OpUpdateMod:
		return "%="
	case OpUpdateAlt:
		return "//="
	default:
		panic(op)
	}
}

// GoString implements GoStringer.
func (op Operator) GoString() (str string) {
	defer func() { str = "gojq." + str }()
	switch op {
	case Operator(0):
		return "Operator(0)"
	case OpPipe:
		return "OpPipe"
	case OpComma:
		return "OpComma"
	case OpAdd:
		return "OpAdd"
	case OpSub:
		return "OpSub"
	case OpMul:
		return "OpMul"
	case OpDiv:
		return "OpDiv"
	case OpMod:
		return "OpMod"
	case OpEq:
		return "OpEq"
	case OpNe:
		return "OpNe"
	case OpGt:
		return "OpGt"
	case OpLt:
		return "OpLt"
	case OpGe:
		return "OpGe"
	case OpLe:
		return "OpLe"
	case OpAnd:
		return "OpAnd"
	case OpOr:
		return "OpOr"
	case OpAlt:
		return "OpAlt"
	case OpAssign:
		return "OpAssign"
	case OpModify:
		return "OpModify"
	case OpUpdateAdd:
		return "OpUpdateAdd"
	case OpUpdateSub:
		return "OpUpdateSub"
	case OpUpdateMul:
		return "OpUpdateMul"
	case OpUpdateDiv:
		return "OpUpdateDiv"
	case OpUpdateMod:
		return "OpUpdateMod"
	case OpUpdateAlt:
		return "OpUpdateAlt"
	default:
		panic(op)
	}
}

func (op Operator) getFunc() string {
	switch op {
	case OpPipe:
		panic("unreachable")
	case OpComma:
		panic("unreachable")
	case OpAdd:
		return "_add"
	case OpSub:
		return "_subtract"
	case OpMul:
		return "_multiply"
	case OpDiv:
		return "_divide"
	case OpMod:
		return "_modulo"
	case OpEq:
		return "_equal"
	case OpNe:
		return "_notequal"
	case OpGt:
		return "_greater"
	case OpLt:
		return "_less"
	case OpGe:
		return "_greatereq"
	case OpLe:
		return "_lesseq"
	case OpAnd:
		panic("unreachable")
	case OpOr:
		panic("unreachable")
	case OpAlt:
		panic("unreachable")
	case OpAssign:
		return "_assign"
	case OpModify:
		return "_modify"
	case OpUpdateAdd:
		return "_add"
	case OpUpdateSub:
		return "_subtract"
	case OpUpdateMul:
		return "_multiply"
	case OpUpdateDiv:
		return "_divide"
	case OpUpdateMod:
		return "_modulo"
	case OpUpdateAlt:
		return "_alternative"
	default:
		panic(op)
	}
}

func binopTypeSwitch(
	l, r interface{},
	callbackInts func(_, _ int) interface{},
	callbackFloats func(_, _ float64) interface{},
	callbackBigInts func(_, _ *big.Int) interface{},
	callbackStrings func(_, _ string) interface{},
	callbackArrays func(_, _ []interface{}) interface{},
	callbackMaps func(_, _ map[string]interface{}) interface{},
	fallback func(_, _ interface{}) interface{}) interface{} {
	switch l := l.(type) {
	case int:
		switch r := r.(type) {
		case int:
			if minHalfInt <= l && l <= maxHalfInt &&
				minHalfInt <= r && r <= maxHalfInt {
				return callbackInts(l, r)
			}
			return callbackBigInts(big.NewInt(int64(l)), big.NewInt(int64(r)))
		case float64:
			return callbackFloats(float64(l), r)
		case *big.Int:
			return callbackBigInts(big.NewInt(int64(l)), r)
		default:
			return fallback(l, r)
		}
	case float64:
		switch r := r.(type) {
		case int:
			return callbackFloats(l, float64(r))
		case float64:
			return callbackFloats(l, r)
		case *big.Int:
			return callbackFloats(l, bigToFloat(r))
		default:
			return fallback(l, r)
		}
	case *big.Int:
		switch r := r.(type) {
		case int:
			return callbackBigInts(l, big.NewInt(int64(r)))
		case float64:
			return callbackFloats(bigToFloat(l), r)
		case *big.Int:
			return callbackBigInts(l, r)
		default:
			return fallback(l, r)
		}
	case string:
		switch r := r.(type) {
		case string:
			return callbackStrings(l, r)
		default:
			return fallback(l, r)
		}
	case []interface{}:
		switch r := r.(type) {
		case []interface{}:
			return callbackArrays(l, r)
		default:
			return fallback(l, r)
		}
	case map[string]interface{}:
		switch r := r.(type) {
		case map[string]interface{}:
			return callbackMaps(l, r)
		default:
			return fallback(l, r)
		}
	default:
		return fallback(l, r)
	}
}

func funcOpPlus(v interface{}) interface{} {
	switch v := v.(type) {
	case int:
		return v
	case float64:
		return v
	case *big.Int:
		return v
	default:
		return &unaryTypeError{"plus", v}
	}
}

func funcOpNegate(v interface{}) interface{} {
	switch v := v.(type) {
	case int:
		return -v
	case float64:
		return -v
	case *big.Int:
		return new(big.Int).Neg(v)
	default:
		return &unaryTypeError{"negate", v}
	}
}

func funcOpAdd(_, l, r interface{}) interface{} {
	if l == nil {
		return r
	} else if r == nil {
		return l
	}
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} { return l + r },
		func(l, r float64) interface{} { return l + r },
		func(l, r *big.Int) interface{} { return new(big.Int).Add(l, r) },
		func(l, r string) interface{} { return l + r },
		func(l, r []interface{}) interface{} {
			if len(r) == 0 {
				return l
			} else if len(l) == 0 {
				return r
			}
			v := make([]interface{}, 0, len(l)+len(r))
			return append(append(v, l...), r...)
		},
		func(l, r map[string]interface{}) interface{} {
			m := make(map[string]interface{}, len(l)+len(r))
			for k, v := range l {
				m[k] = v
			}
			for k, v := range r {
				m[k] = v
			}
			return m
		},
		func(l, r interface{}) interface{} { return &binopTypeError{"add", l, r} },
	)
}

func funcOpSub(_, l, r interface{}) interface{} {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} { return l - r },
		func(l, r float64) interface{} { return l - r },
		func(l, r *big.Int) interface{} { return new(big.Int).Sub(l, r) },
		func(l, r string) interface{} { return &binopTypeError{"subtract", l, r} },
		func(l, r []interface{}) interface{} {
			a := make([]interface{}, 0, len(l))
			for _, v := range l {
				var found bool
				for _, w := range r {
					if compare(v, w) == 0 {
						found = true
						break
					}
				}
				if !found {
					a = append(a, v)
				}
			}
			return a
		},
		func(l, r map[string]interface{}) interface{} { return &binopTypeError{"subtract", l, r} },
		func(l, r interface{}) interface{} { return &binopTypeError{"subtract", l, r} },
	)
}

func funcOpMul(_, l, r interface{}) interface{} {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} { return l * r },
		func(l, r float64) interface{} { return l * r },
		func(l, r *big.Int) interface{} { return new(big.Int).Mul(l, r) },
		func(l, r string) interface{} { return &binopTypeError{"multiply", l, r} },
		func(l, r []interface{}) interface{} { return &binopTypeError{"multiply", l, r} },
		deepMergeObjects,
		func(l, r interface{}) interface{} {
			multiplyString := func(s string, cnt float64) interface{} {
				if cnt <= 0.0 || cnt > float64(maxHalfInt/(16*(len(s)+1))) || math.IsNaN(cnt) {
					return nil
				}
				if cnt < 1.0 {
					return s
				}
				return strings.Repeat(s, int(cnt))
			}
			if l, ok := l.(string); ok {
				if f, ok := toFloat(r); ok {
					return multiplyString(l, f)
				}
			}
			if r, ok := r.(string); ok {
				if f, ok := toFloat(l); ok {
					return multiplyString(r, f)
				}
			}
			return &binopTypeError{"multiply", l, r}
		},
	)
}

func deepMergeObjects(l, r map[string]interface{}) interface{} {
	m := make(map[string]interface{}, len(l)+len(r))
	for k, v := range l {
		m[k] = v
	}
	for k, v := range r {
		if mk, ok := m[k]; ok {
			if mk, ok := mk.(map[string]interface{}); ok {
				if w, ok := v.(map[string]interface{}); ok {
					v = deepMergeObjects(mk, w)
				}
			}
		}
		m[k] = v
	}
	return m
}

func funcOpDiv(_, l, r interface{}) interface{} {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} {
			if r == 0 {
				if l == 0 {
					return math.NaN()
				}
				return &zeroDivisionError{l, r}
			}
			if l%r == 0 {
				return l / r
			}
			return float64(l) / float64(r)
		},
		func(l, r float64) interface{} {
			if r == 0.0 {
				if l == 0.0 {
					return math.NaN()
				}
				return &zeroDivisionError{l, r}
			}
			return l / r
		},
		func(l, r *big.Int) interface{} {
			if r.Sign() == 0 {
				if l.Sign() == 0 {
					return math.NaN()
				}
				return &zeroDivisionError{l, r}
			}
			d, m := new(big.Int).DivMod(l, r, new(big.Int))
			if m.Sign() == 0 {
				return d
			}
			return bigToFloat(l) / bigToFloat(r)
		},
		func(l, r string) interface{} {
			if l == "" {
				return []interface{}{}
			}
			xs := strings.Split(l, r)
			vs := make([]interface{}, len(xs))
			for i, x := range xs {
				vs[i] = x
			}
			return vs
		},
		func(l, r []interface{}) interface{} { return &binopTypeError{"divide", l, r} },
		func(l, r map[string]interface{}) interface{} { return &binopTypeError{"divide", l, r} },
		func(l, r interface{}) interface{} { return &binopTypeError{"divide", l, r} },
	)
}

func funcOpMod(_, l, r interface{}) interface{} {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} {
			if r == 0 {
				return &zeroModuloError{l, r}
			}
			return l % r
		},
		func(l, r float64) interface{} {
			ri := floatToInt(r)
			if ri == 0 {
				return &zeroModuloError{l, r}
			}
			return floatToInt(l) % ri
		},
		func(l, r *big.Int) interface{} {
			if r.Sign() == 0 {
				return &zeroModuloError{l, r}
			}
			return new(big.Int).Rem(l, r)
		},
		func(l, r string) interface{} { return &binopTypeError{"modulo", l, r} },
		func(l, r []interface{}) interface{} { return &binopTypeError{"modulo", l, r} },
		func(l, r map[string]interface{}) interface{} { return &binopTypeError{"modulo", l, r} },
		func(l, r interface{}) interface{} { return &binopTypeError{"modulo", l, r} },
	)
}

func funcOpAlt(_, l, r interface{}) interface{} {
	if l == nil || l == false {
		return r
	}
	return l
}

func funcOpEq(_, l, r interface{}) interface{} {
	return compare(l, r) == 0
}

func funcOpNe(_, l, r interface{}) interface{} {
	return compare(l, r) != 0
}

func funcOpGt(_, l, r interface{}) interface{} {
	return compare(l, r) > 0
}

func funcOpLt(_, l, r interface{}) interface{} {
	return compare(l, r) < 0
}

func funcOpGe(_, l, r interface{}) interface{} {
	return compare(l, r) >= 0
}

func funcOpLe(_, l, r interface{}) interface{} {
	return compare(l, r) <= 0
}
