package dash

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/midbel/charts"
	"github.com/midbel/slices"
)

type Limit struct {
	Beg int
	End int
}

type File struct {
	Path       string
	Ident      string
	X          int
	Y          Selector
	TimeFormat string
	Limit
	Style
}

func (f File) Name() string {
	if f.Ident != "" {
		return f.Ident
	}
	return strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
}

func (f File) makeTimeSerie(g Style, timefmt string, x charts.Scaler[time.Time], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := f.makeTimeRenderer(g)
	if err != nil {
		return nil, err
	}
	if f.TimeFormat == "" {
		f.TimeFormat = timefmt
	}
	parseTime, err := makeParseTime(f.TimeFormat)
	if err != nil {
		return nil, err
	}

	points, err := loadTimePoints(f, parseTime)
	if err != nil {
		return nil, err
	}
	ser := charts.Serie[time.Time, float64]{
		Title:    f.Name(),
		Renderer: rdr,
		Points:   points,
		X:        x,
		Y:        y,
	}
	return ser, err
}

func (f File) makeNumberSerie(g Style, x charts.Scaler[float64], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := f.makeNumberRenderer(g)
	if err != nil {
		return nil, err
	}
	points, err := loadNumberPoints(f)
	if err != nil {
		return nil, err
	}
	ser := charts.Serie[float64, float64]{
		Title:    f.Name(),
		Renderer: rdr,
		Points:   points,
		X:        x,
		Y:        y,
	}
	return ser, nil
}

func (f File) makeCategorySerie(g Style, x charts.Scaler[string], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := f.makeCategoryRenderer(g)
	if err != nil {
		return nil, err
	}
	points, err := loadCategoryPoints(f)
	if err != nil {
		return nil, err
	}
	ser := charts.Serie[string, float64]{
		Title:    f.Name(),
		Renderer: rdr,
		Points:   points,
		X:        x,
		Y:        y,
	}
	return ser, nil
}

func loadCategoryPoints(f File) ([]charts.Point[string, float64], error) {
	get := func(row []string) (charts.Point[string, float64], error) {
		var (
			pt  charts.Point[string, float64]
			err error
		)
		pt.X = row[f.X]
		values, err := f.Y.Select(row)
		if err != nil {
			return pt, err
		}
		if len(values) == 1 {
			pt.Y = slices.Fst(values)
		} else {
			var total float64
			for i := range values {
				s := charts.CategoryPoint(fmt.Sprintf("%d", i), values[i])
				pt.Sub = append(pt.Sub, s)
				total += values[i]
			}
			pt.Y = total
		}
		return pt, nil
	}
	return loadPoints[string, float64](f.Path, get)
}

func loadNumberPoints(f File) ([]charts.Point[float64, float64], error) {
	get := func(row []string) (charts.Point[float64, float64], error) {
		var (
			pt  charts.Point[float64, float64]
			err error
		)
		pt.X, err = strconv.ParseFloat(row[f.X], 64)
		if err != nil {
			return pt, err
		}
		values, err := f.Y.Select(row)
		if err != nil {
			return pt, err
		}
		pt.Y = slices.Fst(values)
		return pt, nil
	}
	return loadPoints[float64, float64](f.Path, get)
}

func loadTimePoints(f File, parseTime func(string) (time.Time, error)) ([]charts.Point[time.Time, float64], error) {
	get := func(row []string) (charts.Point[time.Time, float64], error) {
		var (
			pt  charts.Point[time.Time, float64]
			err error
		)
		pt.X, err = parseTime(row[f.X])
		if err != nil {
			return pt, err
		}
		values, err := f.Y.Select(row)
		if err != nil {
			return pt, err
		}
		pt.Y = slices.Fst(values)
		return pt, nil
	}
	return loadPoints[time.Time, float64](f.Path, get)
}

func readFrom(location string) (io.ReadCloser, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http", "https":
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("request does not end with success result code")
		}
		return res.Body, nil
	case "", "file":
		return os.Open(u.Path)
	default:
		return nil, fmt.Errorf("%s: unsupported scheme", u.Scheme)
	}
}

type pointFunc[T, U charts.ScalerConstraint] func([]string) (charts.Point[T, U], error)

func loadPoints[T, U charts.ScalerConstraint](file string, get pointFunc[T, U]) ([]charts.Point[T, U], error) {
	r, err := readFrom(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var (
		rs   = csv.NewReader(r)
		list []charts.Point[T, U]
	)
	rs.Read()
	for {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		pt, err := get(row)
		if err != nil {
			return nil, err
		}
		list = append(list, pt)
	}
	return list, nil
}
