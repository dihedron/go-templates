package main

import "sync"

type element struct {
	data interface{}
	next *element
}

// Stack is a generic, synchronised implementation of a LIFO structure.
type Stack struct {
	lock *sync.Mutex
	head *element
	size int
}

// NewStack creates a new Stack.
func NewStack() *Stack {
	return &Stack{
		lock: &sync.Mutex{},
	}
}

// Push adds a new element onto the Stack.
func (s *Stack) Push(data interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.head = &element{
		data: data,
		next: s.head,
	}
	s.size++
}

// Pop removes the element at the top of the Stack and returns it.
func (s *Stack) Pop() interface{} {
	if s.head == nil {
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	r := s.head.data
	s.head = s.head.next
	s.size--
	return r
}

// Top returns the element at the top of the Stack, without removing it.
func (s *Stack) Top() interface{} {
	return s.head.data
}

// Len returns the size of the Stack.
func (s *Stack) Len() int {
	return s.size
}

// IsEmpty returns wheter the stack ontains no elements.
func (s *Stack) IsEmpty() bool {
	return s.size == 0
}
