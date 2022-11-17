package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/midbel/charts"
	"github.com/midbel/slices"
)

const (
	defaultWidth  = 800
	defaultHeight = 600
	defaultTicks  = 7
)

var defaultPad = charts.Padding{
	Top:    80,
	Right:  80,
	Bottom: 80,
	Left:   80,
}

type Renderer interface {
	Render(io.Writer, ...charts.Data)
}

type scalerConstraint = charts.ScalerConstraint

func main() {
	var (
		title  = flag.String("title", "", "chart title")
		theme  = flag.String("theme", "", "theme file")
		kind   = flag.String("type", "", "chart type")
		xdata  = flag.String("xdata", "number", "x data type")
		ydata  = flag.String("ydata", "number", "y data type")
		xcol   = flag.Int("xcol", 0, "index of x column")
		ycol   = flag.Int("ycol", 1, "index of y column")
		xtics  = flag.Int("xtics", defaultTicks, "ticks on x axis")
		ytics  = flag.Int("ytics", defaultTicks, "ticks on y axis")
		xdom   = flag.String("xdom", "", "domain for x values")
		ydom   = flag.String("ydom", "", "domain for y values")
		width  = flag.Float64("width", defaultWidth, "chart width")
		height = flag.Float64("height", defaultHeight, "chart height")
		noaxis = flag.Bool("no-axis", false, "remove axis")
		result = flag.String("file", "", "output file")
	)
	flag.Parse()

	var (
		err error
		rdr Renderer
		get dataFunc
	)
	switch {
	case *xdata == "number" && *ydata == "number":
		ch, err0 := numberChart(*title, *theme, *width, *height)
		xscale, err1 := numberScale(*xdom, 0, *width-defaultPad.Horizontal(), false)
		yscale, err2 := numberScale(*ydom, 0, *height-defaultPad.Vertical(), true)
		if withAxis(*noaxis, *kind) {
			ch.Bottom = getAxis[float64](xscale, *xtics, charts.OrientBottom)
			ch.Left = getAxis[float64](yscale, *ytics, charts.OrientLeft)
		}

		get = numberSerie(*kind, *xcol, *ycol, xscale, yscale)

		rdr, err = ch, hasError(err0, err1, err2)
	case *xdata == "time" && *ydata == "number":
		ch, err0 := timeChart(*title, *theme, *width, *height)
		xscale, err1 := timeScale(*xdom, 0, *width-defaultPad.Horizontal(), false)
		yscale, err2 := numberScale(*ydom, 0, *height-defaultPad.Vertical(), true)
		if withAxis(*noaxis, *kind) {
			ch.Bottom = getAxis[time.Time](xscale, *xtics, charts.OrientBottom)
			ch.Left = getAxis[float64](yscale, *ytics, charts.OrientLeft)
		}

		get = timeSerie(*kind, *xcol, *ycol, xscale, yscale)

		rdr, err = ch, hasError(err0, err1, err2)
	case *xdata == "string" && *ydata == "number":
		ch, err0 := categoryChart(*title, *theme, *width, *height)
		xscale, err1 := stringScale(*xdom, 0, *width-defaultPad.Horizontal())
		yscale, err2 := numberScale(*ydom, 0, *height-defaultPad.Vertical(), true)
		if withAxis(*noaxis, *kind) {
			ch.Bottom = getAxis[string](xscale, *xtics, charts.OrientBottom)
			ch.Left = getAxis[float64](yscale, *ytics, charts.OrientLeft)
		}

		get = stringSerie(*kind, *xcol, *ycol, xscale, yscale)

		rdr, err = ch, hasError(err0, err1, err2)
	default:
		fmt.Fprintln(os.Stderr, "%s/%s: unsupported chart type", *xdata, *ydata)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating chart: %s", err)
		os.Exit(2)
	}
	var series []charts.Data
	for _, f := range flag.Args() {
		dat, err := get(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, "fail creating data from %s (%s): %s", f, *kind, err)
			os.Exit(2)
		}
		series = append(series, dat)
	}
	if err = renderChart(*result, rdr, series); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func hasError(es ...error) error {
	for i := range es {
		if es[i] != nil {
			return es[i]
		}
	}
	return nil
}

