package dash

import (
	"time"

	"github.com/midbel/charts"
)

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
