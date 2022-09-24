package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

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

func (p Padding) Horiz() float64 {
	return p.Left + p.Right
}

func (p Padding) Vert() float64 {
	return p.Top + p.Bottom
}

type Alignment int

const (
	AlignTop Alignment = iota
	AlignTopRight
	AlignRight
	AlignBottomRight
	AlignBottom
	AlignBottomLeft
	AlignLeft
	AlignTopLeft
)

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
	// Domain struct {
	// 	Values []float64
	// 	Start  float64
	// 	End    float64
	// }
	Ticks int
	Label string
	Color string
	Format func(float64) string

	Font struct {
		Size   int
		Family string
		Color  string
	}

	WithTickLabel  bool
	WithInnerTicks bool
	WithOuterTicks bool
}

func DefaultAxis() Axis {
	return Axis{
		Ticks:          10,
		Color:          "black",
		WithTickLabel:  true,
		WithInnerTicks: true,
		WithOuterTicks: true,
		Format: func(v float64) string {
			return fmt.Sprintf("%.1f", v)
		},
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
		Align Alignment
		Title string
	}
	Padding
	Axis map[Placement]Axis
	// Axis struct {
	// 	Top    *Axis
	// 	Right  *Axis
	// 	Left   *Axis
	// 	Bottom *Axis
	// }

	Background string // background color
	Area       string // area background color
	Font       struct {
		Size   int
		Family string
		Color  string
	}
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
	if c.Background != "" {
		rec := svg.NewRect(svg.WithDimension(c.Width, c.Height))
		rec.Fill = svg.NewFill(c.Background)
		el.Append(rec.AsElement())
	}
	area := c.drawArea()
	for _, s := range series {
		el := c.drawSerie(s)
		if el == nil {
			continue
		}
		area.Append(el)
	}
	el.Append(area.AsElement())
	el.Append(c.drawAxis())
	if g := c.drawLegend(series); g != nil {
		el.Append(c.drawLegend(series))
	}

	bw := bufio.NewWriter(w)
	defer bw.Flush()
	el.Render(bw)
}

func (c Chart) drawLegend(series []Serie) svg.Element {
	var (
		pt Point
		ln int
	)
	for i, s := range series {
		n := len(s.Label) * 5
		if i == 0 || n > ln {
			ln = n
		}
	}
	if ln == 0 {
		return nil
	}
	switch c.Legend.Align {
	case AlignTop:
		pt.X = (c.DrawingWidth() + float64(ln)) / 2
		pt.Y = c.Padding.Top
	case AlignTopRight:
		pt.X = c.DrawingWidth() - float64(ln)
		pt.Y = c.Padding.Top
	case AlignRight:
		pt.X = c.DrawingWidth() - float64(ln)
		pt.Y = (c.DrawingHeight() + float64(len(series)*15)) / 2
	case AlignBottomRight:
		pt.X = c.DrawingWidth() - float64(ln)
		pt.Y = c.DrawingHeight() - float64(len(series)*15)
	case AlignBottom:
		pt.X = (c.DrawingWidth() + float64(ln)) / 2
		pt.Y = c.DrawingHeight() - float64(len(series)*15)
	case AlignBottomLeft:
		pt.X = c.Padding.Left
		pt.Y = float64(len(series) * 15)
	case AlignLeft:
		pt.X = c.Padding.Left
		pt.Y = (c.DrawingHeight() + float64(len(series)*15)) / 2
	case AlignTopLeft:
		pt.X = c.Padding.Left
		pt.Y = c.Padding.Top
	default:
		return nil
	}
	g := svg.NewGroup(svg.WithTranslate(pt.X, pt.Y))
	t := 20.0
	for i, s := range series {
		off := float64(i) * 15
		opts := []svg.Option{
			svg.WithPosition(t+5, off),
			svg.WithFont(font),
			svg.WithFill(svg.NewFill(s.Color)),
			svg.WithDominantBaseline("middle"),
		}
		txt := svg.NewText(s.Label, opts...)
		stroke := svg.NewStroke(s.Color, 2)
		line := svg.NewLine(svg.NewPos(0, off), svg.NewPos(t, off), stroke.Option())
		g.Append(txt.AsElement())
		g.Append(line.AsElement())
	}
	return g.AsElement()
}

func (c Chart) getBottomAxis() Axis {
	return c.Axis[PlacementBottom]
}

func (c Chart) getLeftAxis() Axis {
	return c.Axis[PlacementLeft]
}

