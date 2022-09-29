package charts

import (
	"math"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

type Renderer[T, U ScalerConstraint] interface {
	Render(Serie[T, U]) svg.Element
}

type RenderFunc[T, U ScalerConstraint] func(Serie[T, U]) svg.Element

func GetCirclePoint(pos svg.Pos) svg.Circle {
	ci := svg.NewCircle()
	ci.Radius = 2.5
	ci.Pos = pos
	ci.Class = append(ci.Class, "point")
	return ci
}

type cubicRenderer[T, U ScalerConstraint] struct {
	stretch       float64
	ignoreMissing bool
}

func CubicRender[T, U ScalerConstraint](stretch float64, ignoreMissing bool) Renderer[T, U] {
	return cubicRenderer[T, U]{
		stretch:       stretch,
		ignoreMissing: ignoreMissing,
	}
}

func (r cubicRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsMoveTo(pos)
	if serie.WithPoint != nil {
		grp.Append(serie.WithPoint(pos))
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		var (
			ctrl1 = ori
			ctrl2 = pos
			diff  = (pos.X - ori.X) * r.stretch
		)
		ctrl1.X += diff
		ctrl2.X -= diff

		pat.AbsCubicCurve(pos, ctrl1, ctrl2)
		ori = pos
		if serie.WithPoint != nil {
			grp.Append(serie.WithPoint(pos))
		}
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type quadraticRenderer[T, U ScalerConstraint] struct {
	stretch       float64
	ignoreMissing bool
}

func QuadraticRender[T, U ScalerConstraint](stretch float64, ignoreMissing bool) Renderer[T, U] {
	return quadraticRenderer[T, U]{
		stretch:       stretch,
		ignoreMissing: ignoreMissing,
	}
}

func (r quadraticRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsMoveTo(pos)
	if serie.WithPoint != nil {
		grp.Append(serie.WithPoint(pos))
	}
	for _, pt := range slices.Rest(serie.Points) {
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		ctrl := ori
		ctrl.X += (pos.X - ori.X) * r.stretch
		pat.AbsQuadraticCurve(pos, ctrl)
		ori = pos
		if serie.WithPoint != nil {
			grp.Append(serie.WithPoint(pos))
		}
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type stepRenderer[T, U ScalerConstraint] struct {
	ignoreMissing bool
}

func StepRender[T, U ScalerConstraint](ignoreMissing bool) Renderer[T, U] {
	return stepRenderer[T, U]{
		ignoreMissing: ignoreMissing,
	}
}

func (r stepRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
		nan bool
	)
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pat.AbsLineTo(pos)
	if serie.WithPoint != nil {
		grp.Append(serie.WithPoint(pos))
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)
		if nan && r.ignoreMissing {
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
		if serie.WithPoint != nil {
			grp.Append(serie.WithPoint(pos))
		}
	}
	if serie.WithArea {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func isFloat[T any](v T) (float64, bool) {
	x, ok := any(v).(float64)
	return x, ok
}

type linearRenderer[T, U ScalerConstraint] struct {
	ignoreMissing bool
}

func LinearRender[T, U ScalerConstraint](ignoreMissing bool) Renderer[T, U] {
	return linearRenderer[T, U]{
		ignoreMissing: ignoreMissing,
	}
}

func (r linearRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
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
		if i == 0 || (nan && r.ignoreMissing) {
			nan = false
			pat.AbsMoveTo(pos)
		} else {
			pat.AbsLineTo(pos)
		}
		if serie.WithPoint != nil {
			grp.Append(serie.WithPoint(pos))
		}
	}
	if serie.WithTitle {
		font := svg.NewFont(FontSize)
		font.Fill = serie.Color
		txt := svg.NewText(serie.Title)
		txt.Font = font
		txt.Baseline = "middle"
		txt.Pos = svg.NewPos(pos.X+10, pos.Y)
		grp.Append(txt.AsElement())
	}

	if serie.WithArea {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
		pos.X = serie.X.Min()
		pat.AbsLineTo(pos)
		pat.ClosePath()
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type stepAfterRenderer[T, U ScalerConstraint] struct {
	ignoreMissing bool
}

func StepAfterRender[T, U ScalerConstraint](ignoreMissing bool) Renderer[T, U] {
	return stepAfterRenderer[T, U]{
		ignoreMissing: ignoreMissing,
	}
}

func (r stepAfterRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos svg.Pos
		ori svg.Pos
		nan bool
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Max()
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	if serie.WithPoint != nil {
		grp.Append(serie.WithPoint(pos))
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		if nan && r.ignoreMissing {
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

		if serie.WithPoint != nil {
			grp.Append(serie.WithPoint(pos))
		}
	}
	if serie.WithArea {
		pos.X = serie.X.Max()
		pat.AbsLineTo(pos)
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type stepBeforeRenderer[T, U ScalerConstraint] struct {
	ignoreMissing bool
}

func StepBeforeRender[T, U ScalerConstraint](ignoreMissing bool) Renderer[T, U] {
	return stepBeforeRenderer[T, U]{
		ignoreMissing: ignoreMissing,
	}
}

func (r stepBeforeRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
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
	if serie.WithPoint != nil {
		grp.Append(serie.WithPoint(pos))
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		if nan && r.ignoreMissing {
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

		if serie.WithPoint != nil {
			grp.Append(serie.WithPoint(pos))
		}
	}
	if serie.WithArea {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type rendererFunc[T, U ScalerConstraint] struct {
	render RenderFunc[T, U]
}

func (r rendererFunc[T, U]) Render(serie Serie[T, U]) svg.Element {
	return r.render(serie)
}

func getBasePath(stroke string, fill bool) svg.Path {
	pat := svg.NewPath()
	pat.Stroke = svg.NewStroke(stroke, 1)
	if fill {
		pat.Fill = svg.NewFill(stroke)
		pat.Fill.Opacity = 0.5
	} else {
		pat.Fill = svg.NewFill("none")
	}
	return pat
}

func getBaseGroup() svg.Group {
	var g svg.Group
	g.Class = append(g.Class, "line")
	return g
}
