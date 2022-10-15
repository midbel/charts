package charts

import (
	"fmt"
	"time"

	"github.com/midbel/svg"
)

type Data interface {
	OffsetX() float64
	OffsetY() float64

	Render() svg.Element

	fmt.Stringer
}

type Serie[T, U ScalerConstraint] struct {
	Title string

	X      Scaler[T]
	Y      Scaler[U]
	Points []Point[T, U]
	Series []Serie[T, U]

	Renderer[T, U]
}

func (s Serie[T, U]) OffsetX() float64 {
	if s.X == nil {
		return 0
	}
	return s.X.Min()
}
func (s Serie[T, U]) OffsetY() float64 {
	if s.Y == nil {
		return 0
	}
	return s.Y.Min()
}

func (s Serie[T, U]) String() string {
	return s.Title
}

func (s Serie[T, U]) Render() svg.Element {
	return s.Renderer.Render(s)
}

func (s Serie[T, U]) Depth() int {
	if len(s.Series) == 0 {
		return 1
	}
	var depth int
	for _, e := range s.Series {
		d := e.Depth()
		if d > depth {
			depth = d
		}
	}
	return depth + 1
}

type Point[T, U ScalerConstraint] struct {
	X     T
	Y     U
	Label string
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

func (p Point[T, U]) String() string {
	return fmt.Sprintf("%v,%v", p.X, p.Y)
}

func (p Point[T, U]) Reverse() Point[U, T] {
	return Point[U, T]{
		X: p.Y,
		Y: p.X,
	}
}

func sumY[T ScalerConstraint, U ~float64](points []Point[T, U]) float64 {
	var s float64
	for _, p := range points {
		s += any(p.Y).(float64)
	}
	return s
}
