package gojq

import (
	"math"
	"math/big"
)

func compare(l, r interface{}) int {
	return binopTypeSwitch(l, r,
		func(l, r int) interface{} {
			switch {
			case l < r:
				return -1
			case l == r:
				return 0
			default:
				return 1
			}
		},
		func(l, r float64) interface{} {
			switch {
			case l < r || math.IsNaN(l):
				return -1
			case l == r:
				return 0
			default:
				return 1
			}
		},
		func(l, r *big.Int) interface{} {
			return l.Cmp(r)
		},
		func(l, r string) interface{} {
			switch {
			case l < r:
				return -1
			case l == r:
				return 0
			default:
				return 1
			}
		},
		func(l, r []interface{}) interface{} {
			for i := 0; ; i++ {
				if i >= len(l) {
					if i >= len(r) {
						return 0
					}
					return -1
				}
				if i >= len(r) {
					return 1
				}
				if cmp := compare(l[i], r[i]); cmp != 0 {
					return cmp
				}
			}
		},
		func(l, r map[string]interface{}) interface{} {
			lk, rk := funcKeys(l), funcKeys(r)
			if cmp := compare(lk, rk); cmp != 0 {
				return cmp
			}
			for _, k := range lk.([]interface{}) {
				if cmp := compare(l[k.(string)], r[k.(string)]); cmp != 0 {
					return cmp
				}
			}
			return 0
		},
		func(l, r interface{}) interface{} {
			ln, rn := getTypeOrdNum(l), getTypeOrdNum(r)
			switch {
			case ln < rn:
				return -1
			case ln == rn:
				return 0
			default:
				return 1
			}
		},
	).(int)
}

func getTypeOrdNum(v interface{}) int {
	switch v := v.(type) {
	case nil:
		return 0
	case bool:
		if v {
			return 2
		}
		return 1
	case int, float64, *big.Int:
		return 3
	case string:
		return 4
	case []interface{}:
		return 5
	case map[string]interface{}:
		return 6
	default:
		return -1
	}
}
