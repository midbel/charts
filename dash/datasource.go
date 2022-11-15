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

type DataSource interface {
	TimeSerie(Style, string, charts.Scaler[time.Time], charts.Scaler[float64]) (charts.Data, error)
	NumberSerie(Style, string, charts.Scaler[float64], charts.Scaler[float64]) (charts.Data, error)
	CategorySerie(Style, string, charts.Scaler[string], charts.Scaler[float64]) (charts.Data, error)
}

type Data struct {
	Ident   string
	Content string
	Style
}

func (d Data) TimeSerie(g Style, timefmt string, x charts.Scaler[time.Time], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := d.makeTimeRenderer(g)
	if err != nil {
		return nil, err
	}
	get, err := getTimePointFunc(0, SelectSingle(1), timefmt)
	if err != nil {
		return nil, err
	}
	points, err := loadPointsFromReader(strings.NewReader(d.Content), get)
	if err != nil {
		return nil, err
	}
	ser := createSerie[time.Time, float64](d.Ident, rdr, points)
	ser.X = x
	ser.Y = y
	return ser, err
}

func (d Data) NumberSerie(g Style, x charts.Scaler[float64], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := d.makeNumberRenderer(g)
	if err != nil {
		return nil, err
	}
	get := getNumberPointFunc(0, SelectSingle(1))
	points, err := loadPointsFromReader(strings.NewReader(d.Content), get)
	if err != nil {
		return nil, err
	}
	ser := createSerie[float64, float64](d.Ident, rdr, points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

func (d Data) CategorySerie(g Style, x charts.Scaler[string], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := d.makeCategoryRenderer(g)
	if err != nil {
		return nil, err
	}
	get := getCategoryPointFunc(0, SelectSingle(1))
	points, err := loadPointsFromReader(strings.NewReader(d.Content), get)
	if err != nil {
		return nil, err
	}
	ser := createSerie[string, float64](d.Ident, rdr, points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

type Limit struct {
	Offset int
	Count  int
}

type File struct {
	Path  string
	Ident string
	X     int
	Y     Selector
	Limit
	Style
}

func (f File) Name() string {
	if f.Ident != "" {
		return f.Ident
	}
	return strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
}

func (f File) TimeSerie(g Style, timefmt string, x charts.Scaler[time.Time], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := f.makeTimeRenderer(g)
	if err != nil {
		return nil, err
	}
	points, err := loadTimePoints(f, timefmt)
	if err != nil {
		return nil, err
	}
	ser := createSerie[time.Time, float64](f.Name(), rdr, points)
	ser.X = x
	ser.Y = y
	return ser, err
}

func (f File) NumberSerie(g Style, x charts.Scaler[float64], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := f.makeNumberRenderer(g)
	if err != nil {
		return nil, err
	}
	points, err := loadNumberPoints(f)
	if err != nil {
		return nil, err
	}
	ser := createSerie[float64, float64](f.Name(), rdr, points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

func (f File) CategorySerie(g Style, x charts.Scaler[string], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := f.makeCategoryRenderer(g)
	if err != nil {
		return nil, err
	}
	points, err := loadCategoryPoints(f)
	if err != nil {
		return nil, err
	}
	ser := createSerie[string, float64](f.Name(), rdr, points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

func loadCategoryPoints(f File) ([]charts.Point[string, float64], error) {
	get := getCategoryPointFunc(f.X, f.Y)
	return loadPoints[string, float64](f.Path, f.Limit, get)
}

func loadNumberPoints(f File) ([]charts.Point[float64, float64], error) {
	get := getNumberPointFunc(f.X, f.Y)
	return loadPoints[float64, float64](f.Path, f.Limit, get)
}

func loadTimePoints(f File, timefmt string) ([]charts.Point[time.Time, float64], error) {
	get, err := getTimePointFunc(f.X, f.Y, timefmt)
	if err != nil {
		return nil, err
	}
	return loadPoints[time.Time, float64](f.Path, f.Limit, get)
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

func loadPoints[T, U charts.ScalerConstraint](file string, lim Limit, get pointFunc[T, U]) ([]charts.Point[T, U], error) {
	r, err := readFrom(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	list, err := loadPointsFromReader[T, U](r, get)
	if err != nil {
		return nil, err
	}

	z := len(list)
	if lim.Offset < 0 {
		lim.Offset = z + lim.Offset
	}
	if lim.Offset > 0 && lim.Offset < z {
		list = list[lim.Offset:]
	}
	if lim.Count > 0 && lim.Count < len(list) {
		list = list[:lim.Count]
	}
	return list, nil
}

func loadPointsFromReader[T, U charts.ScalerConstraint](r io.Reader, get pointFunc[T, U]) ([]charts.Point[T, U], error) {
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

func getCategoryPointFunc(x int, y Selector) pointFunc[string, float64] {
	get := func(row []string) (charts.Point[string, float64], error) {
		var (
			pt  charts.Point[string, float64]
			err error
		)
		pt.X = row[x]
		values, err := y.Select(row)
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
	return get
}

func getTimePointFunc(x int, y Selector, timefmt string) (pointFunc[time.Time, float64], error) {
	parseTime, err := makeParseTime(timefmt)
	if err != nil {
		return nil, err
	}
	get := func(row []string) (charts.Point[time.Time, float64], error) {
		var (
			pt  charts.Point[time.Time, float64]
			err error
		)
		if pt.X, err = parseTime(row[x]); err != nil {
			return pt, err
		}
		values, err := y.Select(row)
		if err != nil {
			return pt, err
		}
		pt.Y = slices.Fst(values)
		return pt, nil
	}
	return get, nil
}

func getNumberPointFunc(x int, y Selector) pointFunc[float64, float64] {
	get := func(row []string) (charts.Point[float64, float64], error) {
		var (
			pt  charts.Point[float64, float64]
			err error
		)
		if pt.X, err = strconv.ParseFloat(row[x], 64); err != nil {
			return pt, err
		}
		values, err := y.Select(row)
		if err != nil {
			return pt, err
		}
		pt.Y = slices.Fst(values)
		return pt, nil
	}
	return get
}

func createSerie[T, U charts.ScalerConstraint](ident string, rdr charts.Renderer[T, U], points []charts.Point[T, U]) charts.Serie[T, U] {
	return charts.Serie[T, U]{
		Title:    ident,
		Renderer: rdr,
		Points:   points,
	}
}