func getAxis[T scalerConstraint](scale charts.Scaler[T], ticks int, orient charts.Orientation) charts.Axis[T] {
	return charts.Axis[T]{
		Ticks:          ticks,
		WithOuterTicks: true,
		WithInnerTicks: true,
		WithLabelTicks: true,
		Scaler:         scale,
		Orientation:    orient,
	}
}

func numberRenderer[T, U scalerConstraint](kind string) (charts.Renderer[T, U], error) {
	var rdr charts.Renderer[T, U]
	switch kind {
	case "line", "":
		rdr = charts.LinearRenderer[T, U]{
			Color: "blue",
		}
	case "step":
		rdr = charts.StepRenderer[T, U]{
			Color: "blue",
		}
	case "step-before":
		rdr = charts.StepBeforeRenderer[T, U]{
			Color: "blue",
		}
	case "step-after":
		rdr = charts.StepAfterRenderer[T, U]{
			Color: "blue",
		}
	case "polar":
	default:
		return nil, fmt.Errorf("%s: unrecognized chart renderer", kind)
	}
	return rdr, nil
}

func categoryRenderer[T ~string, U ~float64](kind string, radius float64) (charts.Renderer[T, U], error) {
	var rdr charts.Renderer[T, U]
	switch kind {
	case "bar", "":
		rdr = charts.BarRenderer[T, U]{
			Width: 0.9,
		}
	case "group":
	case "stack":
	case "pie":
		rdr = charts.PieRenderer[T, U]{
			InnerRadius: radius / 2,
			OuterRadius: radius / 2,
		}
	case "donut":
		rdr = charts.PieRenderer[T, U]{
			InnerRadius: radius / 4,
			OuterRadius: radius / 2,
		}
	case "sun", "sunburst":
	case "polar":
	default:
		return nil, fmt.Errorf("%s: unrecognized chart renderer", kind)
	}
	return rdr, nil
}

func withAxis(no bool, kind string) bool {
	if !no {
		return !no
	}
	return kind != "pie" && kind != "polar" && kind != "sun" && kind != "sunburst" && kind != "donut"
}

