package dsl

import (
	"fmt"
	"strconv"
	"time"

	"github.com/midbel/charts"
	"github.com/midbel/slices"
)

const (
	numberType = "number"
	timeType   = "time"
	stringType = "string"
)

const outFile = "out.svg"

const blackFill = "black"

const (
	defaultWidth  = 800
	defaultHeight = 600
)

type scale struct {
	Ident  string
	Type   string
	Range  []float64
	Domain []string
}

func defaultScale(ident string) scale {
	return scale{
		Ident: ident,
		Type:  numberType,
	}
}

func createScaler[T charts.ScalerConstraint](s scale) (charts.Scaler[T], error) {
	var (
		r = charts.NewRange(slices.Fst(s.Range), slices.Lst(s.Range))
		x charts.Scaler[T]
	)
	switch s.Type {
	default:
		return nil, validType(s.Type)
	case timeType:
		var (
			fd, err1 = time.Parse("2006-01-02", slices.Fst(s.Domain))
			td, err2 = time.Parse("2006-01-02", slices.Lst(s.Domain))
		)
		if err := hasError(err1, err2); err != nil {
			return nil, err
		}
		d := charts.TimeDomain(fd, td)
		x = charts.TimeScaler(d, r).(charts.Scaler[T])
	case numberType:
		var (
			f, err1 = strconv.ParseFloat(slices.Fst(s.Domain), 64)
			t, err2 = strconv.ParseFloat(slices.Fst(s.Domain), 64)
		)
		if err := hasError(err1, err2); err != nil {
			return nil, err
		}
		d := charts.NumberDomain(f, t)
		x = charts.NumberScaler(d, r).(charts.Scaler[T])
	case stringType:
		x = charts.StringScaler(s.Domain, r).(charts.Scaler[T])
	}
	return x, nil
}

type axis struct {
	Ident string
	Type  string
	Title string
	Ticks int
	Outer bool
	Inner bool
	Bands bool
	Label bool
	Color string
}

func defaultAxis(ident string) axis {
	return axis{
		Ident: ident,
		Title: ident,
		Type:  numberType,
		Ticks: 10,
		Inner: true,
		Label: true,
		Color: blackFill,
	}
}

func createAxis(a axis) (charts.Axis, error) {
	switch a.Type {
	default:
		return nil, validType(a.Type)
	case numberType:
	case timeType:
	case stringType:
	}
	return nil, nil
}

type chart struct {
	Ident  string
	Title  string
	Width  float64
	Height float64

	// Axis
	Left   interface{}
	Right  interface{}
	Bottom interface{}
	Top    interface{}
}

func defaultChart(ident string) chart {
	return chart{
		Ident:  ident,
		Title:  ident,
		Width:  defaultWidth,
		Height: defaultHeight,
	}
}

type lineRenderer struct {
	Ident         string
	Color         string
	Point         string
	IgnoreMissing bool
}

func defaultLineRenderer(ident string) lineRenderer {
	return lineRenderer{
		Ident:         ident,
		Color:         blackFill,
		IgnoreMissing: true,
	}
}

type pieRenderer struct {
	Ident  string
	Colors []string
	Inner  float64
	Outer  float64
}

func defaultPieRenderer(ident string) pieRenderer {
	return pieRenderer{
		Ident:  ident,
		Colors: []string{blackFill},
	}
}

type serie struct {
	Ident    string
	Title    string
	Values   interface{}
	Renderer lineRenderer
}

func defaultSerie(ident string) serie {
	return serie{
		Ident: ident,
		Title: ident,
	}
}

type file struct {
	Path string
}

func validType(str string) error {
	switch str {
	case numberType, timeType, stringType:
	default:
		return fmt.Errorf("%s: unknown type given")
	}
	return nil
}
