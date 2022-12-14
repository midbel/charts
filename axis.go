package charts

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/midbel/svg"
)

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
		data   = a.Domain
		offset float64
		font   = svg.NewFont(FontSize)
	)
	if _, ok := any(a).(Axis[string]); ok {
		offset = a.Scaler.Space() / 2
	}
	if len(data) == 0 {
		data = a.Scaler.Values(a.Ticks)
	}
	if a.Format == nil {
		a.Format = defaultLabelFormat[T]
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
			tick := lineTick(a.Orientation, offset, FontSize*0.8, d.Stroke)
			grp.Append(tick.AsElement())
		}
		if a.WithLabelTicks && a.Format != nil {
			text := tickText(a.Orientation, a.Format(t), offset, a.Rotate, font)
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
	d.Class = append(d.Class, "domain")
	d.Stroke = svg.NewStroke("black", 1)
	return d
}

func bandTick(orient Orientation, width, height float64) svg.Rect {
	var rec svg.Rect
	rec.Class = append(rec.Class, "band-ticks")
	rec.Pos = svg.NewPos(0, 0)
	rec.Dim = svg.NewDim(width, height)
	switch orient {
	case OrientTop:
		rec.Dim.W, rec.Dim.H = rec.Dim.H, rec.Dim.W
	case OrientBottom:
		rec.Dim.W, rec.Dim.H = rec.Dim.H, rec.Dim.W
		rec.Transform.RA = 180
		rec.Transform.TX = rec.Dim.W
	case OrientRight:
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
	switch orient {
	case OrientLeft:
		pos2.X, pos2.Y = -pos2.Y, pos2.X
		pos1.X, pos1.Y = 0, offset
	case OrientRight:
		pos2.X, pos2.Y = pos2.Y, pos2.X
		pos1.X, pos1.Y = 0, offset
	case OrientTop:
		pos2.Y = -pos2.Y
	default:
	}
	tick := svg.NewLine(pos1, pos2)
	tick.Class = append(tick.Class, "tick")
	tick.Stroke = stroke
	return tick
}

func axisText(orient Orientation, str string, x, y float64, font svg.Font) svg.Text {
	txt := svg.NewText(str)
	txt.Font = font
	txt.Anchor = "middle"
	txt.Baseline = "auto"
	txt.Pos = svg.NewPos(x, y)
	txt.Class = append(txt.Class, "axis-label")
	switch orient {
	case OrientBottom:
		txt.Baseline = "start"
	case OrientTop:
		txt.Baseline = "hanging"
	case OrientLeft, OrientRight:
		txt.Transform.RX = txt.Pos.X
		txt.Transform.RY = txt.Pos.Y
		txt.Transform.RA = -90
	}
	return txt
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
	switch orient {
	case OrientLeft:
		base = "middle"
		anchor = "end"
		x, y = -y, x
	case OrientRight:
		base = "middle"
		anchor = "start"
		x, y = y, x
	case OrientTop:
		base = "auto"
		y = -y
	default:
	}
	txt := svg.NewText(str)
	txt.Pos = svg.NewPos(x, y)
	txt.Font = font
	txt.Anchor = anchor
	txt.Baseline = base
	txt.Class = append(txt.Class, "tick-label")
	return rotateText(orient, rotate, txt)
}

func normRotate(rotate float64) float64 {
	angle := math.Mod(math.Abs(rotate), 90)
	if angle == 0 {
		return 90
	}
	return angle
}

func rotateText(orient Orientation, rotate float64, text svg.Text) svg.Text {
	if rotate == 0 {
		return text
	}
	angle := normRotate(rotate)
	if orient == OrientLeft || orient == OrientBottom {
		angle = -angle
	}
	ratio := 90 / angle
	text.Anchor, text.Baseline = "end", "middle"
	if orient.Vertical() {
		text.Baseline = "start"
		if orient == OrientRight {
			text.Anchor = "start"
		}
		text.Pos.Y = -text.Pos.X / ratio
		if abs := math.Abs(angle); abs == 90 {
			text.Pos.X = 0
		} else if abs < 45 {
			text.Pos.X *= 1.2
		}
	} else {
		text.Pos.X = text.Pos.Y / ratio
		if abs := math.Abs(angle); abs == 90 {
			text.Pos.Y = 0
		} else if abs < 45 {
			text.Pos.Y *= 1.2
		}
	}
	text.Transform.RA = angle
	return text
}

func defaultLabelFormat[T ScalerConstraint](v T) string {
	switch v := any(v).(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', 3, 64)
	case string:
		return v
	case time.Time:
		return v.Format("2006-01-02")
	default:
		return fmt.Sprintf("%v", v)
	}
}
