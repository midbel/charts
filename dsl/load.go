package dsl

import (
	"github.com/midbel/charts"
)

type Loader interface {
	loadTimeSeries() ([]charts.Data, error)
	loadNumberSeries() ([]charts.Data, error)
	loadStringSeries() ([]charts.Data, error)
}

type pointFunc[T, U charts.ScalerConstraint] func([]string) (charts.Point[T, U], error)

type loader struct {
	X     int
	Y     Selector
	Index struct {
		Fst int
		Lst int
	}
	Style
}

type FileLoader struct {
	loader
	Path string
}

func (f FileLoader) loadTimeSeries() ([]charts.Data, error) {
	return nil, nil
}

func (f FileLoader) loadNumberSeries() ([]charts.Data, error) {
	return nil, nil
}

func (f FileLoader) loadStringSeries() ([]charts.Data, error) {
	return nil, nil
}

type CommandLoader struct {
	loader
	Command string
}

func (c CommandLoader) loadTimeSeries() ([]charts.Data, error) {
	return nil, nil
}

func (c CommandLoader) loadNumberSeries() ([]charts.Data, error) {
	return nil, nil
}

func (c CommandLoader) loadStringSeries() ([]charts.Data, error) {
	return nil, nil
}

type RemoteLoader struct {
	loader
	Addr    string
	Headers map[string][]string
}

func (h RemoteLoader) loadTimeSeries() ([]charts.Data, error) {
	return nil, nil
}

func (h RemoteLoader) loadNumberSeries() ([]charts.Data, error) {
	return nil, nil
}

func (h RemoteLoader) loadStringSeries() ([]charts.Data, error) {
	return nil, nil
}
