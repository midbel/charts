package dsl

type Selector interface {
	Select([]string) (float64, error)
}