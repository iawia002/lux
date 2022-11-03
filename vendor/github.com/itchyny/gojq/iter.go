package gojq

// Iter is an interface for an iterator.
type Iter interface {
	Next() (interface{}, bool)
}

// NewIter creates a new Iter from values.
func NewIter(values ...interface{}) Iter {
	switch len(values) {
	case 0:
		return emptyIter{}
	case 1:
		return &unitIter{value: values[0]}
	default:
		iter := sliceIter(values)
		return &iter
	}
}

type emptyIter struct{}

func (emptyIter) Next() (interface{}, bool) {
	return nil, false
}

type unitIter struct {
	value interface{}
	done  bool
}

func (iter *unitIter) Next() (interface{}, bool) {
	if iter.done {
		return nil, false
	}
	iter.done = true
	return iter.value, true
}

type sliceIter []interface{}

func (iter *sliceIter) Next() (interface{}, bool) {
	if len(*iter) == 0 {
		return nil, false
	}
	value := (*iter)[0]
	*iter = (*iter)[1:]
	return value, true
}
