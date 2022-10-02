package charts

import (
	"github.com/midbel/svg"
)

var DefaultSize float64 = 4

type PointFunc func(svg.Pos) svg.Element

func GetCircle(pos svg.Pos) svg.Element {
	var el svg.Circle
	el.Pos = pos
	el.Fill = svg.NewFill(currentColour)
	el.Radius = DefaultSize / 2
	return el.AsElement()
}

func GetSquare(pos svg.Pos) svg.Element {
	half := DefaultSize / 2
	pos.X -= half
	pos.Y -= half

	var el svg.Rect
	el.Pos = pos
	el.Dim = svg.NewDim(DefaultSize, DefaultSize)
	el.Fill = svg.NewFill(currentColour)

	return el.AsElement()
}

func GetDiamond(pos svg.Pos) svg.Element {
	half := DefaultSize / 2
	pos.X -= half
	pos.Y -= half

	var el svg.Rect
	el.Pos = pos
	el.Dim = svg.NewDim(DefaultSize, DefaultSize)
	el.Fill = svg.NewFill(currentColour)
	el.Transform.RA = 45
	el.Transform.RX = pos.X + half
	el.Transform.RY = pos.Y + half

	return el.AsElement()
}
