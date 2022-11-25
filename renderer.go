package charts

import (
	"fmt"
	"html"
	"math"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

type PolarType int

const (
	PolarDefault PolarType = 1 << iota
	PolarArea
	PolarPolygon
)

type Renderer[T, U ScalerConstraint] interface {
	Render(Serie[T, U]) svg.Element
}

type PolarRenderer[T ~string, U ~float64] struct {
	Fill       []string
	Radius     float64
	Ticks      int
	TicksStyle LineStyle
	Type       PolarType
	Stacked    bool
	Angular    bool
	Normalize  bool
	Point      PointFunc
}

func (r PolarRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = classGroup("", "polar")
		el  svg.Element
	)
	grp.Transform = svg.Translate(serie.X.Max()/2, serie.Y.Max()/2)
	grp.Append(r.drawTicks(serie))
	if r.Stacked {
		el = r.drawStackedArcs(serie)
	} else {
		el = r.drawArcs(serie)
	}
	if el != nil {
		grp.Append(el)
	}
	return grp.AsElement()
}

func (r PolarRenderer[T, U]) drawStackedArcs(serie Serie[T, U]) svg.Element {
	var (
		angle = fullcircle / float64(len(serie.Points))
		scale = serie.Y.replace(NewRange(0, r.Radius))
		grp   svg.Group
	)
	for i, pt := range serie.Points {
		var (
			ori svg.Pos
			add float64
			pre float64
		)
		for j, p := range pt.Sub {
			var (
				pat  svg.Path
				fill = svg.NewFill(r.Fill[j%len(r.Fill)])
				y    float64
				f    float64
				ag1  = angle * float64(i) * deg2rad
				ag2  = angle * float64(i+1) * deg2rad
			)
			pre = add
			f = scale.Scale(U(pre))
			add += any(p.Y).(float64)
			y = scale.Scale(U(add))
			pat.Fill = fill
			pat.Fill.Opacity = 0.7
			pat.Stroke = svg.NewStroke("white", 1)

			pat.AbsMoveTo(ori)
			pat.AbsLineTo(getPosFromAngle(ag1, y))
			pat.AbsArcTo(getPosFromAngle(ag2, y), y, y, 0, false, true)
			pat.AbsLineTo(getPosFromAngle(ag2, f))
			pat.AbsArcTo(ori, f, f, 0, false, false)
			pat.ClosePath()

			ori = getPosFromAngle(ag1, y)

			grp.Append(pat.AsElement())
		}
	}
	return grp.AsElement()
}

func (r PolarRenderer[T, U]) drawArcs(serie Serie[T, U]) svg.Element {
	var (
		angle = fullcircle / float64(len(serie.Points))
		scale = serie.Y.replace(NewRange(0, r.Radius))
		grp   svg.Group
	)
	for i, pt := range serie.Points {
		var (
			pat  svg.Path
			fill = svg.NewFill(r.Fill[i%len(r.Fill)])
			ag1  = angle * float64(i) * deg2rad
			ag2  = angle * float64(i+1) * deg2rad
			y    = scale.Scale(pt.Y)
		)
		pat.Fill = fill
		pat.Fill.Opacity = 0.9
		pat.Stroke = svg.NewStroke("white", 2)
		pat.AbsMoveTo(svg.NewPos(0, 0))
		pat.AbsLineTo(getPosFromAngle(ag1, y))
		pat.AbsArcTo(getPosFromAngle(ag2, y), y, y, 0, false, true)
		pat.AbsLineTo(svg.NewPos(0, 0))
		pat.ClosePath()
		grp.Append(pat.AsElement())
	}
	return grp.AsElement()
}

