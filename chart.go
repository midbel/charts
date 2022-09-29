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
}

func (c Chart[T, U]) DrawingWidth() float64 {
	return c.Width - c.Padding.Horizontal()
}

func (c Chart[T, U]) DrawingHeight() float64 {
	return c.Height - c.Padding.Vertical()
}

func (c Chart[T, U]) Render(w io.Writer, series ...Serie[T, U]) {
	el := svg.NewSVG(svg.WithDimension(c.Width, c.Height))
	ar := c.getArea()

	el.Append(c.drawAxis())
	for _, s := range series {
		g := s.Render()
		ar.Append(g)
	}
	el.Append(ar.AsElement())

	bw := bufio.NewWriter(w)
	defer bw.Flush()
	el.Render(bw)
}

func (c Chart[T, U]) getArea() svg.Group {
	var g svg.Group
	g.Id = "area"
	g.Transform.TX = c.Padding.Left
	g.Transform.TY = c.Padding.Top
	return g
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
