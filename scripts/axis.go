package main

import (
	"bufio"
	"os"
	"strconv"
	"time"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

const FontSize = 12.0

type Orientation int

const (
	OrientTop Orientation = -(iota + 1)
	OrientRight
	OrientBottom
	OrientLeft
)

func (o Orientation) Vertical() bool {
	return o == OrientLeft || o == OrientRight
}

func (o Orientation) Reverse() bool {
	return o == OrientRight || o == OrientTop
}

type Axis interface {
	Render(float64, float64, float64) svg.Element
}

type TimeAxis struct {
	Orientation
	Ticks  int
	Domain struct {
		First  time.Time
		Last   time.Time
		Values []time.Time
	}
	Format func(time.Time) string
}

func (a TimeAxis) Render(length, left, top float64) svg.Element {
	g := svg.NewGroup(svg.WithTranslate(left, top))
	d := domainLine(a.Orientation, length, svg.NewStroke("black", 1))
	g.Append(d.AsElement())

	var (
		interval = length / a.Range()
		font     = svg.NewFont(FontSize)
		first    = slices.Fst(a.Domain.Values)
		format   = a.Format
	)
	if format == nil {
		format = func(t time.Time) string {
			return t.Format("2006-01-02")
		}
	}
	for _, t := range a.Domain.Values {
		var (
			pos  = float64(t.Sub(first)) * interval
			grp  = svg.NewGroup(svg.WithTranslate(pos, 0))
			text = tickText(a.Orientation, format(t), 0, font)
			tick = innerTick(a.Orientation, 0, d.Stroke)
		)
    if a.Vertical() {
      grp.Transform.TX = 0
      grp.Transform.TY = pos
    }
		grp.Append(tick.AsElement())
		grp.Append(text.AsElement())
		g.Append(grp.AsElement())
	}

	return g.AsElement()
}

func (a TimeAxis) Range() float64 {
	var diff time.Duration
	if len(a.Domain.Values) >= 2 {
		diff = slices.Lst(a.Domain.Values).Sub(slices.Fst(a.Domain.Values))
	} else {
		diff = a.Domain.Last.Sub(a.Domain.First)
	}
	return float64(diff)
}

type NumberAxis struct {
	Orientation
	Ticks  int
	Domain struct {
		First  float64
		Last   float64
		Values []float64
	}
	Format func(float64) string
}

func (a NumberAxis) Render(length, left, top float64) svg.Element {
	g := svg.NewGroup(svg.WithTranslate(left, top))
	d := domainLine(a.Orientation, length, svg.NewStroke("black", 1))
	g.Append(d.AsElement())

	var (
		interval = length / a.Range()
		font     = svg.NewFont(FontSize)
		first    = slices.Fst(a.Domain.Values)
		format   = a.Format
	)
	if format == nil {
		format = func(f float64) string {
			return strconv.FormatFloat(f, 'f', 2, 64)
		}
	}
	for _, f := range a.Domain.Values {
		var (
			pos  = float64(f-first) * interval
			grp  = svg.NewGroup(svg.WithTranslate(pos, 0))
			text = tickText(a.Orientation, format(f), 0, font)
			tick = innerTick(a.Orientation, 0, d.Stroke)
		)
    if a.Vertical() {
      grp.Transform.TX = 0
      grp.Transform.TY = pos
    }
		grp.Append(tick.AsElement())
		grp.Append(text.AsElement())
		g.Append(grp.AsElement())
	}

	return g.AsElement()
}

func (a NumberAxis) Range() float64 {
	if len(a.Domain.Values) >= 2 {
		return slices.Lst(a.Domain.Values) - slices.Fst(a.Domain.Values)
	}
	return a.Domain.Last - a.Domain.First
}

type CategoryAxis struct {
	Orientation
	Ticks  int
	Domain []string
}

func (a CategoryAxis) Render(length, left, top float64) svg.Element {
	g := svg.NewGroup(svg.WithTranslate(left, top))
	d := domainLine(a.Orientation, length, svg.NewStroke("black", 1))
	g.Append(d.AsElement())

	var (
		interval = length / a.Range()
		align    = interval / 2
		font     = svg.NewFont(FontSize)
	)
	for i, s := range a.Domain {
		var (
			pos  = float64(i) * interval
			text = tickText(a.Orientation, s, align, font)
			tick = innerTick(a.Orientation, align, d.Stroke)
      grp  = svg.NewGroup(svg.WithTranslate(pos, 0))
		)
    if a.Vertical() {
      grp.Transform.TX = 0
      grp.Transform.TY = pos
    }
		grp.Append(tick.AsElement())
		grp.Append(text.AsElement())
		g.Append(grp.AsElement())
	}

	return g.AsElement()
}

func (a CategoryAxis) Range() float64 {
	return float64(len(a.Domain))
}

func domainLine(orient Orientation, length float64, stroke svg.Stroke) svg.Line {
	x, y := length, 0.0
	if orient.Vertical() {
		x, y = y, x
	}
	d := svg.NewLine(svg.NewPos(0, 0), svg.NewPos(x, y))
	d.Stroke = svg.NewStroke("black", 1)
	return d
}

func innerTick(orient Orientation, offset float64, stroke svg.Stroke) svg.Line {
  var (
    pos1 = svg.NewPos(offset, 0)
    pos2 = svg.NewPos(offset, FontSize*0.8)
  )
	switch {
	case orient.Vertical() && !orient.Reverse():
		pos2.X, pos2.Y = -pos2.Y, pos2.X
    pos1.X, pos1.Y = 0, offset
	case orient.Vertical() && orient.Reverse():
		pos2.X, pos2.Y = pos2.Y, pos2.X
    pos1.X, pos1.Y = 0, offset
	case !orient.Vertical() && orient.Reverse():
	  pos2.Y = -pos2.Y
	default:
	}
	tick := svg.NewLine(pos1, pos2)
	tick.Stroke = stroke
	return tick
}

func tickText(orient Orientation, str string, offset float64, font svg.Font) svg.Text {
	var (
		base   = "hanging"
		anchor = "middle"
		x, y   = offset, FontSize * 1.2
	)
	switch {
	case orient.Vertical() && !orient.Reverse():
		base = "middle"
		anchor = "end"
		x, y = -y, x
	case orient.Vertical() && orient.Reverse():
		base = "middle"
		anchor = "start"
		x, y = y, x
	case !orient.Vertical() && orient.Reverse():
		base = "auto"
		y = -y
	default:
	}
	text := svg.NewText(str)
	text.Pos = svg.NewPos(x, y)
	text.Font = font
	text.Anchor = anchor
	text.Baseline = base
	return text
}

func main() {
	orient := OrientLeft
	area := svg.NewSVG(svg.WithDimension(800, 800))
	cat := CategoryAxis{
		Orientation: orient,
		Domain:      []string{"go", "python", "javascript", "rust", "c++"},
	}
	elem := cat.Render(720, 100, 40)
	area.Append(elem)

	var tim TimeAxis
	tim.Orientation = orient
	tim.Domain.Values = []time.Time{
		time.Date(2022, 6, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 7, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 9, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 11, 19, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	elem = tim.Render(720, 300, 40)
	area.Append(elem)

	var num NumberAxis
	num.Orientation = orient
	num.Domain.Values = []float64{
		1.0,
		6.67,
		37.19,
		67.1,
		88.9,
		110,
		128.1981,
	}

	elem = num.Render(720, 500, 40)
	area.Append(elem)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	area.Render(w)
}
