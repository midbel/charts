package main

import (
	"bufio"
	"os"
	"time"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
	"github.com/midbel/charts"
)

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

type Serie[T, U charts.ScalerConstraint] struct {
	WithPoint bool
	WithArea bool
	Color     string

	Points []charts.Point[T, U]
	X      charts.Scaler[T]
	Y      charts.Scaler[U]

	Renderer RenderFunc[T, U]
}

func (s Serie[T, U]) Render() svg.Element {
	return s.Renderer(s)
}

type Renderer[T, U charts.ScalerConstraint] interface {
	Render([]charts.Point[T, U]) svg.Element
}

type RenderFunc[T, U charts.ScalerConstraint] func(Serie[T, U]) svg.Element

func stepRender[T, U charts.ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos svg.Pos
		ori svg.Pos
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsMoveTo(pos)
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
		// pos.X = ori
		pos.X = serie.X.Min()
		pat.AbsLineTo(pos)
		pat.ClosePath()
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func linearRender[T, U charts.ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos svg.Pos
		// ori = serie.X.Scale(slices.Fst(serie.Points).X)
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

func stepAfterRender[T, U charts.ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos svg.Pos
		ori svg.Pos
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsMoveTo(pos)
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
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
		pos.X = serie.X.Min()
		pat.AbsLineTo(pos)
		pat.ClosePath()
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func stepBeforeRender[T, U charts.ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color, serie.WithArea)
		pos svg.Pos
		ori svg.Pos
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsMoveTo(pos)
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
		pos.X = serie.X.Min()
		pat.AbsLineTo(pos)
		pat.ClosePath()
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func main() {
	var (
		size   = charts.NewRange(0, 720)
		orient = charts.OrientTop
		area   = svg.NewSVG(svg.WithDimension(800, 800))
	)
	cat := charts.CategoryAxis{
		Orientation: orient,
		Domain:      []string{"go", "python", "javascript", "rust", "c++"},
	}
	elem := cat.Render(720, 40, 40)
	area.Append(elem)

	var (
		tim     charts.TimeAxis
		dtstart = time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC)
		dtend   = time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)
	)
	tim.Ticks = 6
	tim.Orientation = orient
	tim.Scaler = charts.TimeScaler(charts.TimeDomain(dtstart, dtend), size)
	tim.Domain = []time.Time{
		time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 9, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 9, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 9, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 9, 26, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
	}
	elem = tim.Render(720, 40, 240)
	area.Append(elem)

	var num charts.NumberAxis
	num.Orientation = orient
	num.Scaler = charts.NumberScaler(charts.NumberDomain(0, 130), size)
	num.Domain = []float64{
		1.0,
		6.67,
		37.19,
		67.1,
		88.9,
		110,
		128.1981,
	}

	elem = num.Render(720, 40, 440)
	area.Append(elem)

	var serie1 Serie[float64, float64]
	serie1.Renderer = linearRender[float64, float64]
	serie1.Color = "blue"
	serie1.WithArea = true
	serie1.WithPoint = true
	serie1.X = charts.NumberScaler(charts.NumberDomain(0, 720), charts.NewRange(0, 720))
	serie1.Y = charts.NumberScaler(charts.NumberDomain(720, 0), charts.NewRange(0, 540))
	serie1.Points = []charts.Point[float64, float64]{
		charts.NumberPoint(100, 245),
		charts.NumberPoint(150, 567),
		charts.NumberPoint(324, 98),
		charts.NumberPoint(461, 19),
		charts.NumberPoint(511, 563),
		charts.NumberPoint(541, 463),
		charts.NumberPoint(571, 113),
		charts.NumberPoint(591, 703),
		charts.NumberPoint(645, 301),
		charts.NumberPoint(716, 341),
	}

	pat := serie1.Render()
	area.Append(pat)

	dtstart = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	dtend = time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)

	var serie2 Serie[time.Time, float64]
	serie2.Renderer = linearRender[time.Time, float64]
	serie2.Color = "red"
	serie2.WithArea = true
	serie2.WithPoint = true
	serie2.X = charts.TimeScaler(charts.TimeDomain(dtstart, dtend), charts.NewRange(0, 720))
	serie2.Y = charts.NumberScaler(charts.NumberDomain(100, 0), charts.NewRange(0, 540))
	serie2.Points = []charts.Point[time.Time, float64]{
		charts.TimePoint(time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC), 34),
		charts.TimePoint(time.Date(2022, 1, 20, 0, 0, 0, 0, time.UTC), 39),
		charts.TimePoint(time.Date(2022, 2, 10, 0, 0, 0, 0, time.UTC), 40),
		charts.TimePoint(time.Date(2022, 2, 26, 0, 0, 0, 0, time.UTC), 45),
		charts.TimePoint(time.Date(2022, 3, 7, 0, 0, 0, 0, time.UTC), 43),
		charts.TimePoint(time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC), 43),
		charts.TimePoint(time.Date(2022, 6, 11, 0, 0, 0, 0, time.UTC), 67),
		charts.TimePoint(time.Date(2022, 6, 29, 0, 0, 0, 0, time.UTC), 80),
		charts.TimePoint(time.Date(2022, 7, 6, 0, 0, 0, 0, time.UTC), 89),
		charts.TimePoint(time.Date(2022, 9, 5, 0, 0, 0, 0, time.UTC), 98),
		charts.TimePoint(time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC), 19),
		charts.TimePoint(time.Date(2022, 11, 19, 0, 0, 0, 0, time.UTC), 98),
		charts.TimePoint(time.Date(2022, 12, 6, 0, 0, 0, 0, time.UTC), 86),
		charts.TimePoint(time.Date(2022, 12, 16, 0, 0, 0, 0, time.UTC), 54),
		charts.TimePoint(time.Date(2022, 12, 25, 0, 0, 0, 0, time.UTC), 12),
		charts.TimePoint(time.Date(2022, 12, 30, 0, 0, 0, 0, time.UTC), 1),
	}

	pat = serie2.Render()
	area.Append(pat)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	area.Render(w)
}
