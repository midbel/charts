package dsl

import (
	"fmt"
)

type env struct {
	values map[string][]string
}

func emptyEnv() *env {
	return &env{
		values: make(map[string][]string),
	}
}

func (e *env) Resolve(name string) ([]string, error) {
	v, ok := e.values[name]
	if !ok {
		return nil, fmt.Errorf("%s undefined variable")
	}
	return v, nil
}

func (e *env) Define(name string, values []string) {
	e.values[name] = values
}
