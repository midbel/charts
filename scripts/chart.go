package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	// "strconv"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

type Padding struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

func NewPadding(top, right, bottom, left float64) Padding {
	return Padding{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

func NewPadding2(horiz, vert float64) Padding {
	return NewPadding(vert, horiz, vert, horiz)
}

type Placement int

const (
	PlacementTop Placement = iota
	PlacementRight
	PlacementBottom
	PlacementLeft
)

func (p Placement) Vertical() bool {
	return p == PlacementLeft || p == PlacementRight
}

func (p Placement) Reverse() bool {
	return p == PlacementRight || p == PlacementTop
}

type Axis struct {
	Placement
	Domain []float64
	Ticks  int
	Label  string
	Stroke string

	WithLabel      bool
	WithInnerTicks bool
	WithOuterTicks bool
}

func DefaultAxis() Axis {
	return Axis{
		Ticks:          10,
		Stroke:         "black",
		WithLabel:      true,
		WithInnerTicks: true,
		WithOuterTicks: true,
	}
}

func CreateAxis(from, to float64) Axis {
	a := DefaultAxis()
	a.Domain = append(a.Domain, from, to)
	return a
}

func (a *Axis) Diff() float64 {
	return slices.Lst(a.Domain) - slices.Fst(a.Domain)
}

func (a *Axis) SetDomain(values []float64) {
	a.Domain = append(a.Domain[:0], values...)
	sort.Float64s(a.Domain)
}

func (a *Axis) DomainValues() []float64 {
	if len(a.Domain) != 2 {
		if a.Vertical() {
			return slices.Reverse(a.Domain)
		}
		return a.Domain
	}
	var (
		diff   = a.Diff() / float64(a.Ticks)
		base   = slices.Fst(a.Domain)
		values []float64
	)
	if a.Vertical() {
		base = slices.Lst(a.Domain)
		diff = -diff
	}
	for i := 0; i <= a.Ticks; i++ {
		v := (diff * float64(i)) + base
		values = append(values, v)
	}
	return values
}

func (a *Axis) TicksCount() int {
	if n := len(a.Domain); n != 2 {
		return n
	}
	return a.Ticks
}

type Chart struct {
	Title  string
	Width  float64
	Height float64
	Legend struct {
		Placement
		Title string
	}
	Padding
	Axis map[Placement]Axis
}

func DefaultChart() Chart {
	return Chart{
		Width:  480,
		Height: 360,
		Axis:   make(map[Placement]Axis),
	}
}

func (c Chart) DrawingWidth() float64 {
	return c.Width - c.Padding.Left - c.Padding.Right
}

func (c Chart) DrawingHeight() float64 {
	return c.Height - c.Padding.Top - c.Padding.Bottom
}

func (c Chart) AddAxis(where Placement, axis Axis) {
	axis.Placement = where
	c.Axis[where] = axis
}

func (c Chart) Render(w io.Writer, series []Serie) {
	el := svg.NewSVG(svg.WithDimension(c.Width, c.Height))
	el.Append(c.drawAxis())

	area := c.drawArea()
	for _, s := range series {
		el := c.drawSerie(s)
		if el == nil {
			continue
		}
		area.Append(el)
	}
	el.Append(area.AsElement())

	if c.showLegend() {

	}

	bw := bufio.NewWriter(w)
	defer bw.Flush()
	el.Render(bw)
}

func (c Chart) showLegend() bool {
	return c.Legend.Placement != 0 && c.Legend.Title != ""
}

func (c Chart) getBottomAxis() Axis {
	return c.Axis[PlacementBottom]
}

func (c Chart) getLeftAxis() Axis {
	return c.Axis[PlacementLeft]
}

func (c Chart) drawSerie(serie Serie) svg.Element {
	var (
		xax  = c.getBottomAxis()
		yax  = c.getLeftAxis()
		offx = c.DrawingWidth() / xax.Diff()
		offy = c.DrawingHeight() / yax.Diff()
		fstx = slices.Fst(xax.Domain)
		fsty = slices.Fst(yax.Domain)
		grp  = svg.NewGroup()
		pat  = svg.NewPath(svg.WithFill(svg.NewFill("none")), svg.WithStroke(svg.NewStroke(serie.Stroke, 2)))
	)
	for i, pt := range serie.Values {
		var (
			x = (pt.X - fstx) * offx
			y = c.DrawingHeight() - ((pt.Y - fsty) * offy)
			s = []svg.Option{
				svg.WithPosition(x, y),
				svg.WithFill(svg.NewFill(serie.Stroke)),
				svg.WithRadius(4),
			}
			ci = svg.NewCircle(s...)
		)
		ci.Title = fmt.Sprintf("%.1f - %.1f", pt.X, pt.Y)
		if p := svg.NewPos(x, y); i == 0 {
			pat.AbsMoveTo(p)
		} else {
			pat.AbsLineTo(p)
		}
		grp.Append(ci.AsElement())
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func (c Chart) drawAxis() svg.Element {
	g := svg.NewGroup(svg.WithID("axis"))
	if a, ok := c.Axis[PlacementBottom]; ok {
		ga := makeDomainLine(a, c.DrawingWidth(), 0, c.Padding.Left, c.Height-c.Padding.Bottom)
		gx := makeTicks(ga, a, c.DrawingWidth(), c.DrawingHeight())
		g.Append(gx)
	}
	if a, ok := c.Axis[PlacementLeft]; ok {
		ga := makeDomainLine(a, 0, c.DrawingHeight(), c.Padding.Left, c.Padding.Top)
		gx := makeTicks(ga, a, c.DrawingHeight(), c.DrawingWidth())
		g.Append(gx)
	}
	if a, ok := c.Axis[PlacementTop]; ok {
		ga := makeDomainLine(a, c.DrawingWidth(), 0, c.Padding.Left, c.Padding.Top)
		gx := makeTicks(ga, a, c.DrawingWidth(), 0)
		g.Append(gx)
	}
	if a, ok := c.Axis[PlacementRight]; ok {
		ga := makeDomainLine(a, 0, c.DrawingHeight(), c.Width-c.Padding.Right, c.Padding.Top)
		gx := makeTicks(ga, a, c.DrawingHeight(), 0)
		g.Append(gx)
	}
	return g.AsElement()
}

func makeTicks(ga svg.Group, a Axis, length, size float64) svg.Element {
	var (
		ticks  = a.TicksCount()
		offset = length / a.Diff()
		values = a.DomainValues()
		first  = slices.Fst(values)
	)
	if a.Vertical() {
		first = slices.Lst(values)
	}
	for i, v := range values {
		pt := NewPoint((v-first)*offset, 0.0)
		if a.Vertical() {
			pt.X, pt.Y = pt.Y, length-pt.X
		}
		g := svg.NewGroup(svg.WithTranslate(pt.X, pt.Y))
		if a.WithInnerTicks {
			line := innerTickLine(a)
			g.Append(line.AsElement())
		}
		if a.WithOuterTicks && i > 0 && i < ticks {
			line := outerTickLine(a, size)
			g.Append(line.AsElement())
		}
		if a.WithLabel {
			txt := tickText(a, v)
			g.Append(txt.AsElement())
		}
		ga.Append(g.AsElement())
	}
	return ga.AsElement()
}

var font = svg.NewFont(10)

func outerTickLine(a Axis, length float64) svg.Line {
	line := NewPoint(0, -length)
	switch {
	case a.Vertical() && !a.Reverse():
		line.X, line.Y = -line.Y, line.X
	case a.Vertical() && a.Reverse():
		line.X, line.Y = -line.Y, line.X
		line.X = -line.X
	case !a.Vertical() && a.Reverse():
		line.Y = -line.Y
	default:
	}
	var (
		pos1   = svg.NewPos(0, 0)
		pos2   = svg.NewPos(line.X, line.Y)
		stroke = svg.NewStroke(a.Stroke, 0.25)
	)
	return svg.NewLine(pos1, pos2, svg.WithStroke(stroke))
}

func innerTickLine(a Axis) svg.Line {
	line := NewPoint(0, 5)
	switch {
	case a.Vertical() && !a.Reverse():
		line.X, line.Y = -line.Y, line.X
	case a.Vertical() && a.Reverse():
		line.X, line.Y = -line.Y, line.X
		line.X = -line.X
	case !a.Vertical() && a.Reverse():
		line.Y = -line.Y
	default:
	}
	var (
		pos1   = svg.NewPos(0, 0)
		pos2   = svg.NewPos(line.X, line.Y)
		stroke = svg.NewStroke(a.Stroke, 0.25)
	)
	return svg.NewLine(pos1, pos2, svg.WithStroke(stroke))
}

func tickText(a Axis, value float64) svg.Text {
	var (
		base   = "hanging"
		anchor = "middle"
		point  = NewPoint(0, 10)
	)
	switch {
	case a.Vertical() && !a.Reverse():
		base = "middle"
		anchor = "end"
		point.X, point.Y = -point.Y, point.X
	case a.Vertical() && a.Reverse():
		base = "middle"
		anchor = "start"
		point.X, point.Y = point.Y, point.X
	case !a.Vertical() && a.Reverse():
		base = "auto"
		point.Y = -point.Y
	default:
	}
	options := []svg.Option{
		svg.WithFont(font),
		svg.WithPosition(point.X, point.Y),
		svg.WithDominantBaseline(base),
		svg.WithAnchor(anchor),
	}
	str := fmt.Sprintf("%.1f", value)
	return svg.NewText(str, options...)
}

//
// func tickText(a Axis, offset float64) svg.Text {
// 	var (
// 		base   = "hanging"
// 		anchor = "middle"
// 		point  = NewPoint(0, 10)
// 		value  = slices.Fst(a.Domain)
// 	)
// 	if a.Vertical() {
// 		value = slices.Lst(a.Domain)
// 		offset = -offset
// 	}
// 	switch {
// 	case a.Vertical() && !a.Reverse():
// 		base = "middle"
// 		anchor = "end"
// 		point.X, point.Y = -point.Y, point.X
// 	case a.Vertical() && a.Reverse():
// 		base = "middle"
// 		anchor = "start"
// 		point.X, point.Y = point.Y, point.X
// 	case !a.Vertical() && a.Reverse():
// 		base = "auto"
// 		point.Y = -point.Y
// 	default:
// 	}
// 	options := []svg.Option{
// 		svg.WithFont(font),
// 		svg.WithPosition(point.X, point.Y),
// 		svg.WithDominantBaseline(base),
// 		svg.WithAnchor(anchor),
// 	}
// 	str := fmt.Sprintf("%.1f", value+offset)
// 	return svg.NewText(str, options...)
// }

func (c Chart) drawArea() svg.Group {
	rec := svg.NewRect(svg.WithDimension(c.DrawingWidth(), c.DrawingHeight()))
	gos := []svg.Option{
		svg.WithID("area"),
		svg.WithTranslate(c.Padding.Left, c.Padding.Top),
	}
	g := svg.NewGroup(gos...)
	g.Append(rec.AsElement())

	return g
}

func makeDomainLine(a Axis, x, y, left, top float64) svg.Group {
	var (
		ga = svg.NewGroup(svg.WithTranslate(left, top))
		sk = svg.NewStroke(a.Stroke, 0.75)
		li = svg.NewLine(svg.NewPos(0, 0), svg.NewPos(x, y), svg.WithStroke(sk))
	)
	ga.Append(li.AsElement())
	return ga
}

type Point struct {
	X float64
	Y float64
}

func (p Point) Inverse() Point {
	p.X, p.Y = p.Y, p.X
	return p
}

func NewPoint(x, y float64) Point {
	return Point{
		X: x,
		Y: y,
	}
}

type Serie struct {
	Label  string
	Values []Point

	Stroke string
}

func (s *Serie) Add(x, y float64) {
	s.Values = append(s.Values, NewPoint(x, y))
}

func main() {
	var ser1 Serie
	ser1.Label = "serie 1 (blue)"
	ser1.Stroke = "blue"
	ser1.Add(-3, -3)
	ser1.Add(-1, -5)
	ser1.Add(5, 9)
	ser1.Add(5, 9)
	ser1.Add(12, 2)
	ser1.Add(17, 3)
	ser1.Add(27, -1)
	ser1.Add(31, -6)
	ser1.Add(36, 0)
	ser1.Add(43, 10)
	ser1.Add(47, 13)
	ser1.Add(57, 11)
	ser1.Add(63, 7)
	ser1.Add(69, 12)

	var ser2 Serie
	ser2.Label = "serie 2 (red)"
	ser2.Stroke = "red"
	ser2.Add(-6, 5)
	ser2.Add(3, 7)
	ser2.Add(19, 3)
	ser2.Add(28, 0)
	ser2.Add(37, 11)
	ser2.Add(40, 11)
	ser2.Add(56, 4)
	ser2.Add(61, 1)
	ser2.Add(67, 7)
	ser2.Add(69, 8)

	var (
		ch  = DefaultChart()
		xax = CreateAxis(-7, 69)
		yax = CreateAxis(-6, 13)
	)
	xax.WithOuterTicks = false
	yax.SetDomain([]float64{-6, -2, 2, 6, 10, 14})

	ch.Width = 800
	ch.Height = 600
	ch.Title = "sample chart"
	ch.Padding = NewPadding(20, 20, 40, 40)
	ch.AddAxis(PlacementBottom, xax)
	ch.AddAxis(PlacementLeft, yax)
	ch.Render(os.Stdout, []Serie{ser1, ser2})
	fmt.Println()
}