func (r PolarRenderer[T, U]) drawArea(serie Serie[T, U]) svg.Element {
	var (
		angle = fullcircle / float64(len(serie.Points))
		scale = serie.Y.replace(NewRange(0, r.Radius))
		pg    svg.Polygon
		grp   svg.Group
	)
	pg.Fill = svg.NewFill("blue")
	pg.Fill.Opacity = 0.4
	pg.Stroke = svg.NewStroke("blue", 1.5)

	for i, pt := range serie.Points {
		var (
			ag  = angle * float64(i) * deg2rad
			pos = getPosFromAngle(ag, scale.Scale(pt.Y))
		)
		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
		pg.Points = append(pg.Points, pos)
	}
	grp.Append(pg.AsElement())
	return grp.AsElement()
}

func (r PolarRenderer[T, U]) drawTicks(serie Serie[T, U]) svg.Element {
	var (
		grp   svg.Group
		angle = fullcircle / float64(len(serie.Points))
		step  = r.Radius / float64(r.Ticks)
		sk    = svg.NewStroke("black", 1)
		defs  svg.Defs
	)
	sk.Opacity = 0.25
	switch r.TicksStyle {
	case StyleStraight:
	case StyleDotted:
		sk.DashArray = append(sk.DashArray, 1, 5)
	case StyleDashed:
		sk.DashArray = append(sk.DashArray, 10, 5)
	}
	grp.Append(defs.AsElement())
	for i := 0; i < len(serie.Points); i++ {
		var (
			id  = fmt.Sprintf("polar-tick-pat-%03d", i)
			ag  = angle * float64(i) * deg2rad
			pos = getPosFromAngle(ag, r.Radius)
			li  = svg.NewLine(svg.NewPos(0, 0), pos)
			txt = svg.NewTextPath(any(serie.Points[i].X).(string), id)
			pat svg.Path
		)

		pat.Id = id
		pat.Stroke = svg.NewStroke("green", 2)
		pat.AbsMoveTo(pos)
		pat.AbsArcTo(getPosFromAngle(angle*float64(i+1)*deg2rad, r.Radius), r.Radius, r.Radius, 0, false, true)

		defs.Append(pat.AsElement())

		txt.Font = svg.NewFont(FontSize)
		txt.Shift = svg.NewPos(FontSize*0.5, -FontSize*0.5)
		grp.Append(txt.AsElement())

		li.Stroke = sk
		grp.Append(li.AsElement())
	}
	for i := step; i <= r.Radius; i += step {
		if r.Angular {
			c := r.drawAngularTicks(len(serie.Points), i, sk)
			grp.Append(c)
		} else {
			c := r.drawCircularTicks(float64(i), sk)
			grp.Append(c)
		}
	}
	return grp.AsElement()
}

func (r PolarRenderer[T, U]) drawAngularTicks(n int, radius float64, stroke svg.Stroke) svg.Element {
	var (
		pg svg.Polygon
		ag = fullcircle / float64(n)
	)
	pg.Stroke = stroke
	pg.Fill = svg.NewFill("none")
	for i := 0; i < n; i++ {
		a := ag * float64(i) * deg2rad
		pg.Points = append(pg.Points, getPosFromAngle(a, radius))
	}
	return pg.AsElement()
}

func (r PolarRenderer[T, U]) drawCircularTicks(radius float64, stroke svg.Stroke) svg.Element {
	var ci svg.Circle
	ci.Radius = radius
	ci.Stroke = stroke
	ci.Fill = svg.NewFill("none")
	return ci.AsElement()
}

type SunburstRenderer[T ~string, U ~float64] struct {
	Style
	Fill        []string
	InnerRadius float64
	OuterRadius float64
}

func (r SunburstRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.InnerRadius <= 0 {
		r.InnerRadius = r.OuterRadius
	}
	if len(r.Fill) == 0 {
		r.Fill = Tableau10
	}
	var (
		grp    = classGroup("sun")
		height = (r.OuterRadius - r.InnerRadius) / float64(serie.Depth())
		frac   = fullcircle / serie.Sum()
		offset float64
	)
	grp.Transform = svg.Translate(serie.X.Max()/2, serie.Y.Max()/2)
	for i, pt := range serie.Points {
		var (
			g svg.Group
			f = svg.NewFill(r.Fill[i%len(r.Fill)])
		)
		g.Id = any(pt.X).(string)

		r.drawPoints(&g, f, pt, offset, frac, 0, height)
		grp.Append(g.AsElement())
		offset += any(pt.Y).(float64) * frac
	}
	return grp.AsElement()
}

