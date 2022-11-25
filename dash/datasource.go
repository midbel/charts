package dash

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/charts"
	"github.com/midbel/shlex"
	"github.com/midbel/slices"
)

type Element struct {
	Type  string
	Ident string
	Data  DataSource
	Using
	Style any // one of NumberStyle, CategoryStyle, CircularStyle
}

func (e Element) TimeSerie(timefmt string, x TimeScale, y FloatScale) (charts.Data, error) {
	ser, err := e.resetSource().TimeSerie(timefmt, x, y)
	if err != nil {
		return nil, err
	}
	ser.Renderer, err = getRenderer[time.Time, float64](e.Type, e.Style)
	return ser, err
}

func (e Element) NumberSerie(x FloatScale, y FloatScale) (charts.Data, error) {
	ser, err := e.resetSource().NumberSerie(x, y)
	if err != nil {
		return nil, err
	}
	ser.Renderer, err = getRenderer[float64, float64](e.Type, e.Style)
	return ser, err
}

func (e Element) CategorySerie(x StringScale, y FloatScale) (charts.Data, error) {
	ser, err := e.resetSource().CategorySerie(x, y)
	if err != nil {
		return nil, err
	}
	ser.Renderer, err = getCategoryRenderer[string, float64](e.Type, e.Style)
	return ser, err
}

func (e Element) resetSource() DataSource {
	if !e.Using.valid() {
		return e.Data
	}
	switch d := e.Data.(type) {
	case HttpFile:
		d.Using = e.Using
		return d
	case LocalFile:
		d.Using = e.Using
		return d
	default:
		return e.Data
	}
}

type (
	TimeSerie     = charts.Serie[time.Time, float64]
	NumberSerie   = charts.Serie[float64, float64]
	CategorySerie = charts.Serie[string, float64]

	TimeScale   = charts.Scaler[time.Time]
	FloatScale  = charts.Scaler[float64]
	StringScale = charts.Scaler[string]
)

type DataSource interface {
	TimeSerie(string, TimeScale, FloatScale) (TimeSerie, error)
	NumberSerie(FloatScale, FloatScale) (NumberSerie, error)
	CategorySerie(StringScale, FloatScale) (CategorySerie, error)
}

type Limit struct {
	Offset int
	Count  int
}

func (i Limit) zero() bool {
	return i.Offset == 0 && i.Count == 0
}

type Using struct {
	X int
	Y Selector
}

func (u Using) valid() bool {
	return u.Y != nil
}

type Exec struct {
	Ident   string
	Command string
}

