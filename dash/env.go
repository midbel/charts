package dash

import (
	"fmt"
)

type Environ[T any] struct {
	values map[string]T
}

func EmptyEnv[T any]() *Environ[T] {
	return &Environ[T]{
		values: make(map[string]T),
	}
}

func (e *Environ[T]) Resolve(name string) (T, error) {
	var zero T
	v, ok := e.values[name]
	if !ok {
		return zero, fmt.Errorf("%s undefined variable")
	}
	return v, nil
}

func (e *Environ[T]) Define(name string, values T) {
	e.values[name] = values
}
