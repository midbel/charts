package dsl

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/charts"
)

var (
	DefaultWidth  = 800.0
	DefaultHeight = 600.0

	TimeFormat  = "%y-%m-%d"
	DefaultPath = "out.svg"
)

const (
	TypeNumber = "number"
	TypeTime   = "time"
	TypeString = "string"
)

const (
	RenderLine       = "line"
	RenderPie        = "pie"
	RenderBar        = "bar"
	RenderStep       = "step"
	RenderStepAfter  = "step-after"
	RenderStepBefore = "step-before"
)

type Config struct {
	Title  string
	Path   string
	Width  float64
	Height float64
	Pad    struct {
		Top    float64
		Right  float64
		Bottom float64
		Left   float64
	}
	TimeFormat string
	Types      struct {
		X string
		Y string
	}
	Domains struct {
		X Domain
		Y Domain
	}
	Center struct {
		X string
		Y string
	}
	Files []File

	Style Style
}

func Default() Config {
	cfg := Config{
		Path:       DefaultPath,
		Width:      DefaultWidth,
		Height:     DefaultHeight,
		TimeFormat: TimeFormat,
		Style:      GlobalStyle(),
	}
	cfg.Types.X = TypeNumber
	cfg.Types.Y = TypeNumber

	return cfg
}

func (c Config) Render() error {
	var err error
	switch {
	case c.Types.X == TypeNumber && c.Types.Y == TypeNumber:
	case c.Types.X == TypeTime && c.Types.Y == TypeNumber:
	default:
		err = fmt.Errorf("unsupported chart type %s/%s", c.Types.X, c.Types.Y)
	}
	return err
}

type Domain struct {
	Label      string
	Ticks      int
	Format     string
	Domain     []string
	Position   string
	InnerTicks bool
	OuterTicks bool
	LabelTicks bool
	BandTicks  bool
}

type Style struct {
	Type          string
	Stroke        string
	Fill          bool
	Point         string
	InnerRadius   float64
	OuterRadius   float64
	IgnoreMissing bool
	TextPosition  string
}

func GlobalStyle() Style {
	return Style{
		Type:   RenderLine,
		Stroke: "black",
		Fill:   false,
	}
}

type File struct {
	Path  string
	Ident string
	X     int
	Y     int
	Style
}

func (f File) Name() string {
	if f.Ident != "" {
		return f.Ident
	}
	return strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
}

func (f File) TimeSerie() (charts.Data, error) {
	return nil, nil
}

type PointFunc[T, U charts.ScalerConstraint] func([]string) (charts.Point[T, U], error)

func loadPoints[T, U charts.ScalerConstraint](file string, get PointFunc[T, U]) ([]charts.Point[T, U], error) {
	r, err := os.Open(file)
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