func renderChart(file string, ch Renderer, series []charts.Data) error {
	var w io.Writer = os.Stdout
	if file != "" {
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	ch.Render(w, series...)
	return nil
}

func numberScale(str string, begin, end float64, reverse bool) (charts.Scaler[float64], error) {
	vs := strings.Split(str, ":")
	if len(vs) != 2 {
		return nil, fmt.Errorf("invalid number of values given for domain")
	}
	fn, err := strconv.ParseFloat(slices.Fst(vs), 64)
	if err != nil {
		return nil, err
	}
	tn, err := strconv.ParseFloat(slices.Lst(vs), 64)
	if err != nil {
		return nil, err
	}
	if reverse {
		fn, tn = tn, fn
	}
	return charts.NumberScaler(charts.NumberDomain(fn, tn), charts.NewRange(begin, end)), nil
}

func timeScale(str string, begin, end float64, reverse bool) (charts.Scaler[time.Time], error) {
	vs := strings.Split(str, ":")
	if len(vs) != 2 {
		return nil, fmt.Errorf("invalid number of values given for domain")
	}
	fd, err := time.Parse("2006-01-02", slices.Fst(vs))
	if err != nil {
		return nil, err
	}
	td, err := time.Parse("2006-01-02", slices.Lst(vs))
	if err != nil {
		return nil, err
	}
	if reverse {
		fd, td = td, fd
	}
	return charts.TimeScaler(charts.TimeDomain(fd, td), charts.NewRange(begin, end)), nil
}

func stringScale(str string, width, height float64) (charts.Scaler[string], error) {
	vs := strings.Split(str, ":")
	return charts.StringScaler(vs, charts.NewRange(width, height)), nil
}

func getIdent(file string) string {
	file = filepath.Base(file)
	for {
		e := filepath.Ext(file)
		if e == "" {
			break
		}
		file = strings.TrimSuffix(file, e)
	}
	return file
}

type dataFunc func(string) (charts.Data, error)

func numberSerie(kind string, xcol, ycol int, x, y charts.Scaler[float64]) dataFunc {
	get := func(file string) (charts.Data, error) {
		points, err := readPoints(file, xcol, ycol, getNumberNumber)
		if err != nil {
			return nil, err
		}
		rdr, err := numberRenderer[float64, float64](kind)
		if err != nil {
			return nil, err
		}
		ser := charts.Serie[float64, float64]{
			Title:    getIdent(file),
			Points:   points,
			Renderer: rdr,
			X:        x,
			Y:        y,
		}
		return ser, nil
	}
	return get
}

func timeSerie(kind string, xcol, ycol int, x charts.Scaler[time.Time], y charts.Scaler[float64]) dataFunc {
	get := func(file string) (charts.Data, error) {
		points, err := readPoints(file, xcol, ycol, getTimeNumber)
		if err != nil {
			return nil, err
		}
		rdr, err := numberRenderer[time.Time, float64](kind)
		if err != nil {
			return nil, err
		}
		ser := charts.Serie[time.Time, float64]{
			Title:    getIdent(file),
			Points:   points,
			Renderer: rdr,
			X:        x,
			Y:        y,
		}
		return ser, nil
	}
	return get
}

func stringSerie(kind string, xcol, ycol int, x charts.Scaler[string], y charts.Scaler[float64]) dataFunc {
	get := func(file string) (charts.Data, error) {
		points, err := readPoints(file, xcol, ycol, getStringNumber)
		if err != nil {
			return nil, err
		}
		radius := x.Max()
		if y.Max() < radius {
			radius = y.Max()
		}
		rdr, err := categoryRenderer[string, float64](kind, radius)
		if err != nil {
			return nil, err
		}
		ser := charts.Serie[string, float64]{
			Title:    getIdent(file),
			Points:   points,
			Renderer: rdr,
			X:        x,
			Y:        y,
		}
		return ser, nil
	}
	return get
}

type getFunc[T, U scalerConstraint] func(x, y string) (charts.Point[T, U], error)

func readPoints[T, U scalerConstraint](file string, x, y int, get getFunc[T, U]) ([]charts.Point[T, U], error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var (
		rs     = csv.NewReader(r)
		points []charts.Point[T, U]
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
		if x >= len(row) || x < 0 || y >= len(row) || y <= 0 {
			return nil, fmt.Errorf("invalid x/y index columns given")
		}
		pt, err := get(row[x], row[y])
		if err != nil {
			return nil, err
		}
		points = append(points, pt)
	}
	return points, nil
}

func getStringNumber(x, y string) (charts.Point[string, float64], error) {
	var (
		pt  charts.Point[string, float64]
		err error
	)
	pt.X = x
	pt.Y, err = strconv.ParseFloat(y, 64)
	return pt, err
}

func getTimeNumber(x, y string) (charts.Point[time.Time, float64], error) {
	var (
		pt  charts.Point[time.Time, float64]
		err error
	)
	pt.X, err = time.Parse("2006-01-02", x)
	if err != nil {
		return pt, err
	}
	pt.Y, err = strconv.ParseFloat(y, 64)
	return pt, err
}

func getNumberNumber(x, y string) (charts.Point[float64, float64], error) {
	var (
		pt  charts.Point[float64, float64]
		err error
	)
	pt.X, err = strconv.ParseFloat(x, 64)
	if err != nil {
		return pt, err
	}
	pt.Y, err = strconv.ParseFloat(y, 64)
	return pt, err
}

func numberChart(title, theme string, width, height float64) (charts.Chart[float64, float64], error) {
	ch := createChart[float64, float64](title, theme, width, height)
	return ch, nil

}

func timeChart(title, theme string, width, height float64) (charts.Chart[time.Time, float64], error) {
	ch := createChart[time.Time, float64](title, theme, width, height)
	return ch, nil
}

func categoryChart(title, theme string, width, height float64) (charts.Chart[string, float64], error) {
	ch := createChart[string, float64](title, theme, width, height)
	return ch, nil
}

func createChart[T, U scalerConstraint](title, theme string, width, height float64) charts.Chart[T, U] {
	var style string
	if dat, err := os.ReadFile(theme); err == nil {
		style = string(dat)
	}
	return charts.Chart[T, U]{
		Title:   title,
		Width:   width,
		Height:  height,
		Padding: defaultPad,
		Theme: style,
	}
}