func (r SunburstRenderer[T, U]) drawPoints(grp *svg.Group, fill svg.Fill, pt Point[T, U], offset, frac, level, height float64) {
	var (
		value    = any(pt.Y).(float64) * frac
		distance = r.distanceFromCenter() + (height * level) + height
		pos1     = getPosFromAngle(offset*deg2rad, distance)
		pos2     = getPosFromAngle((offset+value)*deg2rad, distance)
		pos3     = getPosFromAngle((offset+value)*deg2rad, distance-height)
		pos4     = getPosFromAngle(offset*deg2rad, distance-height)
		pat      svg.Path
	)

	pat.Fill = fill
	pat.Rendering = "geometricPrecision"
	pat.Stroke = svg.NewStroke("white", 2)

	pat.AbsMoveTo(pos1)
	pat.AbsArcTo(pos2, distance, distance, 0, value > halfcircle, true)
	pat.AbsLineTo(pos3)

	if pos3.X != pos4.X && pos3.Y != pos4.Y {
		pat.AbsArcTo(pos4, distance-height, distance-height, 0, value > halfcircle, false)
	}
	pat.AbsLineTo(pos1)
	grp.Append(pat.AsElement())

	level += 1
	if pt.isLeaf() {
		return
	}
	for _, p := range pt.Sub {
		sub := frac * any(p.Y).(float64)
		r.drawPoints(grp, fill, p, offset, frac, level, height)
		offset += sub
	}
}

func (r SunburstRenderer[T, U]) distanceFromCenter() float64 {
	if r.OuterRadius == r.InnerRadius {
		return 0
	}
	return r.InnerRadius
}

type PieRenderer[T ~string, U ~float64] struct {
	Style
	Fill        []string
	InnerRadius float64
	OuterRadius float64
	Text        TextPosition
}

func (r PieRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.InnerRadius <= 0 {
		r.InnerRadius = r.OuterRadius
	}
	if len(r.Fill) == 0 {
		r.Fill = Tableau10
	}
	var (
		grp   = classGroup("pie")
		part  = fullcircle / serie.Sum()
		angle float64
	)
	grp.Transform = svg.Translate(serie.X.Max()/2, serie.Y.Max()/2)
	for i, pt := range serie.Points {
		var (
			rad  = angle * deg2rad
			val  = any(pt.Y).(float64) * part
			pos3 = r.getPos3(angle, val)
			pos4 = r.getPos4(rad)
			pat  svg.Path
		)
		pat.Fill = svg.NewFill(r.Fill[i%len(r.Fill)])
		pat.Rendering = "geometricPrecision"
		pat.Stroke = svg.NewStroke("white", 2)
		if s, ok := any(pt.X).(string); ok {
			pat.Title = html.EscapeString(s)
		}

		pat.AbsMoveTo(r.getPos1(rad))
		pat.AbsArcTo(r.getPos2(angle, val), r.OuterRadius, r.OuterRadius, 0, val > halfcircle, true)
		pat.AbsLineTo(pos3)
		if pos3.X != pos4.X && pos3.Y != pos4.Y {
			pat.AbsArcTo(pos4, r.difference(), r.difference(), 0, val > halfcircle, false)
		}
		pat.AbsLineTo(r.getPos1(rad))
		pat.ClosePath()
		grp.Append(pat.AsElement())

		angle += val

		switch r.Text {
		case TextAfter:
		case TextCenter:
		default:
		}

	}
	return grp.AsElement()
}

func (r PieRenderer[T, U]) getPos4(rad float64) svg.Pos {
	return getPosFromAngle(rad, r.difference())
}

