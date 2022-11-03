package gojq

import (
	"math"
	"math/big"
)

func deepEqual(l, r interface{}) bool {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} {
			return l == r
		},
		func(l, r float64) interface{} {
			return l == r || math.IsNaN(l) && math.IsNaN(r)
		},
		func(l, r *big.Int) interface{} {
			return l.Cmp(r) == 0
		},
		func(l, r string) interface{} {
			return l == r
		},
		func(l, r []interface{}) interface{} {
			if len(l) != len(r) {
				return false
			}
			for i, v := range l {
				if !deepEqual(v, r[i]) {
					return false
				}
			}
			return true
		},
		func(l, r map[string]interface{}) interface{} {
			if len(l) != len(r) {
				return false
			}
			for k, v := range l {
				if !deepEqual(v, r[k]) {
					return false
				}
			}
			return true
		},
		func(l, r interface{}) interface{} {
			return l == r
		},
	).(bool)
}
