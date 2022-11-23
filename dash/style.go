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
	Fill  []string
	Width float64
}

func DefaultCategoryStyle() CategoryStyle {
	return CategoryStyle{
		Style: charts.DefaultStyle(),
		Fill:  charts.Tableau10,
		Width: 0.75,
	}
}

func (s CategoryStyle) Copy() CategoryStyle {
	x := s
	x.Fill = make([]string, len(s.Fill))
	copy(x.Fill, s.Fill)
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
	var rdr charts.Renderer[T, U]
	switch kind {
	case RenderLine:
		st, err := getNumberStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.LinearRenderer[T, U]{
			Style:         st.Style,
			Text:          st.TextPosition,
			IgnoreMissing: st.IgnoreMissing,
		}
	case RenderStep:
		st, err := getNumberStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.StepRenderer[T, U]{
			Style:         st.Style,
			Text:          st.TextPosition,
			IgnoreMissing: st.IgnoreMissing,
		}
	case RenderStepBefore:
		st, err := getNumberStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.StepBeforeRenderer[T, U]{
			Style:         st.Style,
			Text:          st.TextPosition,
			IgnoreMissing: st.IgnoreMissing,
		}
	case RenderStepAfter:
		st, err := getNumberStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.StepAfterRenderer[T, U]{
			Style:         st.Style,
			Text:          st.TextPosition,
			IgnoreMissing: st.IgnoreMissing,
		}
	default:
		return nil, fmt.Errorf("%s unrecognized chart type", kind)
	}
	return rdr, nil
}

func getCategoryRenderer[T ~string, U float64](kind string, style any) (charts.Renderer[T, U], error) {
	var rdr charts.Renderer[T, U]
	switch kind {
	case RenderBar:
		st, err := getCategoryStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.BarRenderer[T, U]{
			Fill:  st.Fill,
			Width: st.Width,
		}
	case RenderGroup:
		st, err := getCategoryStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.GroupRenderer[T, U]{
			Fill:  st.Fill,
			Width: st.Width,
		}
	case RenderStack, RenderNormStack:
		st, err := getCategoryStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.StackedRenderer[T, U]{
			Fill:      st.Fill,
			Width:     st.Width,
			Normalize: kind == RenderNormStack,
		}
	case RenderPie:
		st, err := getCircularStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.PieRenderer[T, U]{
			Fill:        st.Fill,
			InnerRadius: st.InnerRadius,
			OuterRadius: st.OuterRadius,
		}
	case RenderSun:
		st, err := getCircularStyle(kind, style)
		if err != nil {
			return nil, err
		}
		rdr = charts.SunburstRenderer[T, U]{
			Fill:        st.Fill,
			InnerRadius: st.InnerRadius,
			OuterRadius: st.OuterRadius,
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