func (e Exec) TimeSerie(timefmt string, x TimeScale, y FloatScale) (ser TimeSerie, err error) {
	out, err := e.execute()
	if err != nil {
		return
	}
	get, err := getTimeFunc(0, SelectSingle(1), timefmt)
	if err != nil {
		return
	}
	points, err := loadPointsFromReader(strings.NewReader(out), get)
	if err != nil {
		return
	}
	ser = createSerie[time.Time, float64](e.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, err
}

func (e Exec) NumberSerie(x FloatScale, y FloatScale) (ser NumberSerie, err error) {
	out, err := e.execute()
	if err != nil {
		return
	}

	get := getNumberFunc(0, SelectSingle(1))
	points, err := loadPointsFromReader(strings.NewReader(out), get)
	if err != nil {
		return
	}
	ser = createSerie[float64, float64](e.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, err
}

func (e Exec) CategorySerie(x StringScale, y FloatScale) (ser CategorySerie, err error) {
	out, err := e.execute()
	if err != nil {
		return
	}

	get := getCategoryFunc(0, SelectSingle(1))
	points, err := loadPointsFromReader(strings.NewReader(out), get)
	if err != nil {
		return
	}
	ser = createSerie[string, float64](e.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

func (e Exec) execute() (string, error) {
	words, err := shlex.Split(strings.NewReader(e.Command))
	if err != nil {
		return "", err
	}
	cmd := exec.Command(slices.Fst(words), slices.Rest(words)...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

type Expr struct {
	Ident string
	Expr  ast.Expression
}

func (e Expr) TimeSerie(timefmt string, x TimeScale, y FloatScale) (ser TimeSerie, err error) {
	return
}

func (e Expr) NumberSerie(x FloatScale, y FloatScale) (ser NumberSerie, err error) {
	return
}

func (e Expr) CategorySerie(x StringScale, y FloatScale) (ser CategorySerie, err error) {
	return
}

type HttpFile struct {
	Uri   string
	Ident string
	Using
	Limit

	Method string
	Body   string

	Username string
	Password string
	Token    string

	Headers http.Header
}

func (f HttpFile) TimeSerie(timefmt string, x TimeScale, y FloatScale) (ser TimeSerie, err error) {
	if !f.Using.valid() {
		return ser, fmt.Errorf("invalid column selector given")
	}
	r, err := f.execute()
	if err != nil {
		return
	}
	defer r.Close()

	get, err := getTimeFunc(f.X, f.Y, timefmt)
	if err != nil {
		return
	}
	points, err := loadPointsFromReader(r, get)
	if err != nil {
		return
	}
	points = splitPoints(f.Limit, points)

	ser = createSerie[time.Time, float64](f.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, err
}

func (f HttpFile) NumberSerie(x FloatScale, y FloatScale) (ser NumberSerie, err error) {
	if !f.Using.valid() {
		return ser, fmt.Errorf("invalid column selector given")
	}
	r, err := f.execute()
	if err != nil {
		return
	}
	defer r.Close()

	get := getNumberFunc(f.X, f.Y)
	points, err := loadPointsFromReader(r, get)
	if err != nil {
		return
	}
	points = splitPoints(f.Limit, points)

	ser = createSerie[float64, float64](f.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, err
}

func (f HttpFile) CategorySerie(x StringScale, y FloatScale) (ser CategorySerie, err error) {
	if !f.Using.valid() {
		return ser, fmt.Errorf("invalid column selector given")
	}
	r, err := f.execute()
	if err != nil {
		return
	}
	defer r.Close()

	get := getCategoryFunc(f.X, f.Y)
	points, err := loadPointsFromReader(r, get)
	if err != nil {
		return
	}
	points = splitPoints(f.Limit, points)

	ser = createSerie[string, float64](f.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

func (f HttpFile) execute() (io.ReadCloser, error) {
	req, err := http.NewRequest(f.Method, f.Uri, strings.NewReader(f.Body))
	if err != nil {
		return nil, err
	}
	req.Header = f.Headers.Clone()
	if set := req.Header.Get("Authorization"); f.Token != "" && len(set) == 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.Token))
	}
	if f.Token == "" && f.Username != "" && f.Password != "" {
		req.SetBasicAuth(f.Username, f.Password)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("%d: %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	return res.Body, nil
}

type LocalData struct {
	Ident   string
	Content string
}

func (d LocalData) TimeSerie(timefmt string, x TimeScale, y FloatScale) (ser TimeSerie, err error) {
	get, err := getTimeFunc(0, SelectSingle(1), timefmt)
	if err != nil {
		return
	}
	points, err := loadPointsFromReader(strings.NewReader(d.Content), get)
	if err != nil {
		return
	}
	ser = createSerie[time.Time, float64](d.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, err
}

func (d LocalData) NumberSerie(x FloatScale, y FloatScale) (ser NumberSerie, err error) {
	get := getNumberFunc(0, SelectSingle(1))
	points, err := loadPointsFromReader(strings.NewReader(d.Content), get)
	if err != nil {
		return
	}
	ser = createSerie[float64, float64](d.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

func (d LocalData) CategorySerie(x StringScale, y FloatScale) (ser CategorySerie, err error) {
	get := getCategoryFunc(0, SelectSingle(1))
	points, err := loadPointsFromReader(strings.NewReader(d.Content), get)
	if err != nil {
		return
	}
	ser = createSerie[string, float64](d.Ident, points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

type LocalFile struct {
	Path  string
	Ident string
	Using
	Limit
}

func (f LocalFile) Name() string {
	if f.Ident != "" {
		return f.Ident
	}
	return strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
}

func (f LocalFile) TimeSerie(timefmt string, x TimeScale, y FloatScale) (ser TimeSerie, err error) {
	if !f.Using.valid() {
		return ser, fmt.Errorf("invalid column selector given")
	}
	r, err := f.open()
	if err != nil {
		return
	}
	defer r.Close()

	get, err := getTimeFunc(f.X, f.Y, timefmt)
	if err != nil {
		return
	}
	points, err := loadPointsFromReader(r, get)
	if err != nil {
		return
	}
	points = splitPoints(f.Limit, points)

	ser = createSerie[time.Time, float64](f.Name(), points)
	ser.X = x
	ser.Y = y
	return ser, err
}

func (f LocalFile) NumberSerie(x FloatScale, y FloatScale) (ser NumberSerie, err error) {
	if !f.Using.valid() {
		return ser, fmt.Errorf("invalid column selector given")
	}
	r, err := f.open()
	if err != nil {
		return
	}
	defer r.Close()

	get := getNumberFunc(f.X, f.Y)
	points, err := loadPointsFromReader(r, get)
	if err != nil {
		return
	}
	points = splitPoints(f.Limit, points)

	ser = createSerie[float64, float64](f.Name(), points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

func (f LocalFile) CategorySerie(x StringScale, y FloatScale) (ser CategorySerie, err error) {
	if !f.Using.valid() {
		return ser, fmt.Errorf("invalid column selector given")
	}
	r, err := f.open()
	if err != nil {
		return
	}
	defer r.Close()

	get := getCategoryFunc(f.X, f.Y)
	points, err := loadPointsFromReader(r, get)
	if err != nil {
		return
	}
	points = splitPoints(f.Limit, points)

	ser = createSerie[string, float64](f.Name(), points)
	ser.X = x
	ser.Y = y
	return ser, nil
}

func (f LocalFile) open() (io.ReadCloser, error) {
	return os.Open(f.Path)
}

func splitPoints[T, U charts.ScalerConstraint](lim Limit, points []charts.Point[T, U]) []charts.Point[T, U] {
	if lim.zero() || lim.Offset < 0 || lim.Count < 0 {
		return points
	}
	if lim.Offset < len(points) {
		points = points[lim.Offset:]
	}
	if lim.Count < len(points) {
		points = points[:lim.Count]
	}
	return points
}

type getFunc[T, U charts.ScalerConstraint] func([]string) (charts.Point[T, U], error)

func loadPointsFromReader[T, U charts.ScalerConstraint](r io.Reader, get getFunc[T, U]) ([]charts.Point[T, U], error) {
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

func getCategoryFunc(x int, y Selector) getFunc[string, float64] {
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

func getTimeFunc(x int, y Selector, timefmt string) (getFunc[time.Time, float64], error) {
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

func getNumberFunc(x int, y Selector) getFunc[float64, float64] {
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

func createSerie[T, U charts.ScalerConstraint](ident string, points []charts.Point[T, U]) charts.Serie[T, U] {
	return charts.Serie[T, U]{
		Title:  ident,
		Points: points,
	}
}
