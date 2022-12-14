package dash

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/midbel/charts"
	"github.com/midbel/slices"
)

type ScalerMaker interface {
	TimeScale(charts.Range, string, bool) (charts.Scaler[time.Time], error)
	NumberScale(charts.Range, bool) (charts.Scaler[float64], error)
	CategoryScale(charts.Range) (charts.Scaler[string], error)
}

var errValues = errors.New("not enough values given for domain")

type listScaler struct {
	values []string
}

func ScaleFromList(vs []string) ScalerMaker {
	return listScaler{
		values: vs,
	}
}

func (s listScaler) TimeScale(rg charts.Range, format string, reverse bool) (charts.Scaler[time.Time], error) {
	if len(s.values) < 2 {
		return nil, errValues
	}
	parseTime, err := makeParseTime(format)
	if err != nil {
		return nil, err
	}
	fst, err := parseTime(slices.Fst(s.values))
	if err != nil {
		return nil, err
	}
	lst, err := parseTime(slices.Lst(s.values))
	if err != nil {
		return nil, err
	}
	if reverse {
		fst, lst = lst, fst
	}
	return charts.TimeScaler(charts.TimeDomain(fst, lst), rg), nil
}

func (s listScaler) NumberScale(rg charts.Range, reverse bool) (charts.Scaler[float64], error) {
	if len(s.values) < 2 {
		return nil, errValues
	}
	fst, err := strconv.ParseFloat(slices.Fst(s.values), 64)
	if err != nil {
		return nil, err
	}
	lst, err := strconv.ParseFloat(slices.Lst(s.values), 64)
	if err != nil {
		return nil, err
	}
	if reverse {
		fst, lst = lst, fst
	}
	return charts.NumberScaler(charts.NumberDomain(fst, lst), rg), nil
}

func (s listScaler) CategoryScale(rg charts.Range) (charts.Scaler[string], error) {
	return charts.StringScaler(s.values, rg), nil
}

type fileScaler struct {
	path string
	Indexer
}

func ScaleFromFile(path string, ix Indexer) ScalerMaker {
	return fileScaler{
		path:    path,
		Indexer: ix,
	}
}

func (s fileScaler) TimeScale(rg charts.Range, format string, reverse bool) (charts.Scaler[time.Time], error) {
	parseTime, err := makeParseTime(format)
	if err != nil {
		return nil, err
	}
	cols := s.columns()
	if len(cols) != 1 {
		return nil, fmt.Errorf("invalid number of column given")
	}
	var (
		fd time.Time
		td time.Time
		ix = slices.Fst(cols)
	)
	err = s.readFile(func(row []string) error {
		if ix < 0 || ix >= len(row) {
			return ErrIndex
		}
		when, err := parseTime(row[ix])
		if err != nil {
			return err
		}
		if fd.IsZero() || when.Before(fd) {
			fd = when
		}
		if td.IsZero() || when.After(td) {
			td = when
		}
		return nil
	})
	return charts.TimeScaler(charts.TimeDomain(fd, td), rg), err
}

func (s fileScaler) NumberScale(rg charts.Range, reverse bool) (charts.Scaler[float64], error) {
	sel, ok := s.Indexer.(Selector)
	if !ok {
		return nil, fmt.Errorf("invalid selection string")
	}
	var (
		min float64
		max float64
	)
	err := s.readFile(func(row []string) error {
		vs, err := sel.Select(row)
		if err != nil {
			return err
		}
		if len(vs) != 1 {
			return fmt.Errorf("too many values retrieved from selector")
		}
		min = math.Min(min, slices.Fst(vs))
		max = math.Max(max, slices.Fst(vs))
		return nil
	})
	if reverse {
		min, max = max, min
	}
	return charts.NumberScaler(charts.NumberDomain(min, max), rg), err
}

func (s fileScaler) CategoryScale(rg charts.Range) (charts.Scaler[string], error) {
	cols := s.columns()
	if len(cols) != 1 {
		return nil, fmt.Errorf("invalid number of column given")
	}
	var (
		seen  = make(map[string]struct{})
		empty = struct{}{}
		list  []string
		ix    = slices.Fst(cols)
	)
	err := s.readFile(func(row []string) error {
		if ix < 0 || ix >= len(row) {
			return fmt.Errorf("invalid index")
		}
		_, ok := seen[row[ix]]
		if !ok {
			list = append(list, row[ix])
			seen[row[ix]] = empty
		}
		return nil
	})
	return charts.StringScaler(list, rg), err
}

func (s fileScaler) readFile(read func(row []string) error) error {
	r, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer r.Close()

	rs := csv.NewReader(r)
	rs.Read()
	for {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if err := read(row); err != nil {
			return err
		}
	}
	return nil
}
