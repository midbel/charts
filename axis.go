package charts

import (
	"math"

	"github.com/midbel/svg"
)

const FontSize = 12.0

type Orientation int

const (
	OrientTop Orientation = 1 << iota
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

type Axis[T ScalerConstraint] struct {
	Label  string
	Rotate float64
	Orientation
	Ticks          int
	Scaler         Scaler[T]
	Domain         []T
	Format         func(T) string
	WithInnerTicks bool
	WithLabelTicks bool
	WithOuterTicks bool
	WithBands      bool
	WithArrow      bool
}

func (a Axis[T]) Render(length, size, left, top float64) svg.Element {
	var g svg.Group
	g.Transform = svg.Translate(left, top)
	if a.Vertical() {
		g.Class = append(g.Class, "axis", "y-axis")
	} else {
		g.Class = append(g.Class, "axis", "x-axis")
	}
	d := domainLine(a.Orientation, length, svg.NewStroke("black", 1))
	g.Append(d.AsElement())

	var (
		data = a.Domain
		font = svg.NewFont(FontSize)
	)
	if len(data) == 0 {
		data = a.Scaler.Values(a.Ticks)
	}
	for i, t := range data {
		var (
			pos = a.Scaler.Scale(t)
			grp svg.Group
		)
		if a.Vertical() {
			grp.Transform = svg.Translate(0, pos)
		} else {
			grp.Transform = svg.Translate(pos, 0)
		}
		if a.WithInnerTicks {
			tick := lineTick(a.Orientation, 0, FontSize*0.8, d.Stroke)
			grp.Append(tick.AsElement())
		}
		if a.WithLabelTicks && a.Format != nil {
			text := tickText(a.Orientation, a.Format(t), 0, a.Rotate, font)
			grp.Append(text.AsElement())
		}
		if i < len(data)-1 {
			if a.WithOuterTicks {
				sk := d.Stroke
				sk.Opacity = 0.05
				tick := lineTick(a.Orientation, 0, -size, sk)
				grp.Append(tick.AsElement())
			}
			if i%2 == 0 && a.WithBands {
				rec := bandTick(a.Orientation, size, length/float64(len(data)-1))
				grp.Append(rec.AsElement())
			}
		}
		g.Append(grp.AsElement())
	}

	return g.AsElement()
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

func bandTick(orient Orientation, width, height float64) svg.Rect {
	var rec svg.Rect
	rec.Pos = svg.NewPos(0, 0)
	rec.Dim = svg.NewDim(width, height)
	switch {
	case !orient.Vertical() && orient.Reverse():
		rec.Dim.W, rec.Dim.H = rec.Dim.H, rec.Dim.W
	case !orient.Vertical() && !orient.Reverse():
		rec.Dim.W, rec.Dim.H = rec.Dim.H, rec.Dim.W
		rec.Transform.RA = 180
		rec.Transform.TX = rec.Dim.W
	case orient.Vertical() && orient.Reverse():
		rec.Transform.TX = -rec.Dim.W
	default:
	}
	rec.Fill = svg.NewFill("currentColor")
	rec.Fill.Opacity = 0.05
	return rec
}

func lineTick(orient Orientation, offset, size float64, stroke svg.Stroke) svg.Line {
	var (
		pos1 = svg.NewPos(offset, 0)
		pos2 = svg.NewPos(offset, size)
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

func tickText(orient Orientation, str string, offset, rotate float64, font svg.Font) svg.Text {
	var (
		base   = "hanging"
		anchor = "middle"
		x, y   = offset, FontSize * 1.2
	)
	if rotate != 0 {
		anchor = "end"
	}
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
	if rotate != 0 {
		anchor = "end"
		if orient.Vertical() && math.Abs(rotate) == 90 {
			anchor = "middle"
		}
	}
	text := svg.NewText(str)
	text.Transform.RA = rotate
	text.Transform.RX = 0
	text.Transform.RY = 0
	text.Pos = svg.NewPos(x, y)
	text.Font = font
	text.Anchor = anchor
	text.Baseline = base
	return text
}
