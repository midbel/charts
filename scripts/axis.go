package main

import (
	"bufio"
	"os"
	"strconv"
	"time"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

type ScalerConstraint interface {
	~float64 | ~string | time.Time
}

type Domain[T ScalerConstraint] interface {
	Diff(T) float64
	Extend() float64
	Values(int) []T
}

type numberDomain struct {
	fst float64
	lst float64
}

func NumberDomain(f, t float64) Domain[float64] {
	return numberDomain{
		fst: f,
		lst: t,
	}
}

func (n numberDomain) Diff(v float64) float64 {
	return v - n.fst
}

func (n numberDomain) Extend() float64 {
	return n.lst - n.fst
}

func (n numberDomain) Values(c int) []float64 {
	var (
		all  = make([]float64, c)
		step = n.Extend() / float64(c)
	)
	for i := 0; i < c; i++ {
		all[i] = n.fst + float64(i+1)*step
	}
	all = append(all, n.lst)
	return all
}

type timeDomain struct {
	fst time.Time
	lst time.Time
}

func TimeDomain(f, t time.Time) Domain[time.Time] {
	return timeDomain{
		fst: f,
		lst: t,
	}
}

func (t timeDomain) Diff(v time.Time) float64 {
	diff := v.Sub(t.fst)
	return float64(diff)
}

func (t timeDomain) Extend() float64 {
	diff := t.lst.Sub(t.fst)
	return float64(diff)
}

func (t timeDomain) Values(c int) []time.Time {
	var (
		all  = make([]time.Time, c)
		step = t.Extend() / float64(c)
	)
	for i := 0; i < c; i++ {
		all[i] = t.fst.Add(time.Duration(float64(i) * step))
	}
	all = append(all, t.lst)
	return all
}

type Range struct {
	F float64
	T float64
}

func NewRange(f, t float64) Range {
	return Range{
		F: f,
		T: t,
	}
}

func (r Range) Len() float64 {
	return r.T - r.F
}

type Scaler[T ScalerConstraint] interface {
	Scale(T) float64
	Space() float64
	Values(int) []T
}

type numberScaler struct {
	Range
	Domain[float64]
}

func NumberScaler(dom Domain[float64], rg Range) Scaler[float64] {
	return numberScaler{
		Range:  rg,
		Domain: dom,
	}
}

func (n numberScaler) Scale(v float64) float64 {
	return n.Diff(v) * n.Space()
}

func (n numberScaler) Space() float64 {
	return n.Len() / n.Extend()
}

type timeScaler struct {
	Range
	Domain[time.Time]
}

func TimeScaler(dom Domain[time.Time], rg Range) Scaler[time.Time] {
	return timeScaler{
		Range:  rg,
		Domain: dom,
	}
}

func (s timeScaler) Scale(v time.Time) float64 {
	return s.Diff(v) * s.Space()
}

func (s timeScaler) Space() float64 {
	return s.Len() / s.Extend()
}

type stringScaler struct {
	Range
	Strings []string
}

func StringScaler(str []string, rg Range) Scaler[string] {
	return stringScaler{
		Range:   rg,
		Strings: str,
	}
}

func (s stringScaler) Scale(v string) float64 {
	var x int
	for i := range s.Strings {
		if s.Strings[i] == v {
			x = i
			break
		}
	}
	return float64(x) * s.Space()
}

func (s stringScaler) Space() float64 {
	return s.Len() / float64(len(s.Strings))
}

func (s stringScaler) Values(c int) []string {
	return nil
}

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
	Range() float64
}

type TimeAxis struct {
	Label  string
	Rotate float64
	Orientation
	Ticks  int
	Scaler Scaler[time.Time]
	Domain []time.Time
	Format func(time.Time) string
}

