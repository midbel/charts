package charts

import (
	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

type Renderer[T, U ScalerConstraint] interface {
	Render(Serie[T, U]) svg.Element
}

type RenderFunc[T, U ScalerConstraint] func(Serie[T, U]) svg.Element

func LinearRender[T, U ScalerConstraint]() Renderer[T, U] {
	return rendererFunc[T, U]{
		render: linearRender[T, U],
	}
}

func StepRender[T, U ScalerConstraint]() Renderer[T, U] {
	return rendererFunc[T, U]{
		render: stepRender[T, U],
	}
}

func StepBeforeRender[T, U ScalerConstraint]() Renderer[T, U] {
	return rendererFunc[T, U]{
		render: stepBeforeRender[T, U],
	}
}

func StepAfterRender[T, U ScalerConstraint]() Renderer[T, U] {
	return rendererFunc[T, U]{
		render: stepAfterRender[T, U],
	}
}

func stepRender[T, U ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
	)
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pat.AbsLineTo(pos)
	if serie.WithPoint {
		ci := getCircle(pos, serie.Color)
		grp.Append(ci.AsElement())
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		ori.X += (pos.X - ori.X) / 2
		pat.AbsLineTo(ori)
		ori.Y = pos.Y
		pat.AbsLineTo(ori)
		pat.AbsLineTo(pos)
		ori = pos
		if serie.WithPoint {
			ci := getCircle(pos, serie.Color)
			grp.Append(ci.AsElement())
		}
	}
	if serie.WithArea {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func linearRender[T, U ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos svg.Pos
	)
	for i, pt := range serie.Points {
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)
		if i == 0 {
			pat.AbsMoveTo(pos)
		} else {
			pat.AbsLineTo(pos)
		}
		if serie.WithPoint {
			ci := getCircle(pos, serie.Color)
			grp.Append(ci.AsElement())
		}
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

func stepAfterRender[T, U ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos svg.Pos
		ori svg.Pos
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Max()
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	if serie.WithPoint {
		ci := getCircle(pos, serie.Color)
		grp.Append(ci.AsElement())
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		ori.X = pos.X
		pat.AbsLineTo(ori)
		ori.Y = pos.Y
		pat.AbsLineTo(ori)
		pat.AbsLineTo(pos)
		ori = pos

		if serie.WithPoint {
			ci := getCircle(pos, serie.Color)
			grp.Append(ci.AsElement())
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

func stepBeforeRender[T, U ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos svg.Pos
		ori svg.Pos
	)
	pos.X = serie.X.Min()
	pos.Y = serie.Y.Max()
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pat.AbsLineTo(pos)
	if serie.WithPoint {
		ci := getCircle(pos, serie.Color)
		grp.Append(ci.AsElement())
	}
	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		ori.Y = pos.Y
		pat.AbsLineTo(ori)
		ori.X = pos.X
		pat.AbsLineTo(ori)
		pat.AbsLineTo(pos)
		ori = pos

		if serie.WithPoint {
			ci := getCircle(pos, serie.Color)
			grp.Append(ci.AsElement())
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

func getCircle(pos svg.Pos, fill string) svg.Circle {
	ci := svg.NewCircle()
	ci.Radius = 5
	ci.Pos = pos
	ci.Fill = svg.NewFill(fill)
	return ci
}
