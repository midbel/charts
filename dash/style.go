package dash

import (
	"fmt"
	"time"

	"github.com/midbel/charts"
)

const (
	StyleStraight = "straight"
	StyleSolid    = "solid"
	StyleDotted   = "dotted"
	StyleDashed   = "dashed"
)

const (
	RenderLine       = "line"
	RenderStep       = "step"
	RenderStepAfter  = "step-after"
	RenderStepBefore = "step-before"
	RenderPie        = "pie"
	RenderBar        = "bar"
	RenderSun        = "sun"
	RenderStack      = "stack"
	RenderNormStack  = "stack-normalize"
	RenderGroup      = "group"
	RenderPolar      = "polar"
)

// type Style = charts.Style

type NumberStyle struct {
	charts.Style
	TextPosition  string
	LineType      string
	IgnoreMissing bool
	Color         string
}

type CategoryStyle struct {
	charts.Style
	Fill  []string
	Width float64
}

func (s CategoryStyle) Copy() CategoryStyle {
	x := s
	x.Fill = make([]string, len(s.Fill))
	copy(x.Fill, s.Fill)
	return x
}

type CircularStyle struct {
	charts.Style
	Fill        []string
	InnerRadius float64
	OuterRadius float64
}

func (s CircularStyle) Copy() CircularStyle {
	x := s
	x.Fill = make([]string, len(s.Fill))
	copy(x.Fill, s.Fill)
	return x
}

type Style struct {
	Type          string
	Stroke        string
	Fill          bool
	Point         string
	Width         float64
	InnerRadius   float64
	OuterRadius   float64
	IgnoreMissing bool
	TextPosition  string
	LineStyle     string
}

func GlobalStyle() Style {
	return Style{
		Type:   RenderLine,
		Stroke: "black",
		Fill:   false,
	}
}

func (s Style) getTextPosition() charts.TextPosition {
	var pos charts.TextPosition
	switch s.TextPosition {
	case "text-before":
		pos = charts.TextBefore
	case "text-after":
		pos = charts.TextAfter
	default:
	}
	return pos
}

func (s Style) getPointFunc() charts.PointFunc {
	switch s.Point {
	case "circle":
		return charts.GetCircle
	case "square":
		return charts.GetSquare
	default:
		return nil
	}
}

func (s Style) getLineStyle() charts.LineStyle {
	var i charts.LineStyle
	switch s.LineStyle {
	case "", StyleStraight:
		i = charts.StyleStraight
	case StyleDotted:
		i = charts.StyleDotted
	case StyleDashed:
		i = charts.StyleDashed
	}
	return i
}

func (s Style) makeTimeRenderer(g Style) (charts.Renderer[time.Time, float64], error) {
	return createRenderer[time.Time, float64](s.merge(g))
}

func (s Style) makeNumberRenderer(g Style) (charts.Renderer[float64, float64], error) {
	return createRenderer[float64, float64](s.merge(g))
}

func (s Style) makeCategoryRenderer(g Style) (charts.Renderer[string, float64], error) {
	return createCategoryRenderer(s.merge(g))
}

func (s Style) merge(g Style) Style {
	if s.Type == "" {
		s.Type = g.Type
	}
	if s.Stroke == "" {
		s.Stroke = g.Stroke
	}
	if s.Point == "" {
		s.Point = g.Point
	}
	if s.InnerRadius == 0 && g.InnerRadius != 0 {
		s.InnerRadius = g.InnerRadius
	}
	if s.OuterRadius == 0 && g.OuterRadius != 0 {
		s.OuterRadius = g.OuterRadius
	}
	if s.TextPosition == "" {
		s.TextPosition = g.TextPosition
	}
	return s
}

func createCategoryRenderer(style Style) (charts.Renderer[string, float64], error) {
	var rdr charts.Renderer[string, float64]
	switch style.Type {
	case RenderBar:
		rdr = charts.BarRenderer[string, float64]{
			Fill:  charts.Tableau10,
			Width: style.Width,
		}
	case RenderPie:
		rdr = charts.PieRenderer[string, float64]{
			Fill:        charts.Tableau10,
			InnerRadius: style.InnerRadius,
			OuterRadius: style.OuterRadius,
		}
	case RenderSun:
		rdr = charts.SunburstRenderer[string, float64]{
			Fill:        charts.Tableau10,
			InnerRadius: style.InnerRadius,
			OuterRadius: style.OuterRadius,
		}
	case RenderStack, RenderNormStack:
		rdr = charts.StackedRenderer[string, float64]{
			Fill:      charts.Tableau10,
			Width:     style.Width,
			Normalize: style.Type == RenderNormStack,
		}
	case RenderGroup:
		rdr = charts.GroupRenderer[string, float64]{
			Width: style.Width,
			Fill:  charts.Tableau10,
		}
	case RenderPolar:
		rdr = charts.PolarRenderer[string, float64]{
			Fill:       charts.Tableau10,
			Ticks:      10,
			TicksStyle: charts.StyleStraight,
		}
	default:
		return nil, fmt.Errorf("%s: can not use for number chart", style.Type)
	}
	return rdr, nil
}

func createRenderer[T, U charts.ScalerConstraint](style Style) (charts.Renderer[T, U], error) {
	var rdr charts.Renderer[T, U]
	switch style.Type {
	case RenderLine:
		rdr = charts.LinearRenderer[T, U]{
			Color:         style.Stroke,
			IgnoreMissing: style.IgnoreMissing,
			Text:          style.getTextPosition(),
			Point:         style.getPointFunc(),
			Style:         style.getLineStyle(),
		}
	case RenderStep:
		rdr = charts.StepRenderer[T, U]{
			Color:         style.Stroke,
			IgnoreMissing: style.IgnoreMissing,
			Text:          style.getTextPosition(),
			Point:         style.getPointFunc(),
			Style:         style.getLineStyle(),
		}
	case RenderStepAfter:
		rdr = charts.StepAfterRenderer[T, U]{
			Color:         style.Stroke,
			IgnoreMissing: style.IgnoreMissing,
			Text:          style.getTextPosition(),
			Point:         style.getPointFunc(),
			Style:         style.getLineStyle(),
		}
	case RenderStepBefore:
		rdr = charts.StepBeforeRenderer[T, U]{
			Color:         style.Stroke,
			IgnoreMissing: style.IgnoreMissing,
			Text:          style.getTextPosition(),
			Point:         style.getPointFunc(),
			Style:         style.getLineStyle(),
		}
	default:
		return nil, fmt.Errorf("%s: can be not used for number/time chart", style.Type)
	}
	return rdr, nil
}