func (a TimeAxis) Render(length, left, top float64) svg.Element {
	g := svg.NewGroup(svg.WithTranslate(left, top))
	d := domainLine(a.Orientation, length, svg.NewStroke("black", 1))
	g.Append(d.AsElement())

	var (
		data   = a.Domain
		font   = svg.NewFont(FontSize)
		format = a.Format
	)
	if len(data) == 0 {
		data = a.Scaler.Values(a.Ticks)
	}
	if format == nil {
		format = func(t time.Time) string {
			return t.Format("2006-01-02")
		}
	}
	for _, t := range data {
		var (
			pos  = a.Scaler.Scale(t)
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

type NumberAxis struct {
	Label  string
	Rotate float64
	Orientation
	Ticks  int
	Scaler Scaler[float64]
	Domain []float64
	Format func(float64) string
}

func (a NumberAxis) Render(length, left, top float64) svg.Element {
	g := svg.NewGroup(svg.WithTranslate(left, top))
	d := domainLine(a.Orientation, length, svg.NewStroke("black", 1))
	g.Append(d.AsElement())

	var (
		data   = a.Domain
		font   = svg.NewFont(FontSize)
		format = a.Format
	)
	if len(data) == 0 {
		data = a.Scaler.Values(a.Ticks)
	}
	if format == nil {
		format = func(f float64) string {
			return strconv.FormatFloat(f, 'f', 2, 64)
		}
	}
	for _, f := range data {
		var (
			pos  = a.Scaler.Scale(f)
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

func getBasePath(stroke string) svg.Path {
	pat := svg.NewPath()
	pat.Stroke = svg.NewStroke(stroke, 1)
	pat.Fill = svg.NewFill("none")
	return pat
}

func getCircle(pos svg.Pos, fill string) svg.Circle {
	ci := svg.NewCircle()
	ci.Radius = 5
	ci.Pos = pos
	ci.Fill = svg.NewFill(fill)
	return ci
}

type Point[T, U any] struct {
	X T
	Y U
}

func NumberPoint(x, y float64) Point[float64, float64] {
	return Point[float64, float64]{
		X: x,
		Y: y,
	}
}

func TimePoint(x time.Time, y float64) Point[time.Time, float64] {
	return Point[time.Time, float64]{
		X: x,
		Y: y,
	}
}

func (p Point[T, U]) Reverse() Point[U, T] {
	return Point[U, T]{
		X: p.Y,
		Y: p.X,
	}
}

type Serie[T, U ScalerConstraint] struct {
	WithPoint bool
	Color     string

	Points []Point[T, U]
	X      Scaler[T]
	Y      Scaler[U]

	Renderer RenderFunc[T, U]
}

func (s Serie[T, U]) Render() svg.Element {
	return s.Renderer(s)
}

type Renderer[T, U ScalerConstraint] interface {
	Render([]Point[T, U]) svg.Element
}

type RenderFunc[T, U ScalerConstraint] func(Serie[T, U]) svg.Element

func stepRender[T, U ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color)
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
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func linearRender[T, U ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color)
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
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func stepAfterRender[T, U ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color)
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
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func stepBeforeRender[T, U ScalerConstraint](serie Serie[T, U]) svg.Element {
	var (
		grp = svg.NewGroup()
		pat = getBasePath(serie.Color)
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
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func main() {
	var (
		size   = NewRange(0, 720)
		orient = OrientTop
		area   = svg.NewSVG(svg.WithDimension(800, 800))
	)
	cat := CategoryAxis{
		Orientation: orient,
		Domain:      []string{"go", "python", "javascript", "rust", "c++"},
	}
	elem := cat.Render(720, 40, 40)
	area.Append(elem)

	var (
		tim     TimeAxis
		dtstart = time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC)
		dtend   = time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)
	)
	tim.Ticks = 6
	tim.Orientation = orient
	tim.Scaler = TimeScaler(TimeDomain(dtstart, dtend), size)
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

	var num NumberAxis
	num.Orientation = orient
	num.Scaler = NumberScaler(NumberDomain(0, 130), size)
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
	serie1.Renderer = stepBeforeRender[float64, float64]
	serie1.Color = "blue"
	serie1.WithPoint = true
	serie1.X = NumberScaler(NumberDomain(0, 720), NewRange(0, 720))
	serie1.Y = NumberScaler(NumberDomain(720, 0), NewRange(0, 540))
	serie1.Points = []Point[float64, float64]{
		NumberPoint(100, 245),
		NumberPoint(150, 567),
		NumberPoint(324, 98),
		NumberPoint(461, 19),
		NumberPoint(511, 563),
		NumberPoint(541, 463),
		NumberPoint(571, 113),
		NumberPoint(591, 703),
		NumberPoint(645, 301),
		NumberPoint(716, 341),
	}

	pat := serie1.Render()
	area.Append(pat)

	dtstart = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	dtend = time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)

	var serie2 Serie[time.Time, float64]
	serie2.Renderer = stepAfterRender[time.Time, float64]
	serie2.Color = "red"
	serie2.WithPoint = true
	serie2.X = TimeScaler(TimeDomain(dtstart, dtend), NewRange(0, 720))
	serie2.Y = NumberScaler(NumberDomain(100, 0), NewRange(0, 540))
	serie2.Points = []Point[time.Time, float64]{
		TimePoint(time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC), 34),
		TimePoint(time.Date(2022, 1, 20, 0, 0, 0, 0, time.UTC), 39),
		TimePoint(time.Date(2022, 2, 10, 0, 0, 0, 0, time.UTC), 40),
		TimePoint(time.Date(2022, 2, 26, 0, 0, 0, 0, time.UTC), 45),
		TimePoint(time.Date(2022, 3, 7, 0, 0, 0, 0, time.UTC), 43),
		TimePoint(time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC), 43),
		TimePoint(time.Date(2022, 6, 11, 0, 0, 0, 0, time.UTC), 67),
		TimePoint(time.Date(2022, 6, 29, 0, 0, 0, 0, time.UTC), 80),
		TimePoint(time.Date(2022, 7, 6, 0, 0, 0, 0, time.UTC), 89),
		TimePoint(time.Date(2022, 9, 5, 0, 0, 0, 0, time.UTC), 98),
		TimePoint(time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC), 19),
		TimePoint(time.Date(2022, 11, 19, 0, 0, 0, 0, time.UTC), 98),
		TimePoint(time.Date(2022, 12, 6, 0, 0, 0, 0, time.UTC), 86),
		TimePoint(time.Date(2022, 12, 16, 0, 0, 0, 0, time.UTC), 54),
		TimePoint(time.Date(2022, 12, 25, 0, 0, 0, 0, time.UTC), 12),
		TimePoint(time.Date(2022, 12, 30, 0, 0, 0, 0, time.UTC), 1),
	}

	pat = serie2.Render()
	area.Append(pat)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	area.Render(w)
}