func (c Chart) drawSerie(serie Serie) svg.Element {
	if serie.Curve == nil {
		serie.Curve = Linear
	}
	var (
		li = serie.Curve(c.DrawingWidth(), c.DrawingHeight())
		el = li.Draw(serie, c.getBottomAxis(), c.getLeftAxis())
		gp = svg.NewGroup(svg.WithClipPath("clip-area"))
	)
	gp.Append(el)
	return gp.AsElement()
}

func (c Chart) drawAxis() svg.Element {
	g := svg.NewGroup(svg.WithID("axis"))
	if a, ok := c.Axis[PlacementBottom]; ok {
		ga := makeDomainLine(a, c.DrawingWidth(), 0, c.Padding.Left, c.Height-c.Padding.Bottom)
		if a.Label != "" {
			posx := c.DrawingWidth() + float64(len(a.Label)*5)
			opts := []svg.Option{
				svg.WithPosition(posx/2, c.Padding.Bottom*0.8),
				svg.WithFont(font),
				svg.WithAnchor("middle"),
				svg.WithFill(svg.NewFill(a.Color)),
				svg.WithClass("axis-title"),
			}
			txt := svg.NewText(a.Label, opts...)
			ga.Append(txt.AsElement())
		}
		gx := makeTicks(ga, a, c.DrawingWidth(), c.DrawingHeight())
		g.Append(gx)
	}
	if a, ok := c.Axis[PlacementLeft]; ok {
		ga := makeDomainLine(a, 0, c.DrawingHeight(), c.Padding.Left, c.Padding.Top)
		if a.Label != "" {
			posy := c.DrawingHeight() + 10
			opts := []svg.Option{
				svg.WithPosition(-c.Padding.Left*0.7, posy/2),
				svg.WithRotate(-90, -c.Padding.Left*0.7, posy/2),
				svg.WithFont(font),
				svg.WithAnchor("middle"),
				svg.WithFill(svg.NewFill(a.Color)),
				svg.WithClass("axis-title"),
			}
			txt := svg.NewText(a.Label, opts...)
			ga.Append(txt.AsElement())
		}
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
		if a.WithOuterTicks && i > 0 && i < len(values)-1 {
			line := outerTickLine(a, size)
			g.Append(line.AsElement())
		}
		if a.WithTickLabel {
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
		stroke = svg.NewStroke(a.Color, 1)
	)
	stroke.DashArray(5)
	return svg.NewLine(pos1, pos2, stroke.Option())
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
		stroke = svg.NewStroke(a.Color, 1)
	)
	return svg.NewLine(pos1, pos2, svg.WithStroke(stroke), svg.WithClass("axis-tick"))
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
		svg.WithFill(svg.NewFill(a.Color)),
		svg.WithPosition(point.X, point.Y),
		svg.WithDominantBaseline(base),
		svg.WithAnchor(anchor),
		svg.WithClass("axis-tick-text"),
	}
	return svg.NewText(a.Format(value), options...)
}

func (c Chart) drawArea() svg.Group {
	gos := []svg.Option{
		svg.WithID("area"),
		svg.WithTranslate(c.Padding.Left, c.Padding.Top),
	}
	g := svg.NewGroup(gos...)

	var (
		defs svg.Defs
		clip = svg.NewClipPath(svg.WithID("clip-area"))
		rec  = svg.NewRect(svg.WithDimension(c.DrawingWidth(), c.DrawingHeight()))
	)
	clip.Append(rec.AsElement())
	defs.Append(clip.AsElement())
	g.Append(defs.AsElement())

	if c.Area != "" {
		rec = svg.NewRect(svg.WithDimension(c.DrawingWidth(), c.DrawingHeight()))
		rec.Fill = svg.NewFill(c.Area)
		g.Append(rec.AsElement())
	}

	return g
}

func makeDomainLine(a Axis, x, y, left, top float64) svg.Group {
	var (
		ga = svg.NewGroup(svg.WithTranslate(left, top))
		sk = svg.NewStroke(a.Color, 1)
		os = []svg.Option{
			svg.WithStroke(sk),
			svg.WithClass("axis-domain"),
		}
		li = svg.NewLine(svg.NewPos(0, 0), svg.NewPos(x, y), os...)
	)
	ga.Append(li.AsElement())
	return ga
}

type Point struct {
	X    float64
	Y    float64
	Show bool
}

func NewPoint(x, y float64) Point {
	return Point{
		X: x,
		Y: y,
	}
}

type Line interface {
	Draw(Serie, Axis, Axis) svg.Element
}

