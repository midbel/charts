package main

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

const FontSize = 12.0

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
	Range() float64
}

type TimeAxis struct {
	Label  string
	Rotate float64
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
	for _, t := range a.Values() {
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

func (a TimeAxis) Values() []time.Time {
	if len(a.Domain.Values) > 0 || a.Ticks == 0 {
		return a.Domain.Values
	}
	var (
		diff = a.Domain.Last.Sub(a.Domain.First)
		step = diff / time.Duration(a.Ticks)
	)
	var vs []time.Time
	for i := 0; i < a.Ticks; i++ {
		w := a.Domain.First.Add(step * time.Duration(i))
		vs = append(vs, w)
	}
	return vs
}

type NumberAxis struct {
	Label  string
	Rotate float64
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
	for _, f := range a.Values() {
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

func (a NumberAxis) Values() []float64 {
	if len(a.Domain.Values) > 0 || a.Ticks == 0 {
		return a.Domain.Values
	}
	var (
		diff = a.Domain.Last - a.Domain.First
		step = diff / float64(a.Ticks)
	)
	var vs []float64
	for i := 0; i < a.Ticks; i++ {
		w := a.Domain.First + (step * float64(i))
		vs = append(vs, w)
	}
	return vs
}

type CategoryAxis struct {
	Label  string
	Rotate float64
	Orientation
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
	for i, s := range a.Values() {
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

func (a CategoryAxis) Values() []string {
	return a.Domain
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

type Padding struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

func (p Padding) Horizontal() float64 {
	return p.Left + p.Right
}

func (p Padding) Vertical() float64 {
	return p.Top + p.Bottom
}

type Context struct {
	Width  float64
	Height float64
}

type Chart struct {
	Title  string
	Width  float64
	Height float64
	Padding

	Legend struct {
		Title string
		Align Alignment
	}

	Left   Axis
	Bottom Axis
	Right  Axis
	Top    Axis

	Background string
	Area       string
}

func (c Chart) DrawingWidth() float64 {
	return c.Width - c.Padding.Horizontal()
}

func (c Chart) DrawingHeight() float64 {
	return c.Height - c.Padding.Vertical()
}

func (c Chart) Render(w io.Writer, series ...Serie) {
	var (
		el = svg.NewSVG(svg.WithDimension(c.Width, c.Height))
		cv = c.getCanvas()
	)
	for _, s := range series {
		sr := s.Render(c.DrawingWidth(), c.DrawingHeight())
		cv.Append(sr)
	}
	el.Append(cv.AsElement())

	bw := bufio.NewWriter(w)
	defer bw.Flush()
	el.Render(bw)
}

func (c Chart) getCanvas() svg.Group {
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

	return g
}

type Serie interface {
	Render(float64, float64) svg.Element
}

type PointConstraint interface {
	~float64 | ~string | time.Time
}

type Point[T PointConstraint, U PointConstraint] struct {
	Fst T
	Lst U
	Rev bool
}

func TimePoint(x time.Time, y float64) Point[time.Time, float64] {
	return Point[time.Time, float64]{
		Fst: x,
		Lst: y,
	}
}

func NumberPoint(x, y float64) Point[float64, float64] {
	return Point[float64, float64]{
		Fst: x,
		Lst: y,
	}
}

func CategoryPoint(x string, y float64) Point[string, float64] {
	return Point[string, float64]{
		Fst: x,
		Lst: y,
	}
}

func getBasePath(stroke string) svg.Path {
	pat := svg.NewPath()
	pat.Stroke = svg.NewStroke(stroke, 1)
	return pat
}

func getCircle(pos svg.Pos, fill string) svg.Circle {
	ci := svg.NewCircle()
	ci.Radius = 5
	ci.Pos = pos
	ci.Fill = svg.NewFill(fill)
	return ci
}

type NumberSerie struct {
	Title     string
	Stroke    string
	WithPoint bool
	Values    []Point[float64, float64]

	X Axis
	Y Axis
}

func (s NumberSerie) Render(width, height float64) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(s.Stroke)
		off = NumberPoint(width/s.X.Range(), height/s.Y.Range())
		fst = NumberPoint(0, 0)
		pos svg.Pos
	)

	for i, pt := range s.Values {
		pos.X = (pt.Fst - fst.Fst) * off.Fst
		pos.Y = height - ((pt.Lst - fst.Lst) * off.Lst)
		if i == 0 {
			pat.AbsMoveTo(pos)
		} else {
			pat.AbsLineTo(pos)
		}
		if s.WithPoint {
			ci := getCircle(pos, s.Stroke)
			grp.Append(ci.AsElement())
		}
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

type TimeSerie struct {
	Title  string
	Stroke string
	Values []Point[time.Time, float64]

	X Axis
	Y Axis
}

func (s TimeSerie) Render(width, height float64) svg.Element {
	g := svg.NewGroup()
	return g.AsElement()
}

type CategorySerie struct {
	Title  string
	Stroke string
	Values []Point[string, float64]
}

func (s CategorySerie) Render(width, height float64) svg.Element {
	g := svg.NewGroup()
	return g.AsElement()
}

func main() {
	orient := OrientTop
	area := svg.NewSVG(svg.WithDimension(800, 800))
	cat := CategoryAxis{
		Orientation: orient,
		Domain:      []string{"go", "python", "javascript", "rust", "c++"},
	}
	elem := cat.Render(720, 40, 40)
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
	elem = tim.Render(720, 40, 240)
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

	elem = num.Render(720, 40, 440)
	area.Append(elem)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	area.Render(w)
}
