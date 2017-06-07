package main

type Stack struct {
	data []interface{}
}

func NewStack() *Stack {
	return &Stack{
		data: []interface{}{},
	}
}

func (s *Stack) Push(item interface{}) {
	s.data = append(s.data, item)
}

func (s *Stack) Pop() interface{} {
	if len(s.data) > 0 {
		item := s.data[len(s.data)-1]
		s.data = s.data[:len(s.data)-1]
		return item
	}
	return nil
}

func (s *Stack) Peek() interface{} {
	if len(s.data) > 0 {
		item := s.data[len(s.data)-1]
		return item
	}
	return nil
}

func (s *Stack) Len() int {
	return len(s.data)
}

func (s *Stack) IsEmpty() bool {
	return s.Len() == 0
}

func (s *Stack) Clear() {
	s.data = s.data[:0]
}
