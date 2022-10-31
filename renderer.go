package charts

import (
	"fmt"
	"html"
	"math"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

type TextPosition int

const (
	TextBefore TextPosition = 1 << iota
	TextAfter
	TextCenter
)

type LineStyle int

const (
	StyleStraight LineStyle = 1 << iota
	StyleDotted
	StyleDashed
)

type PolarType int

const (
	PolarDefault PolarType = 1 << iota
	PolarArea
	PolarPolygon
)

const currentColour = "currentColour"

type Renderer[T, U ScalerConstraint] interface {
	Render(Serie[T, U]) svg.Element
	NeedAxis() bool
}

type PolarRenderer[T ~string, U ~float64] struct {
	Fill       []string
	Radius     float64
	Ticks      int
	TicksStyle LineStyle
	Type       PolarType
	Stacked    bool
	Angular    bool
	Point      PointFunc
}

func (r PolarRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		g  = getBaseGroup("", "polar")
		el svg.Element
	)
	g.Append(r.drawTicks(serie))
	if r.Stacked {
		el = r.drawStackedArcs(serie)
	} else {
		el = r.drawArcs(serie)
	}
	if el != nil {
		g.Append(el)
	}
	return g.AsElement()
}

func (_ PolarRenderer[T, U]) NeedAxis() bool {
	return false
}

