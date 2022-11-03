package gojq

import (
	"encoding/json"
	"math"
	"math/big"
	"strings"
)

func normalizeNumber(v json.Number) interface{} {
	if i, err := v.Int64(); err == nil && minInt <= i && i <= maxInt {
		return int(i)
	}
	if strings.ContainsAny(v.String(), ".eE") {
		if f, err := v.Float64(); err == nil {
			return f
		}
	}
	if bi, ok := new(big.Int).SetString(v.String(), 10); ok {
		return bi
	}
	if strings.HasPrefix(v.String(), "-") {
		return math.Inf(-1)
	}
	return math.Inf(1)
}

func normalizeNumbers(v interface{}) interface{} {
	switch v := v.(type) {
	case json.Number:
		return normalizeNumber(v)
	case *big.Int:
		if v.IsInt64() {
			if i := v.Int64(); minInt <= i && i <= maxInt {
				return int(i)
			}
		}
		return v
	case int64:
		if v > maxInt || v < minInt {
			return new(big.Int).SetInt64(v)
		}
		return int(v)
	case int32:
		return int(v)
	case int16:
		return int(v)
	case int8:
		return int(v)
	case uint:
		if v > maxInt {
			return new(big.Int).SetUint64(uint64(v))
		}
		return int(v)
	case uint64:
		if v > maxInt {
			return new(big.Int).SetUint64(v)
		}
		return int(v)
	case uint32:
		if uint64(v) > maxInt {
			return new(big.Int).SetUint64(uint64(v))
		}
		return int(v)
	case uint16:
		return int(v)
	case uint8:
		return int(v)
	case float32:
		return float64(v)
	case map[string]interface{}:
		for k, x := range v {
			v[k] = normalizeNumbers(x)
		}
		return v
	case []interface{}:
		for i, x := range v {
			v[i] = normalizeNumbers(x)
		}
		return v
	default:
		return v
	}
}

// It's ok to delete destructively because this function is used right after
// updatePaths, where it shallow-copies maps or slices on updates.
func deleteEmpty(v interface{}) interface{} {
	switch v := v.(type) {
	case struct{}:
		return nil
	case map[string]interface{}:
		for k, w := range v {
			if w == struct{}{} {
				delete(v, k)
			} else {
				v[k] = deleteEmpty(w)
			}
		}
		return v
	case []interface{}:
		var j int
		for _, w := range v {
			if w != struct{}{} {
				v[j] = deleteEmpty(w)
				j++
			}
		}
		for i := j; i < len(v); i++ {
			v[i] = nil
		}
		return v[:j]
	default:
		return v
	}
}