func (r PieRenderer[T, U]) getPos3(angle, rad float64) svg.Pos {
	return getPosFromAngle((angle+rad)*deg2rad, r.difference())
}

func (r PieRenderer[T, U]) getPos2(angle, rad float64) svg.Pos {
	return getPosFromAngle((angle+rad)*deg2rad, r.OuterRadius)
}

func (r PieRenderer[T, U]) getPos1(rad float64) svg.Pos {
	return getPosFromAngle(rad, r.OuterRadius)
}

func (r PieRenderer[T, U]) difference() float64 {
	return r.OuterRadius - r.InnerRadius
}

type GroupRenderer[T ~string, U float64] struct {
	Style
	Fill  []string
	Width float64
}

func (r GroupRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.Width <= 0 {
		r.Width = 1
	}
	if len(r.Fill) == 0 {
		r.Fill = Tableau10
	}
	var (
		grp = classGroup("bar")
		sub = serie.X.replace(NewRange(0, serie.X.Space()))
	)
	for _, pt := range serie.Points {
		g := classGroup("group", "bar-group")
		g.Transform = svg.Translate(serie.X.Scale(pt.X), 0)

		if r, ok := sub.(scalerReset[T]); ok {
			var dat []T
			for _, s := range pt.Sub {
				dat = append(dat, s.X)
			}
			sub = r.reset(dat)
		}
		for i, s := range pt.Sub {
			el := getRect(s, sub, serie.Y, r.Width, r.Fill[i%len(r.Fill)])
			g.Append(el)
		}
		grp.Append(g.AsElement())
	}
	return grp.AsElement()
}

type StackedRenderer[T ~string, U ~float64] struct {
	Style
	Width     float64
	Normalize bool
}

func (r StackedRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.Width <= 0 {
		r.Width = 1
	}
	var (
		grp  svg.Group
		pal  = r.FillList.Clone()
		max  = serie.Y.Max()
		size = serie.X.Space()
	)
	for _, parent := range serie.Points {
		r.FillList = pal.Clone()
		var (
			offset float64
			bar    = classGroup("bar", "bar-stack")
		)
		bar.Transform = svg.Translate(serie.X.Scale(parent.X), 0)
		for _, pt := range parent.Sub {
			if r.Normalize {
				pt.Y = pt.Y / parent.Y
			}
			var (
				val = serie.Y.Scale(pt.Y)
				wid = size * r.Width
				off = (size - wid) / 2
				rec = r.Rect(wid, max-val)
			)
			rec.Pos = svg.NewPos(off, val-offset)
			bar.Append(rec.AsElement())

			offset += max - val
		}
		grp.Append(bar.AsElement())
	}
	return grp.AsElement()
}

type BarRenderer[T ~string, U ~float64] struct {
	Style
	Width float64
}

func (r BarRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.Width <= 0 {
		r.Width = 1
	}
	grp := classGroup("bar")
	for _, pt := range serie.Points {
		var (
			width  = serie.X.Space() * r.Width
			height = serie.Y.Max() - serie.Y.Scale(pt.Y)
			offset = (serie.X.Space() - width) / 2
			rec    = r.Rect(width, height)
		)
		rec.Pos = svg.NewPos(serie.X.Scale(pt.X)+offset, serie.Y.Scale(pt.Y))
		grp.Append(rec.AsElement())
	}
	return grp.AsElement()
}

type PointRenderer[T, U ScalerConstraint] struct {
	Fill  string
	Point PointFunc
}

func (r PointRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.Point == nil {
		r.Point = GetCircle
	}
	grp := classGroup("scatter")
	for _, pt := range serie.Points {
		var (
			x = serie.X.Scale(pt.X)
			y = serie.Y.Scale(pt.Y)
		)
		grp.Append(r.Point(svg.NewPos(x, y)))
	}
	return grp.AsElement()
}