type LineFunc func(float64, float64) Line

func getBasePath(serie Serie) svg.Path {
	opts := []svg.Option{
		svg.WithFill(svg.NewFill("none")),
		svg.WithStroke(svg.NewStroke(serie.Color, 2)),
		svg.WithRendering("geometricPrecision"),
		svg.WithClipPath("clip-area"),
	}
	return svg.NewPath(opts...)
}

type cubicCurve struct {
	width   float64
	height  float64
	stretch float64
}

func Cubic(w, h, s float64) Line {
	return cubicCurve{
		width:   w,
		height:  h,
		stretch: s,
	}
}

func (c cubicCurve) Draw(serie Serie, xaxis, yaxis Axis) svg.Element {
	var (
		off  = NewPoint(c.width/xaxis.Diff(), c.height/yaxis.Diff())
		fstx = slices.Fst(xaxis.Domain)
		fsty = slices.Fst(yaxis.Domain)
		pat  = getBasePath(serie)
		pos  svg.Pos
	)
	for i, pt := range serie.Values {
		pos.X = (pt.X - fstx) * off.X
		pos.Y = c.height - ((pt.Y - fsty) * off.Y)

		if i == 0 {
			pat.AbsMoveTo(pos)
		} else {
			pat.AbsLineTo(pos)
		}
	}
	return pat.AsElement()
}

type linearCurve struct {
	width  float64
	height float64
}

func Linear(w, h float64) Line {
	return linearCurve{
		width:  w,
		height: h,
	}
}

func (c linearCurve) Draw(serie Serie, xaxis, yaxis Axis) svg.Element {
	var (
		off  = NewPoint(c.width/xaxis.Diff(), c.height/yaxis.Diff())
		fstx = slices.Fst(xaxis.Domain)
		fsty = slices.Fst(yaxis.Domain)
		pat  = getBasePath(serie)
		pos  svg.Pos
	)
	for i, pt := range serie.Values {
		pos.X = (pt.X - fstx) * off.X
		pos.Y = c.height - ((pt.Y - fsty) * off.Y)

		if i == 0 {
			pat.AbsMoveTo(pos)
		} else {
			pat.AbsLineTo(pos)
		}
	}
	return pat.AsElement()
}

type stepCurve struct {
	width  float64
	height float64
}

func Step(w, h float64) Line {
	return stepCurve{
		width:  w,
		height: h,
	}
}

func (c stepCurve) Draw(serie Serie, xaxis, yaxis Axis) svg.Element {
	var (
		off  = NewPoint(c.width/xaxis.Diff(), c.height/yaxis.Diff())
		fstx = slices.Fst(xaxis.Domain)
		fsty = slices.Fst(yaxis.Domain)
		pat  = getBasePath(serie)
		pos  svg.Pos
		ori  svg.Pos
	)
	pos.X = (slices.Fst(serie.Values).X - fstx) * off.X
	pos.Y = c.height - (slices.Fst(serie.Values).Y-fsty)*off.Y
	pat.AbsMoveTo(pos)
	ori = pos
	for _, pt := range slices.Rest(serie.Values) {
		pos.X = (pt.X - fstx) * off.X
		pos.Y = c.height - ((pt.Y - fsty) * off.Y)

		ori.X += (pos.X - ori.X) / 2
		pat.AbsLineTo(ori)
		ori.Y = pos.Y
		pat.AbsLineTo(ori)
		pat.AbsLineTo(pos)
		ori = pos
	}
	return pat.AsElement()
}

type stepAfterCurve struct {
	width  float64
	height float64
}

func StepAfter(w, h float64) Line {
	return stepAfterCurve{
		width:  w,
		height: h,
	}
}

func (c stepAfterCurve) Draw(serie Serie, xaxis, yaxis Axis) svg.Element {
	var (
		off  = NewPoint(c.width/xaxis.Diff(), c.height/yaxis.Diff())
		fstx = slices.Fst(xaxis.Domain)
		fsty = slices.Fst(yaxis.Domain)
		pat  = getBasePath(serie)
		pos  svg.Pos
		ori  svg.Pos
	)
	pos.X = (slices.Fst(serie.Values).X - fstx) * off.X
	pos.Y = c.height - (slices.Fst(serie.Values).Y-fsty)*off.Y
	pat.AbsMoveTo(pos)
	ori = pos
	for _, pt := range slices.Rest(serie.Values) {
		pos.X = (pt.X - fstx) * off.X
		pos.Y = c.height - ((pt.Y - fsty) * off.Y)

		ori.X = pos.X
		pat.AbsLineTo(ori)
		ori.Y = pos.Y
		pat.AbsLineTo(ori)
		pat.AbsLineTo(pos)
		ori = pos
	}
	return pat.AsElement()
}

