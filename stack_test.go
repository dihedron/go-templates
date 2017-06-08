package main

import "testing"

func TestPushLenPop(t *testing.T) {
	stack := NewStack()

	for i := 0; i < 100; i++ {
		stack.Push(i)

		if stack.Top().(int) != i {
			t.Errorf("invalid top element: expected i, got %d", stack.Top().(int))
		}
	}

	if stack.Len() != 100 {
		t.Errorf("invalid stack length: expected 100, got %d", stack.Len())
	}

	for i := 0; i < 100; i++ {
		j := stack.Pop().(int)
		if i+j != 99 {
			t.Errorf("invalid stack value: expected %d, got %d", 99-i, j)
		}
	}

	if !stack.IsEmpty() {
		t.Errorf("invalid stack length: expected 0, got %d", stack.Len())
	}
}
