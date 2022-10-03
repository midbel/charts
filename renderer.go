package charts

import (
	"math"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

type TextPosition int

const (
	TextBefore TextPosition = 1 << iota
	TextAfter
)

const currentColour = "currentColour"

type Renderer[T, U ScalerConstraint] interface {
	Render(Serie[T, U]) svg.Element
}

// type SunburstRenderer struct[T ~string, U ~float64] struct {
// 	Fill       []string
// 	InnerRadius float64
// 	OuterRadius float64
// }

// func (r SunburstRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
// 	return nil
// }

// type PieRenderer struct[T ~string, U ~float64] struct {
// 	Fill       []string
// 	InnerRadius float64
// 	OuterRadius float64
// }

// func (r PieRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
// 	return nil
// }

type StackedRenderer[T ~string, U ~float64] struct {
	Fill       []string
	Width      float64
	Horizontal bool
	WithText   bool
	WithValue  bool
}

func (r StackedRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.Width <= 0 {
		r.Width = 1
	}
	if r.Horizontal {
		return r.renderHorizontal(serie)
	}
	return r.renderVertical(serie)
}

func (r StackedRenderer[T, U]) renderHorizontal(serie Serie[T, U]) svg.Element {
	var grp svg.Group
	for _, s := range serie.Series {
		bar := getBaseGroup("", "bar")
		bar.Transform = svg.Translate(serie.X.Scale(any(s.Title).(T)), 0)
		for i, pt := range s.Points {
			var (
				w   = s.X.Space() * r.Width
				x   = float64(i) * s.X.Space()
				y   = s.Y.Scale(pt.Y)
				pos = svg.NewPos(x, y)
				dim = svg.NewDim(w, s.Y.Max()-y)
			)
			var el svg.Rect
			el.Title = any(pt.X).(string)
			el.Pos = pos
			el.Dim = dim
			el.Fill = svg.NewFill(r.Fill[i%len(r.Fill)])
			bar.Append(el.AsElement())
		}
		grp.Append(bar.AsElement())
	}
	return grp.AsElement()
}

func (r StackedRenderer[T, U]) renderVertical(serie Serie[T, U]) svg.Element {
	var grp svg.Group
	for _, s := range serie.Series {
		var (
			total  float64
			height = serie.Y.Max()
			bar    = getBaseGroup("", "bar")
		)
		bar.Transform = svg.Translate(serie.X.Scale(any(s.Title).(T)), 0)
		for i, pt := range s.Points {
			total += any(pt.Y).(float64)
			var (
				y  = serie.Y.Scale(any(total).(U))
				w  = serie.X.Space() * r.Width
				o  = (serie.X.Space() - w) / 2
				el svg.Rect
			)
			el.Title = any(pt.X).(string)
			el.Pos = svg.NewPos(o, y)
			el.Dim = svg.NewDim(serie.X.Space()*r.Width, height-y)
			el.Fill = svg.NewFill(r.Fill[i%len(r.Fill)])
			bar.Append(el.AsElement())
			if r.WithText {

			}
			if r.WithValue {

			}
			height = y
		}
		grp.Append(bar.AsElement())
	}
	return grp.AsElement()
}

type BarRenderer[T ~string, U ~float64] struct {
	Fill      []string
	Width     float64
	WithValue bool
}

func (r BarRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.Width <= 0 {
		r.Width = 1
	}
	grp := getBaseGroup("")
	for i, pt := range serie.Points {
		var (
			w   = serie.X.Space() * r.Width
			o   = (serie.X.Space() - w) / 2
			x   = serie.X.Scale(pt.X) + o
			y   = serie.Y.Scale(pt.Y)
			pos = svg.NewPos(x, y)
			dim = svg.NewDim(w, serie.Y.Max()-y)
		)
		var el svg.Rect
		el.Pos = pos
		el.Dim = dim
		el.Fill = svg.NewFill(r.Fill[i%len(r.Fill)])
		grp.Append(el.AsElement())
	}
	return grp.AsElement()
}

type PointRenderer[T, U ScalerConstraint] struct {
	Color string
	Point PointFunc
}

func (r PointRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	grp := getBaseGroup(r.Color, "scatter")
	return grp.AsElement()
}

type CubicRenderer[T, U ScalerConstraint] struct {
	Stretch float64
	Color   string
	Fill    bool
	Point   PointFunc
}

func (r CubicRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line")
		pat = getBasePath(r.Fill)
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsMoveTo(pos)
	if r.Point != nil {
		grp.Append(r.Point(pos))
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		var (
			ctrl1 = ori
			ctrl2 = pos
			diff  = (pos.X - ori.X) * r.Stretch
		)
		ctrl1.X += diff
		ctrl2.X -= diff

		pat.AbsCubicCurve(pos, ctrl1, ctrl2)
		ori = pos
		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type LinearRenderer[T, U ScalerConstraint] struct {
	Fill  bool
	Color string
	Point PointFunc
	Text  TextPosition
}

func (r LinearRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line")
		pat = getBasePath(r.Fill)
		pos svg.Pos
		nan bool
	)
	for i, pt := range serie.Points {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)
		if i == 0 || (nan && serie.IgnoreMissing) {
			nan = false
			pat.AbsMoveTo(pos)
		} else {
			pat.AbsLineTo(pos)
		}
		if r.Point != nil {
			el := r.Point(pos)
			if el != nil {
				grp.Append(r.Point(pos))
			}
		}
	}

	switch r.Text {
	case TextBefore:
		pt := slices.Fst(serie.Points)
		txt := getLineText(serie.Title, 0, serie.Y.Scale(pt.Y), true)
		grp.Append(txt.AsElement())
	case TextAfter:
		pt := slices.Lst(serie.Points)
		txt := getLineText(serie.Title, serie.X.Scale(pt.X), serie.Y.Scale(pt.Y), false)
		grp.Append(txt.AsElement())
	default:
	}

	if r.Fill {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type StepRenderer[T, U ScalerConstraint] struct {
	Color string
	Fill  bool
	Point PointFunc
}

func (r StepRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line")
		pat = getBasePath(r.Fill)
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
		nan bool
	)
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pat.AbsLineTo(pos)
	if r.Point != nil {
		grp.Append(r.Point(pos))
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)
		if nan && serie.IgnoreMissing {
			nan = false
			pat.AbsMoveTo(pos)
		} else {
			ori.X += (pos.X - ori.X) / 2
			pat.AbsLineTo(ori)
			ori.Y = pos.Y
			pat.AbsLineTo(ori)
			pat.AbsLineTo(pos)
		}
		ori = pos
		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
	}
	if r.Fill {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type StepAfterRenderer[T, U ScalerConstraint] struct {
	Color string
	Fill  bool
	Point PointFunc
}

func (r StepAfterRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line")
		pat = getBasePath(r.Fill)
		pos svg.Pos
		ori svg.Pos
		nan bool
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Max()
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	if r.Point != nil {
		grp.Append(r.Point(pos))
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		if nan && serie.IgnoreMissing {
			nan = false
			pat.AbsMoveTo(pos)
		} else {
			ori.X = pos.X
			pat.AbsLineTo(ori)
			ori.Y = pos.Y
			pat.AbsLineTo(ori)
			pat.AbsLineTo(pos)
		}
		ori = pos

		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
	}
	if r.Fill {
		pos.X = serie.X.Max()
		pat.AbsLineTo(pos)
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type StepBeforeRenderer[T, U ScalerConstraint] struct {
	Color string
	Fill  bool
	Point PointFunc
}

func (r StepBeforeRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line")
		pat = getBasePath(r.Fill)
		pos svg.Pos
		ori svg.Pos
		nan bool
	)
	pos.X = serie.X.Min()
	pos.Y = serie.Y.Max()
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pat.AbsLineTo(pos)
	if r.Point != nil {
		grp.Append(r.Point(pos))
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		if nan && serie.IgnoreMissing {
			nan = false
			pat.AbsMoveTo(pos)
		} else {
			ori.Y = pos.Y
			pat.AbsLineTo(ori)
			ori.X = pos.X
			pat.AbsLineTo(ori)
			pat.AbsLineTo(pos)
		}
		ori = pos

		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
	}
	if r.Fill {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func getLineText(str string, x, y float64, before bool) svg.Text {
	txt := svg.NewText(str)
	txt.Font = svg.NewFont(FontSize)
	txt.Pos = svg.NewPos(x, y)
	txt.Anchor = "end"
	txt.Baseline = "middle"
	if !before {
		txt.Anchor = "start"
		txt.Pos.X += FontSize * 0.4
	} else {
		txt.Pos.X -= FontSize * 0.4
	}
	return txt
}

func getBasePath(fill bool) svg.Path {
	var pat svg.Path
	pat.Stroke = svg.NewStroke(currentColour, 1)
	if fill {
		pat.Fill = svg.NewFill(currentColour)
		pat.Fill.Opacity = 0.5
	} else {
		pat.Fill = svg.NewFill("none")
	}
	return pat
}

func getBaseGroup(color string, class ...string) svg.Group {
	var g svg.Group
	if color != "" {
		g.Fill = svg.NewFill(color)
		g.Stroke = svg.NewStroke(color, 1)
	}
	g.Class = class
	return g
}

func isFloat[T any](v T) (float64, bool) {
	x, ok := any(v).(float64)
	return x, ok
}
