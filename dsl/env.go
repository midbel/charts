package dsl

import (
	"fmt"
)

type environ[T any] struct {
	values map[string]T
}

func emptyEnv[T any]() *environ[T] {
	return &environ[T]{
		values: make(map[string]T),
	}
}

func (e *environ[T]) Resolve(name string) (T, error) {
	var zero T
	v, ok := e.values[name]
	if !ok {
		return zero, fmt.Errorf("%s undefined variable")
	}
	return v, nil
}

func (e *environ[T]) Define(name string, values T) {
	e.values[name] = values
}
