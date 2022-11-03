package gojq

type scopeStack struct {
	data  []scopeBlock
	index int
	limit int
}

type scopeBlock struct {
	value scope
	next  int
}

func newScopeStack() *scopeStack {
	return &scopeStack{index: -1, limit: -1}
}

func (s *scopeStack) push(v scope) {
	b := scopeBlock{v, s.index}
	i := s.index + 1
	if i <= s.limit {
		i = s.limit + 1
	}
	s.index = i
	if i < len(s.data) {
		s.data[i] = b
	} else {
		s.data = append(s.data, b)
	}
}

func (s *scopeStack) pop() scope {
	b := s.data[s.index]
	s.index = b.next
	return b.value
}

func (s *scopeStack) empty() bool {
	return s.index < 0
}

func (s *scopeStack) save(index, limit *int) {
	*index, *limit = s.index, s.limit
	if s.index > s.limit {
		s.limit = s.index
	}
}

func (s *scopeStack) restore(index, limit int) {
	s.index, s.limit = index, limit
}
