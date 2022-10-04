package charts

import (
	"bufio"
	"io"

	"github.com/midbel/svg"
)

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

type Chart[T, U ScalerConstraint] struct {
	Title  string
	Width  float64
	Height float64

	Padding

	Left   Axis
	Right  Axis
	Top    Axis
	Bottom Axis

	Legend struct {
		Title  string
		Orient Orientation
	}
	Center Point[T, U]
}

func (c Chart[T, U]) DrawingWidth() float64 {
	return c.Width - c.Padding.Horizontal()
}

func (c Chart[T, U]) DrawingHeight() float64 {
	return c.Height - c.Padding.Vertical()
}

func (c Chart[T, U]) Render(w io.Writer, set ...Data) {
	el := svg.NewSVG(svg.WithDimension(c.Width, c.Height))
	el.OmitProlog = true

	el.Append(c.drawAxis())
	for _, s := range set {
		ar := c.getArea(s)
		ar.Append(s.Render())
		el.Append(ar.AsElement())
	}
	// if lg := c.drawLegend(series); lg != nil {
	// 	el.Append(lg)
	// }

	bw := bufio.NewWriter(w)
	defer bw.Flush()
	el.Render(bw)
}

func (c Chart[T, U]) getArea(s Data) svg.Group {
	var g svg.Group
	g.Class = append(g.Class, "area")
	g.Transform = svg.Translate(c.Padding.Left-s.OffsetX(), c.Padding.Top+s.OffsetY())
	return g
}

func (c Chart[T, U]) drawLegend(series []Serie[T, U]) svg.Element {
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
		if n := float64(len(s.Title)); i == 0 || n > width {
			width = n
		}
		var g svg.Group
		g.Transform = svg.Translate(0, float64(i)*offset)
		li := svg.NewLine(svg.NewPos(0, 0), svg.NewPos(20, 0))
		li.Stroke = svg.NewStroke(s.Color, 1)

		tx := svg.NewText(s.Title)
		tx.Pos = svg.NewPos(30, 0)
		tx.Font = svg.NewFont(FontSize)
		tx.Baseline = "middle"

		g.Append(li.AsElement())
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
	g := svg.NewGroup(svg.WithID("axis"))
	if c.Left != nil {
		el := c.Left.Render(c.DrawingHeight(), c.DrawingWidth(), c.Padding.Left, c.Padding.Top)
		g.Append(el)
	}
	if c.Right != nil {
		el := c.Right.Render(c.DrawingHeight(), c.DrawingWidth(), c.Width-c.Padding.Right, c.Padding.Top)
		g.Append(el)
	}
	if c.Top != nil {
		el := c.Top.Render(c.DrawingWidth(), c.DrawingHeight(), c.Padding.Left, c.Padding.Top)
		g.Append(el)
	}
	if c.Bottom != nil {
		el := c.Bottom.Render(c.DrawingWidth(), c.DrawingHeight(), c.Padding.Left, c.Height-c.Padding.Bottom)
		g.Append(el)
	}
	return g.AsElement()
}