type stepBeforeCurve struct {
	width  float64
	height float64
}

func StepBefore(w, h float64) Line {
	return stepBeforeCurve{
		width:  w,
		height: h,
	}
}

func (c stepBeforeCurve) Draw(serie Serie, xaxis, yaxis Axis) svg.Element {
	var (
		off  = NewPoint(c.width/xaxis.Diff(), c.height/yaxis.Diff())
		fstx = slices.Fst(xaxis.Domain)
		fsty = slices.Fst(yaxis.Domain)
		pat  = getBasePath(serie)
		pos  svg.Pos
		ori  svg.Pos
	)
	pos.X = (slices.Fst(serie.Values).X - fstx) * off.X
	pos.Y = c.height - (slices.Fst(serie.Values).Y-fsty)*off.Y
	pat.AbsMoveTo(pos)
	ori = pos
	for _, pt := range slices.Rest(serie.Values) {
		pos.X = (pt.X - fstx) * off.X
		pos.Y = c.height - ((pt.Y - fsty) * off.Y)

		ori.Y = pos.Y
		pat.AbsLineTo(ori)
		ori.X = pos.X
		pat.AbsLineTo(ori)
		pat.AbsLineTo(pos)
		ori = pos
	}
	return pat.AsElement()
}

type Serie struct {
	Label  string
	Values []Point

	Curve LineFunc

	Color string
	Line
}

func (s *Serie) Add(x, y float64) {
	s.Values = append(s.Values, NewPoint(x, y))
}

func (s *Serie) Domain() (Point, Point) {
	var (
		fst Point
		lst Point
		tmp = make([]Point, len(s.Values))
	)
	copy(tmp, s.Values)
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].X < tmp[j].X
	})
	fst.X = slices.Fst(tmp).X
	lst.X = slices.Lst(tmp).Y
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Y < tmp[j].Y
	})
	fst.Y = slices.Fst(tmp).Y
	lst.Y = slices.Lst(tmp).Y
	return fst, lst
}

const (
	defaultWidth  = 800
	defaultHeight = 600
)

func main() {
	pad := NewPadding(10, 10, 50, 50)

	var ser1 Serie
	ser1.Label = "serie 1 (blue)"
	ser1.Color = "blue"
	ser1.Curve = StepBefore
	ser1.Add(-3, -3)
	ser1.Add(-1, -9)
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
	ser2.Color = "red"
	ser2.Curve = StepAfter
	ser2.Add(-6, 5)
	ser2.Add(3, 0)
	ser2.Add(15, 6)
	ser2.Add(28, 19)
	ser2.Add(37, 15)
	ser2.Add(40, 11)
	ser2.Add(56, 4)
	ser2.Add(61, 1)
	ser2.Add(67, 7)
	ser2.Add(69, 8)

	var ser3 Serie
	ser3.Label = "serie 3 (green)"
	ser3.Color = "green"
	ser3.Curve = Step
	ser3.Add(-7, -10)
	ser3.Add(-5, 23)
	ser3.Add(-1, 19)
	ser3.Add(13, 14)
	ser3.Add(32, 17)
	ser3.Add(41, 16)
	ser3.Add(47, 4)
	ser3.Add(51, -4)
	ser3.Add(56, 4)
	ser3.Add(64, -1)
	ser3.Add(69, -8)

	var (
		ch  = DefaultChart()
		xax = CreateAxis(-7, 69)
		yax = CreateAxis(-13, 28)
		format = func(v float64) string {
			return strconv.FormatFloat(v, 'f', 2, 64)
		}
	)
	yax.Ticks = 7
	yax.Label = "y-axis label"
	yax.Color = "white"
	yax.Format = format
	xax.Ticks = 7
	xax.Label = "x-axis label"
	xax.Color = "white"
	xax.WithOuterTicks = false
	xax.Format = format

	ch.Width = defaultWidth
	ch.Height = defaultHeight
	ch.Title = "sample chart"
	ch.Padding = pad
	ch.Legend.Align = AlignTop
	ch.Area = "black"
	ch.Background = "black"
	ch.AddAxis(PlacementBottom, xax)
	ch.AddAxis(PlacementLeft, yax)
	ch.Render(os.Stdout, []Serie{ser1, ser2, ser3})
	fmt.Println()
}