// type CubicRenderer[T, U ScalerConstraint] struct {
// 	Style
// 	Stretch float64
// 	Fill    bool
// 	Point   PointFunc
// }

// func (r CubicRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
// 	var (
// 		grp = classGroup("line")
// 		pat = r.Path()
// 		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
// 		ori svg.Pos
// 	)
// 	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
// 	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
// 	pat.AbsMoveTo(pos)
// 	if r.Point != nil {
// 		grp.Append(r.Point(pos))
// 	}
// 	ori = pos
// 	for _, pt := range slices.Rest(serie.Points) {
// 		pos.X = serie.X.Scale(pt.X)
// 		pos.Y = serie.Y.Scale(pt.Y)

// 		var (
// 			ctrl1 = ori
// 			ctrl2 = pos
// 			diff  = (pos.X - ori.X) * r.Stretch
// 		)
// 		ctrl1.X += diff
// 		ctrl2.X -= diff

// 		pat.AbsCubicCurve(pos, ctrl1, ctrl2)
// 		ori = pos
// 		if r.Point != nil {
// 			grp.Append(r.Point(pos))
// 		}
// 	}
// 	grp.Append(pat.AsElement())
// 	return grp.AsElement()
// }

type CurveType int

const (
	CurveLine CurveType = 1 << iota
	CurveStep
	CurveBefore
	CurveAfter
	CurveCubic
)

func (c CurveType) Classname() []string {
	switch c {
	default:
		return nil
	case CurveLine:
		return []string{"line"}
	case CurveStep:
		return []string{"line", "line-step"}
	case CurveBefore:
		return []string{"line", "line-step", "step-before"}
	case CurveAfter:
		return []string{"line", "line-step", "step-after"}
	case CurveCubic:
		return []string{"line", "line-cubic"}
	}
}

type AreaRenderer[T, U ScalerConstraint] struct {
	LinearRenderer[T, U]
}

func Area[T, U ScalerConstraint]() AreaRenderer[T, U] {
	return AreaRenderer[T, U]{
		LinearRenderer: Line[T, U](),
	}
}

func (r AreaRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = classGroup(r.Type.Classname()...)
		pat = r.renderLine(serie, true)
	)
	var (
		lst = slices.Lst(serie.Points)
		pos = svg.NewPos(serie.X.Scale(lst.X), serie.Y.Max())
	)
	pat.AbsLineTo(pos)

	if len(r.FillList) == 0 {
		r.FillList = append(r.FillList, r.LineColor)
	}
	pat.Fill = svg.NewFill(slices.Fst(r.FillList))
	pat.Fill.Opacity = r.FillOpacity
	grp.Append(pat.AsElement())
	if el := r.renderText(serie); el != nil {
		grp.Append(el)
	}
	return grp.AsElement()
}

type LinearRenderer[T, U ScalerConstraint] struct {
	Style
	Text          TextPosition
	IgnoreMissing bool
	Type          CurveType
}

func Line[T, U ScalerConstraint]() LinearRenderer[T, U] {
	return createLinearRenderer[T, U](CurveLine)
}

func Step[T, U ScalerConstraint]() LinearRenderer[T, U] {
	return createLinearRenderer[T, U](CurveStep)
}

func StepBefore[T, U ScalerConstraint]() LinearRenderer[T, U] {
	return createLinearRenderer[T, U](CurveBefore)
}

func StepAfter[T, U ScalerConstraint]() LinearRenderer[T, U] {
	return createLinearRenderer[T, U](CurveAfter)
}

func createLinearRenderer[T, U ScalerConstraint](curve CurveType) LinearRenderer[T, U] {
	return LinearRenderer[T, U]{
		Type:  curve,
		Style: DefaultStyle(),
	}
}

