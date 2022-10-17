package dsl

import (
	"errors"
	"strconv"
)

var ErrIndex = errors.New("invalid index")

type Selector interface {
	Select([]string) ([]float64, error)
}

type combined struct {
	selectors []Selector
}

func (c combined) Select(row []string) ([]float64, error) {
	var list []float64
	for _, s := range c.selectors {
		fs, err := s.Select(row)
		if err != nil {
			return nil, err
		}
		list = append(list, fs...)
	}
	return list, nil
}

type summer struct {
	index []int
}

func selectSum(list []int) Selector {
	return summer{
		index: list,
	}
}

func (s summer) Select(row []string) ([]float64, error) {
	var sum float64
	for _, i := range s.index {
		if i < 0 || i >= len(row) {
			return nil, ErrIndex
		}
		f, err := strconv.ParseFloat(row[i], 64)
		if err != nil {
			return nil, err
		}
		sum += f
	}
	return []float64{sum}, nil
}

type multi struct {
	index []int
}

func selectMulti(list []int) Selector {
	return multi{
		index: list,
	}
}

func (m multi) Select(row []string) ([]float64, error) {
	list := make([]float64, 0, len(m.index))
	for _, i := range m.index {
		if i < 0 || i >= len(row) {
			return nil, ErrIndex
		}
		f, err := strconv.ParseFloat(row[i], 64)
		if err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, nil
}

func expandRange(fst, lst int) []int {
	var list []int
	for i := fst; i <= lst; i++ {
		list = append(list, i)
	}
	return list
}
