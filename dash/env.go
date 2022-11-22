package dash

import (
	"fmt"
)

type Environ[T any] struct {
	parent *Environ[T]
	values map[string]T
}

func EnclosedEnv[T any](parent *Environ[T]) *Environ[T] {
	return &Environ[T]{
		parent: parent,
		values: make(map[string]T),
	}
}

func EmptyEnv[T any]() *Environ[T] {
	return EnclosedEnv[T](nil)
}

func (e *Environ[T]) Wrap() *Environ[T] {
	return EnclosedEnv(e)
}

func (e *Environ[T]) Unwrap() *Environ[T] {
	if e.parent == nil {
		return e
	}
	return e.parent
}

func (e *Environ[T]) Resolve(ident string) (T, error) {
	var zero T
	v, ok := e.values[ident]
	if !ok {
		if e.parent != nil {
			return e.parent.Resolve(ident)
		}
		return zero, fmt.Errorf("%s undefined variable", ident)
	}
	return v, nil
}

func (e *Environ[T]) Define(ident string, values T) {
	e.values[ident] = values
}
