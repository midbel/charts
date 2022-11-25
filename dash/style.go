package dash

import (
	"fmt"

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

type Style = charts.Style

type NumberStyle struct {
	Style
	Ident         string
	TextPosition  charts.TextPosition
	IgnoreMissing bool
}

func DefaultNumberStyle() NumberStyle {
	return NumberStyle{
		Style:         charts.DefaultStyle(),
		IgnoreMissing: true,
	}
}

type CategoryStyle struct {
	Style
	Ident string
	Width float64
}

func DefaultCategoryStyle() CategoryStyle {
	return CategoryStyle{
		Style: charts.DefaultStyle(),
		Width: 0.75,
	}
}

func (s CategoryStyle) Copy() CategoryStyle {
	x := s
	return x
}

type CircularStyle struct {
	Style
	Ident       string
	Fill        []string
	InnerRadius float64
	OuterRadius float64
}

func DefaultCircularStyle() CircularStyle {
	return CircularStyle{
		Style:       charts.DefaultStyle(),
		Fill:        charts.Tableau10,
		InnerRadius: 0,
		OuterRadius: 0,
	}
}

func (s CircularStyle) Copy() CircularStyle {
	x := s
	x.Fill = make([]string, len(s.Fill))
	copy(x.Fill, s.Fill)
	return x
}

func GetTextPosition(str string) charts.TextPosition {
	var pos charts.TextPosition
	switch str {
	case "text-before":
		pos = charts.TextBefore
	case "text-after":
		pos = charts.TextAfter
	default:
	}
	return pos
}

func GetPointFunc(str string) charts.PointFunc {
	switch str {
	case "circle":
		return charts.GetCircle
	case "square":
		return charts.GetSquare
	default:
		return nil
	}
}

func GetLineType(str string) charts.LineStyle {
	var i charts.LineStyle
	switch str {
	case "", StyleStraight, StyleSolid:
		i = charts.StyleStraight
	case StyleDotted:
		i = charts.StyleDotted
	case StyleDashed:
		i = charts.StyleDashed
	}
	return i
}

func getRenderer[T, U charts.ScalerConstraint](kind string, style any) (charts.Renderer[T, U], error) {
	var (
		rdr     charts.LinearRenderer[T, U]
		st, err = getNumberStyle(kind, style)
	)
	if err != nil {
		return nil, err
	}
	switch kind {
	case RenderLine:
		rdr = charts.Line[T, U]()
	case RenderStep:
		rdr = charts.Step[T, U]()
	case RenderStepBefore:
		rdr = charts.StepBefore[T, U]()
	case RenderStepAfter:
		rdr = charts.StepAfter[T, U]()
	default:
		return nil, fmt.Errorf("%s unrecognized chart type", kind)
	}
	rdr.Style = st.Style
	rdr.Text = st.TextPosition
	rdr.IgnoreMissing = st.IgnoreMissing
	return rdr, nil
}

func getCircularRenderer[T ~string, U float64](kind string, style any) (charts.Renderer[T, U], error) {
	var (
		rdr     charts.Renderer[T, U]
		st, err = getCircularStyle(kind, style)
	)
	if err != nil {
		return nil, err
	}
	switch kind {
	case RenderPie:
		rdr = charts.PieRenderer[T, U]{
			Style: st.Style,
			InnerRadius: st.InnerRadius,
			OuterRadius: st.OuterRadius,
		}
	case RenderSun:
		rdr = charts.SunburstRenderer[T, U]{
			Style: st.Style,
			InnerRadius: st.InnerRadius,
			OuterRadius: st.OuterRadius,
		}
	default:
		return nil, fmt.Errorf("%s unrecognized chart type", kind)
	}
	return rdr, nil
}

func getCategoryRenderer[T ~string, U float64](kind string, style any) (charts.Renderer[T, U], error) {
	if kind == RenderPie || kind == RenderSun {
		return getCircularRenderer[T, U](kind, style)
	}
	var (
		rdr     charts.Renderer[T, U]
		st, err = getCategoryStyle(kind, style)
	)
	if err != nil {
		return nil, err
	}
	switch kind {
	case RenderBar:
		rdr = charts.BarRenderer[T, U]{
			Style: st.Style,
			Width: st.Width,
		}
	case RenderGroup:
		rdr = charts.GroupRenderer[T, U]{
			Style: st.Style,
			Width: st.Width,
		}
	case RenderStack, RenderNormStack:
		rdr = charts.StackedRenderer[T, U]{
			Style:     st.Style,
			Width:     st.Width,
			Normalize: kind == RenderNormStack,
		}
	default:
		return nil, fmt.Errorf("%s unrecognized chart type", kind)
	}
	return rdr, nil
}

func getNumberStyle(kind string, style any) (NumberStyle, error) {
	st, ok := style.(NumberStyle)
	if !ok {
		return st, fmt.Errorf("invalid style given for %s renderer", kind)
	}
	return st, nil
}

func getCategoryStyle(kind string, style any) (CategoryStyle, error) {
	st, ok := style.(CategoryStyle)
	if !ok {
		return st, fmt.Errorf("invalid style given for %s renderer", kind)
	}
	return st, nil
}

func getCircularStyle(kind string, style any) (CircularStyle, error) {
	st, ok := style.(CircularStyle)
	if !ok {
		return st, fmt.Errorf("invalid style given for %s renderer", kind)
	}
	return st, nil
}
