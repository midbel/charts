package dash

import (
	"errors"
	"strconv"
)

var ErrIndex = errors.New("invalid index")

type Indexer interface {
	columns() []int
}

type Selector interface {
	Select([]string) ([]float64, error)
	Indexer
}

type combined struct {
	selectors []Selector
}

func Combined(xs ...Selector) Selector {
	return combined{
		selectors: xs,
	}
}

func (c combined) columns() []int {
	var list []int
	for _, s := range c.selectors {
		x, ok := s.(Indexer)
		if !ok {
			continue
		}
		list = append(list, x.columns()...)
	}
	return list
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

func SelectSum(list []int) Selector {
	return summer{
		index: list,
	}
}

func (s summer) columns() []int {
	return s.index
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

func SelectSingle(i int) Selector {
	return SelectMulti([]int{i})
}

func SelectMulti(list []int) Selector {
	return multi{
		index: list,
	}
}

func (m multi) columns() []int {
	return m.index
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

func ExpandRange(fst, lst int) []int {
	var list []int
	for i := fst; i <= lst; i++ {
		list = append(list, i)
	}
	return list
}