func (r LinearRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = classGroup(r.Type.Classname()...)
		pat svg.Path
	)
	switch r.Type {
	case CurveLine:
		pat = r.renderLine(serie, false)
	case CurveStep:
		pat = r.renderStep(serie, false)
	case CurveBefore:
		pat = r.renderStepBefore(serie, false)
	case CurveAfter:
		pat = r.renderStepAfter(serie, false)
	default:
		return grp.AsElement()
	}
	grp.Append(pat.AsElement())
	if el := r.renderText(serie); el != nil {
		grp.Append(el)
	}
	return grp.AsElement()
}

func (r LinearRenderer[T, U]) renderLine(serie Serie[T, U], zero bool) svg.Path {
	var (
		pat = r.Path()
		pos svg.Pos
		nan bool
	)
	if zero {
		fst := slices.Fst(serie.Points)
		pos := svg.NewPos(serie.X.Scale(fst.X), serie.Y.Max())
		pat.AbsMoveTo(pos)
	}
	for i, pt := range serie.Points {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)
		if (i == 0 && !zero) || (nan && r.IgnoreMissing) {
			nan = false
			pat.AbsMoveTo(pos)
		} else {
			pat.AbsLineTo(pos)
		}
	}
	return pat
}

func (r LinearRenderer[T, U]) renderStep(serie Serie[T, U], zero bool) svg.Path {
	var (
		pat = r.Path()
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
		nan bool
	)

	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pat.AbsLineTo(pos)

	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)
		if nan && r.IgnoreMissing {
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
	}
	return pat
}

func (r LinearRenderer[T, U]) renderStepAfter(serie Serie[T, U], zero bool) svg.Path {
	var (
		pat = r.Path()
		pos svg.Pos
		ori svg.Pos
		nan bool
	)

	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Max()
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)

	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		if nan && r.IgnoreMissing {
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
	}
	return pat
}

func (r LinearRenderer[T, U]) renderStepBefore(serie Serie[T, U], zero bool) svg.Path {
	var (
		pat = r.Path()
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

	ori = pos
	for _, pt := range slices.Rest(serie.Points) {
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		if nan && r.IgnoreMissing {
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
	}
	return pat
}

func (r LinearRenderer[T, U]) renderText(serie Serie[T, U]) svg.Element {
	switch txt := r.Style.Text(serie.Title); r.Text {
	case TextBefore:
		pt := slices.Fst(serie.Points)
		return getText(txt, 0, serie.Y.Scale(pt.Y), true)
	case TextAfter:
		pt := slices.Lst(serie.Points)
		return getText(txt, serie.X.Scale(pt.X), serie.Y.Scale(pt.Y), false)
	default:
		return nil
	}
}

func getText(txt svg.Text, x, y float64, before bool) svg.Element {
	txt.Pos = svg.NewPos(x, y)
	if !before {
		txt.Anchor = "start"
		txt.Pos.X += txt.Font.Size * 0.4
	} else {
		txt.Anchor = "end"
		txt.Pos.X -= txt.Font.Size * 0.4
	}
	return txt.AsElement()
}

func getRect[T, U ScalerConstraint](pt Point[T, U], x Scaler[T], y Scaler[U], ratio float64, fill string) svg.Element {
	var (
		width  = x.Space() * ratio
		offset = (x.Space() - width) / 2
		pos    = svg.NewPos(x.Scale(pt.X)+offset, y.Scale(pt.Y))
		dim    = svg.NewDim(width, y.Max()-pos.Y)
	)
	var el svg.Rect
	el.Pos = pos
	el.Dim = dim
	el.Fill = svg.NewFill(fill)
	return el.AsElement()
}

const (
	fullcircle = 360.0
	halfcircle = 180.0
	deg2rad    = math.Pi / halfcircle
	rad2deg    = halfcircle / math.Pi
)

func getPosFromAngle(angle, radius float64) svg.Pos {
	var (
		x1 = float64(radius) * math.Cos(angle)
		y1 = float64(radius) * math.Sin(angle)
	)
	return svg.NewPos(x1, y1)
}

func isFloat[T any](v T) (float64, bool) {
	x, ok := any(v).(float64)
	return x, ok
}
