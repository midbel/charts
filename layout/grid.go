package layout

import (
	"bufio"
	"fmt"
	"io"

	"github.com/midbel/svg"
)

type Renderer interface {
	Element() svg.Element
}

type Cell struct {
	X    int
	Y    int
	W    int
	H    int
	Item Renderer
}

type Grid struct {
	Rows   int
	Cols   int
	Width  float64
	Height float64

	Cells []Cell
}

func (g Grid) Render(w io.Writer) error {
	var (
		width  = g.Width / float64(g.Cols)
		height = g.Height / float64(g.Rows)
		grid   svg.SVG
	)
	grid.Dim = svg.NewDim(g.Width, g.Height)
	for i, c := range g.Cells {
		var (
			g svg.Group
			x = width * float64(c.Y)
			y = height * float64(c.X)
		)
		g.Class = append(g.Class, "grid", "cell")
		g.Id = fmt.Sprintf("cell-%03d", i+1)
		g.Transform = svg.Translate(x, y)

		el := c.Item.Element()
		if e, ok := el.(*svg.SVG); ok {
			e.ViewBox.Dim = e.Dim
			e.Dim = svg.NewDim(float64(c.W)*width, float64(c.H)*height)
			e.PreserveRatio.Align = "none"
			el = e.AsElement()
		}
		g.Append(el)

		grid.Append(g.AsElement())
	}

	ws := bufio.NewWriter(w)
	defer ws.Flush()
	grid.Render(ws)
	return nil
}