func (r PolarRenderer[T, U]) drawStackedArcs(serie Serie[T, U]) svg.Element {
	var (
		angle = fullcircle / float64(len(serie.Points))
		scale = serie.Y.replace(NewRange(0, r.Radius))
		grp   svg.Group
	)
	grp.Transform = svg.Translate(serie.X.Max()/2, serie.Y.Max()/2)
	for i, pt := range serie.Points {
		var (
			ori svg.Pos
			add float64
			pre float64
		)
		for j, p := range pt.Sub {
			var (
				pat   svg.Path
				color = r.Fill[j%len(r.Fill)]
				fill  = svg.NewFill(color)
				y     float64
				f     float64
				ag1   = angle * float64(i) * deg2rad
				ag2   = angle * float64(i+1) * deg2rad
			)
			pre = add
			f = scale.Scale(U(pre))
			add += any(p.Y).(float64)
			y = scale.Scale(U(add))
			pat.Fill = fill
			pat.Fill.Opacity = 0.7
			pat.Stroke = svg.NewStroke(color, 2)

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
	grp.Transform = svg.Translate(serie.X.Max()/2, serie.Y.Max()/2)
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
	grp.Transform = svg.Translate(serie.X.Max()/2, serie.Y.Max()/2)
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
		cx    = serie.X.Max() / 2
		cy    = serie.Y.Max() / 2
		sk    = svg.NewStroke("black", 1)
	)
	sk.Opacity = 0.25
	switch r.TicksStyle {
	case StyleStraight:
	case StyleDotted:
		sk.DashArray = append(sk.DashArray, 1, 5)
	case StyleDashed:
		sk.DashArray = append(sk.DashArray, 10, 5)
	}
	grp.Transform = svg.Translate(cx, cy)
	for i := 0; i < len(serie.Points); i++ {
		var (
			ag  = angle * float64(i) * deg2rad
			pos = getPosFromAngle(ag, r.Radius)
			li  = svg.NewLine(svg.NewPos(0, 0), pos)
			txt = svg.NewText(any(serie.Points[i].X).(string))
		)
		li.Stroke = sk
		txt.Font = svg.NewFont(FontSize)
		txt.Pos = pos
		switch a := int(angle) * i; {
		case a >= 345 || a <= 15:
			txt.Anchor, txt.Baseline = "start", "middle"
			txt.Pos.X += FontSize
		case a > 15 && a < 75:
			txt.Anchor, txt.Baseline = "start", "hanging"
			txt.Pos.X += FontSize * 0.5
			txt.Pos.Y += FontSize * 0.5
		case a >= 75 && a <= 105:
			txt.Anchor, txt.Baseline = "middle", "hanging"
			txt.Pos.Y += FontSize
		case a > 105 && a < 165:
			txt.Anchor, txt.Baseline = "end", "middle"
			txt.Pos.X -= FontSize * 0.5
			txt.Pos.Y += FontSize * 0.5
		case a >= 165 && a <= 195:
			txt.Anchor, txt.Baseline = "end", "middle"
			txt.Pos.X -= FontSize
		case a > 195 && a < 255:
			txt.Anchor, txt.Baseline = "end", "middle"
			txt.Pos.X -= FontSize * 0.5
			txt.Pos.Y -= FontSize * 0.5
		case a >= 255 && a <= 285:
			txt.Anchor, txt.Baseline = "middle", "start"
			txt.Pos.Y -= FontSize
		default:
			txt.Anchor, txt.Baseline = "start", "middle"
			txt.Pos.X += FontSize * 0.5
			txt.Pos.Y -= FontSize * 0.5
		}
		grp.Append(txt.AsElement())
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
		grp    = getBaseGroup("", "sun")
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

func (_ SunburstRenderer[T, U]) NeedAxis() bool {
	return false
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
	pat.Title = html.EscapeString(pt.String())
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
		grp   = getBaseGroup("", "pie")
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

func (_ PieRenderer[T, U]) NeedAxis() bool {
	return false
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
	Fill  []string
	Width float64
}

func (r GroupRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	return nil
}

func (_ GroupRenderer[T, U]) NeedAxis() bool {
	return true
}

type StackedRenderer[T ~string, U ~float64] struct {
	Fill      []string
	Width     float64
	Normalize bool
}

func (r StackedRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.Width <= 0 {
		r.Width = 1
	}
	if len(r.Fill) == 0 {
		r.Fill = Tableau10
	}
	var (
		grp  svg.Group
		max  = serie.Y.Max()
		size = serie.X.Space()
	)
	for _, parent := range serie.Points {
		var (
			offset float64
			bar    = getBaseGroup("", "bar")
		)
		bar.Transform = svg.Translate(serie.X.Scale(parent.X), 0)
		for i, pt := range parent.Sub {
			if r.Normalize {
				pt.Y = pt.Y / parent.Y
			}
			var (
				y  = serie.Y.Scale(pt.Y)
				w  = size * r.Width
				o  = (size - w) / 2
				el svg.Rect
			)
			el.Pos = svg.NewPos(o, y-offset)
			el.Dim = svg.NewDim(w, max-y)
			el.Fill = svg.NewFill(r.Fill[i%len(r.Fill)])
			if s, ok := any(pt.X).(string); ok {
				if r.Normalize {
					pt.Y *= 100
				}
				el.Title = fmt.Sprintf("%s: %.0f", html.EscapeString(s), pt.Y)
			}
			bar.Append(el.AsElement())

			offset += max - y
		}
		grp.Append(bar.AsElement())
	}
	return grp.AsElement()
}

func (_ StackedRenderer[T, U]) NeedAxis() bool {
	return true
}

type BarRenderer[T ~string, U ~float64] struct {
	Fill  []string
	Width float64
}

func (r BarRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	if r.Width <= 0 {
		r.Width = 1
	}
	if len(r.Fill) == 0 {
		r.Fill = Tableau10
	}
	grp := getBaseGroup("", "bar")
	for i, pt := range serie.Points {
		var (
			width  = serie.X.Space() * r.Width
			offset = (serie.X.Space() - width) / 2
			pos    = svg.NewPos(serie.X.Scale(pt.X)+offset, serie.Y.Scale(pt.Y))
			dim    = svg.NewDim(width, serie.Y.Max()-pos.Y)
		)
		var el svg.Rect
		el.Pos = pos
		el.Dim = dim
		el.Fill = svg.NewFill(r.Fill[i%len(r.Fill)])
		grp.Append(el.AsElement())
	}
	return grp.AsElement()
}

func (_ BarRenderer[T, U]) NeedAxis() bool {
	return true
}

type PointRenderer[T, U ScalerConstraint] struct {
	Color string
	Skip  int
	Point PointFunc
}

func (r PointRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	grp := getBaseGroup(r.Color, "scatter")
	for i, pt := range serie.Points {
		if r.Skip > 0 && i > 0 && i%r.Skip != 0 {
			continue
		}
		var (
			x = serie.X.Scale(pt.X)
			y = serie.Y.Scale(pt.Y)
		)
		el := r.Point(svg.NewPos(x, y))
		grp.Append(el)
	}
	return grp.AsElement()
}

func (_ PointRenderer[T, U]) NeedAxis() bool {
	return true
}

type CubicRenderer[T, U ScalerConstraint] struct {
	Stretch float64
	Color   string
	Fill    bool
	Skip    int
	Style   LineStyle
	Point   PointFunc
}

func (r CubicRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line")
		pat = getBasePath(r.Fill, r.Style)
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
	)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsMoveTo(pos)
	if r.Point != nil {
		grp.Append(r.Point(pos))
	}
	ori = pos
	for i, pt := range slices.Rest(serie.Points) {
		if r.Skip != 0 && i > 0 && i%r.Skip != 0 {
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)

		var (
			ctrl1 = ori
			ctrl2 = pos
			diff  = (pos.X - ori.X) * r.Stretch
		)
		ctrl1.X += diff
		ctrl2.X -= diff

		pat.AbsCubicCurve(pos, ctrl1, ctrl2)
		ori = pos
		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func (_ CubicRenderer[T, U]) NeedAxis() bool {
	return true
}

type LinearRenderer[T, U ScalerConstraint] struct {
	Fill          bool
	Color         string
	Skip          int
	Point         PointFunc
	Text          TextPosition
	Style         LineStyle
	IgnoreMissing bool
}

func (r LinearRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line")
		pat = getBasePath(r.Fill, r.Style)
		pos svg.Pos
		nan bool
	)
	grp.Id = serie.Title
	for i, pt := range serie.Points {
		if r.Skip != 0 && i > 0 && i%r.Skip == 0 {
			continue
		}
		if f, ok := isFloat(pt.Y); ok && math.IsNaN(f) {
			nan = true
			continue
		}
		pos.X = serie.X.Scale(pt.X)
		pos.Y = serie.Y.Scale(pt.Y)
		if i == 0 || (nan && r.IgnoreMissing) {
			nan = false
			pat.AbsMoveTo(pos)
		} else {
			pat.AbsLineTo(pos)
		}
		if r.Point != nil {
			el := r.Point(pos)
			if el != nil {
				grp.Append(r.Point(pos))
			}
		}
	}

	switch r.Text {
	case TextBefore:
		pt := slices.Fst(serie.Points)
		txt := getLineText(serie.Title, 0, serie.Y.Scale(pt.Y), true)
		grp.Append(txt.AsElement())
	case TextAfter:
		pt := slices.Lst(serie.Points)
		txt := getLineText(serie.Title, serie.X.Scale(pt.X), serie.Y.Scale(pt.Y), false)
		grp.Append(txt.AsElement())
	default:
	}

	if r.Fill {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func (_ LinearRenderer[T, U]) NeedAxis() bool {
	return true
}

type StepRenderer[T, U ScalerConstraint] struct {
	Color         string
	Fill          bool
	Point         PointFunc
	Text          TextPosition
	Style         LineStyle
	IgnoreMissing bool
}

func (r StepRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line", "line-step")
		pat = getBasePath(r.Fill, r.Style)
		pos = svg.NewPos(serie.X.Min(), serie.Y.Max())
		ori svg.Pos
		nan bool
	)
	grp.Id = serie.Title

	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pat.AbsLineTo(pos)
	if r.Point != nil {
		grp.Append(r.Point(pos))
	}
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
		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
	}
	switch r.Text {
	case TextBefore:
		pt := slices.Fst(serie.Points)
		txt := getLineText(serie.Title, 0, serie.Y.Scale(pt.Y), true)
		grp.Append(txt.AsElement())
	case TextAfter:
		pt := slices.Lst(serie.Points)
		txt := getLineText(serie.Title, serie.X.Scale(pt.X), serie.Y.Scale(pt.Y), false)
		grp.Append(txt.AsElement())
	default:
	}
	if r.Fill {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func (_ StepRenderer[T, U]) NeedAxis() bool {
	return true
}

type StepAfterRenderer[T, U ScalerConstraint] struct {
	Color         string
	Fill          bool
	Point         PointFunc
	Text          TextPosition
	Style         LineStyle
	IgnoreMissing bool
}

func (r StepAfterRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line", "line-step-after")
		pat = getBasePath(r.Fill, r.Style)
		pos svg.Pos
		ori svg.Pos
		nan bool
	)
	grp.Id = serie.Title

	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pos.Y = serie.Y.Max()
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	if r.Point != nil {
		grp.Append(r.Point(pos))
	}
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

		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
	}

	switch r.Text {
	case TextBefore:
		pt := slices.Fst(serie.Points)
		txt := getLineText(serie.Title, 0, serie.Y.Scale(pt.Y), true)
		grp.Append(txt.AsElement())
	case TextAfter:
		pt := slices.Lst(serie.Points)
		txt := getLineText(serie.Title, serie.X.Scale(pt.X), serie.Y.Scale(pt.Y), false)
		grp.Append(txt.AsElement())
	default:
	}

	if r.Fill {
		pos.X = serie.X.Max()
		pat.AbsLineTo(pos)
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func (_ StepAfterRenderer[T, U]) NeedAxis() bool {
	return true
}

type StepBeforeRenderer[T, U ScalerConstraint] struct {
	Color         string
	Fill          bool
	Point         PointFunc
	Text          TextPosition
	Style         LineStyle
	IgnoreMissing bool
}

func (r StepBeforeRenderer[T, U]) Render(serie Serie[T, U]) svg.Element {
	var (
		grp = getBaseGroup(r.Color, "line", "line-step-before")
		pat = getBasePath(r.Fill, r.Style)
		pos svg.Pos
		ori svg.Pos
		nan bool
	)
	grp.Id = serie.Title

	pos.X = serie.X.Min()
	pos.Y = serie.Y.Max()
	pat.AbsMoveTo(pos)
	pos.Y = serie.Y.Scale(slices.Fst(serie.Points).Y)
	pat.AbsLineTo(pos)
	pos.X = serie.X.Scale(slices.Fst(serie.Points).X)
	pat.AbsLineTo(pos)
	if r.Point != nil {
		grp.Append(r.Point(pos))
	}
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

		if r.Point != nil {
			grp.Append(r.Point(pos))
		}
	}

	switch r.Text {
	case TextBefore:
		pt := slices.Fst(serie.Points)
		txt := getLineText(serie.Title, 0, serie.Y.Scale(pt.Y), true)
		grp.Append(txt.AsElement())
	case TextAfter:
		pt := slices.Lst(serie.Points)
		txt := getLineText(serie.Title, serie.X.Scale(pt.X), serie.Y.Scale(pt.Y), false)
		grp.Append(txt.AsElement())
	default:
	}

	if r.Fill {
		pos.Y = serie.Y.Max()
		pat.AbsLineTo(pos)
	}
	grp.Append(pat.AsElement())
	return grp.AsElement()
}

func (_ StepBeforeRenderer[T, U]) NeedAxis() bool {
	return true
}

func getLineText(str string, x, y float64, before bool) svg.Text {
	txt := svg.NewText(str)
	txt.Font = svg.NewFont(FontSize)
	txt.Pos = svg.NewPos(x, y)
	txt.Anchor = "end"
	txt.Baseline = "middle"
	if !before {
		txt.Anchor = "start"
		txt.Pos.X += FontSize * 0.4
	} else {
		txt.Pos.X -= FontSize * 0.4
	}
	return txt
}

func getBasePath(fill bool, style LineStyle) svg.Path {
	var pat svg.Path
	pat.Rendering = "geometricPrecision"
	pat.Stroke = svg.NewStroke(currentColour, 1)
	pat.Stroke.LineJoin = "round"
	pat.Stroke.LineCap = "round"
	if fill {
		pat.Fill = svg.NewFill(currentColour)
		pat.Fill.Opacity = 0.5
	} else {
		pat.Fill = svg.NewFill("none")
	}
	switch style {
	case StyleStraight:
	case StyleDotted:
		pat.Stroke.DashArray = append(pat.Stroke.DashArray, 1, 5)
	case StyleDashed:
		pat.Stroke.DashArray = append(pat.Stroke.DashArray, 10, 5)
	default:
	}
	return pat
}

func getBaseGroup(color string, class ...string) svg.Group {
	var g svg.Group
	if color != "" {
		g.Fill = svg.NewFill(color)
		g.Stroke = svg.NewStroke(color, 1)
	}
	g.Class = class
	return g
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
