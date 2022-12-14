package charts

import (
	"bufio"
	"io"

	"github.com/midbel/svg"
)

type Drawner interface {
	Drawn(...Data) svg.Element
}

type Chart[T, U ScalerConstraint] struct {
	Title  string
	Width  float64
	Height float64

	Padding

	Left   Axis[U]
	Right  Axis[U]
	Top    Axis[T]
	Bottom Axis[T]

	Legend struct {
		Title  string
		Orient Orientation
	}
	Center Point[T, U]
	Theme  string
}

func (c Chart[T, U]) DrawingWidth() float64 {
	return c.Width - c.Padding.Horizontal()
}

func (c Chart[T, U]) DrawingHeight() float64 {
	return c.Height - c.Padding.Vertical()
}

func (c Chart[T, U]) Drawn(set ...Data) svg.Element {
	var el svg.SVG
	el.Dim = svg.NewDim(c.Width, c.Height)
	el.OmitProlog = true

	el.Append(c.getDefs())
	if c.Theme != "" {
		s := svg.Style{
			Content: c.Theme,
			Type:    "text/css",
		}
		el.Append(s.AsElement())
	}

	el.Append(c.drawAxis())
	for _, s := range set {
		ar := c.getArea(s)
		el.Append(ar.AsElement())
	}
	// if ld := c.drawLegend(set); ld != nil {
	// 	el.Append(ld)
	// }
	if txt := c.drawTitle(); txt != nil {
		el.Append(txt)
	}
	return el.AsElement()
}

func (c Chart[T, U]) Render(w io.Writer, set ...Data) {
	el := c.Drawn(set...)

	bw := bufio.NewWriter(w)
	defer bw.Flush()
	el.Render(bw)
}

func (c Chart[T, U]) getArea(serie Data) svg.Group {
	var g svg.Group
	g.Class = append(g.Class, "area")
	if id := serie.Id(); id != "" {
		g.Id = id
	}
	g.Clip = "clip-chart"
	g.Transform = svg.Translate(c.Padding.Left, c.Padding.Top)

	g.Append(serie.Render())
	return g
}

func (c Chart[T, U]) getDefs() svg.Element {
	var defs svg.Defs
	defs.Append(c.getClip())
	return defs.AsElement()
}

func (c Chart[T, U]) getClip() svg.Element {
	var (
		w    = c.Width - c.Padding.Horizontal()
		h    = c.Height - c.Padding.Vertical()
		rec  svg.Rect
		clip svg.ClipPath
	)
	rec.Dim = svg.NewDim(w, h)
	clip.Id = "clip-chart"
	clip.Append(rec.AsElement())
	return clip.AsElement()
}

func (c Chart[T, U]) drawTitle() svg.Element {
	if c.Title == "" {
		return nil
	}
	txt := svg.NewText(c.Title)
	txt.Font = svg.NewFont(FontSize * 1.2)
	txt.Class = append(txt.Class, "chart-title")
	txt.Anchor = "middle"
	txt.Baseline = "auto"
	txt.Pos.X = c.Width / 2
	txt.Pos.Y = c.Padding.Top / 2
	if c.Padding.Top == 0 {
		txt.Pos.Y = FontSize * 1.1
	}
	return txt.AsElement()
}

func (c Chart[T, U]) drawLegend(series []Data) svg.Element {
	var (
		offset = FontSize * 1.4
		height = float64(len(series)) * offset
		width  float64
		grp    svg.Group
	)
	if c.Legend.Title != "" {
		height += offset
	}
	for i, s := range series {
		title := s.String()
		if n := float64(len(title)); i == 0 || n > width {
			width = n
		}
		var g svg.Group
		g.Transform = svg.Translate(0, float64(i)*offset)
		// li := svg.NewLine(svg.NewPos(0, 0), svg.NewPos(20, 0))
		// li.Stroke = svg.NewStroke(s.GetColor(), 4)

		tx := svg.NewText(title)
		tx.Pos = svg.NewPos(30, 0)
		tx.Font = svg.NewFont(FontSize)
		tx.Baseline = "middle"

		// g.Append(li.AsElement())
		g.Append(tx.AsElement())
		grp.Append(g.AsElement())
	}
	width *= FontSize * 0.4

	var left, top float64
	switch c.Legend.Orient {
	case OrientRight:
		left = c.Width - c.Padding.Left - width
		top = (c.Height - c.Padding.Top - height) / 2
	case OrientRight | OrientBottom:
		left = c.Width - c.Padding.Left - width
		top = c.Height - c.Padding.Top - height
	case OrientBottom:
		left = (c.Width - width) / 2
		top = c.Height - c.Padding.Top - height
	case OrientLeft | OrientBottom:
		left = c.Padding.Left
		top = c.Height - c.Padding.Top - height
	case OrientLeft:
		left = c.Padding.Left
		top = (c.Height - c.Padding.Vertical() - height) / 2
	case OrientLeft | OrientTop:
		top = c.Padding.Top
		left = c.Padding.Left
	case OrientTop:
		left = (c.Width - width) / 2
		top = c.Padding.Top
	case OrientRight | OrientTop:
		top = c.Padding.Top
		left = c.Width - c.Padding.Left - width
	default:
		return nil
	}
	grp.Transform = svg.Translate(left, top)
	return grp.AsElement()
}

func (c Chart[T, U]) drawAxis() svg.Element {
	var g svg.Group
	g.Id = "axis"
	if c.Left.Scaler != nil {
		c.Left.Orientation = OrientLeft
		el := c.Left.Render(c.DrawingHeight(), c.DrawingWidth(), c.Padding.Left, c.Padding.Top)
		g.Append(el)
		if c.Left.Label != "" {
			txt := axisText(OrientLeft, c.Left.Label, FontSize*1.2, c.Height/2, svg.NewFont(FontSize))
			g.Append(txt.AsElement())
		}
	}
	if c.Right.Scaler != nil {
		c.Right.Orientation = OrientRight
		el := c.Right.Render(c.DrawingHeight(), c.DrawingWidth(), c.Width-c.Padding.Right, c.Padding.Top)
		g.Append(el)
		if c.Right.Label != "" {
			txt := axisText(OrientRight, c.Right.Label, c.Width-FontSize*1.2, c.Height/2, svg.NewFont(FontSize))
			g.Append(txt.AsElement())
		}
	}
	if c.Top.Scaler != nil {
		c.Top.Orientation = OrientTop
		el := c.Top.Render(c.DrawingWidth(), c.DrawingHeight(), c.Padding.Left, c.Padding.Top)
		g.Append(el)
		if c.Top.Label != "" {
			txt := axisText(OrientTop, c.Top.Label, c.Width/2, FontSize*1.2, svg.NewFont(FontSize))
			g.Append(txt.AsElement())
		}
	}
	if c.Bottom.Scaler != nil {
		c.Bottom.Orientation = OrientBottom
		el := c.Bottom.Render(c.DrawingWidth(), c.DrawingHeight(), c.Padding.Left, c.Height-c.Padding.Bottom)
		g.Append(el)
		if c.Bottom.Label != "" {
			txt := axisText(OrientBottom, c.Bottom.Label, c.Width/2, c.Height-FontSize*1.2, svg.NewFont(FontSize))
			g.Append(txt.AsElement())
		}
	}
	return g.AsElement()
}
