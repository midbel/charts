package charts

import (
	"time"

	"github.com/midbel/svg"
)

type Serie[T, U ScalerConstraint] struct {
	Color         string
	Title         string
	IgnoreMissing bool

	X      Scaler[T]
	Y      Scaler[U]
	Points []Point[T, U]
	Series []Serie[T, U]

	Renderer Renderer[T, U]
}

func (s Serie[T, U]) Render() svg.Element {
	return s.Renderer.Render(s)
}

type Point[T, U ScalerConstraint] struct {
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

func CategoryPoint(x string, y float64) Point[string, float64] {
	return Point[string, float64]{
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
