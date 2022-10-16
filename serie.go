package charts

import (
	"fmt"
	"time"

	"github.com/midbel/slices"
	"github.com/midbel/svg"
)

type Data interface {
	Render() svg.Element

	fmt.Stringer
}

type Serie[T, U ScalerConstraint] struct {
	Title string

	X      Scaler[T]
	Y      Scaler[U]
	Points []Point[T, U]

	Renderer[T, U]
}

func (s Serie[T, U]) Sum() float64 {
	pt := slices.Fst(s.Points)
	if _, ok := any(pt.Y).(float64); !ok {
		z := len(s.Points)
		return float64(z)
	}
	var sum float64
	for i := range s.Points {
		sum += any(s.Points[i].Y).(float64)
	}
	return sum
}

func (s Serie[T, U]) String() string {
	return s.Title
}

func (s Serie[T, U]) Render() svg.Element {
	return s.Renderer.Render(s)
}

type Point[T, U ScalerConstraint] struct {
	X   T
	Y   U
	Sub []Point[T, U]
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

func (p Point[T, U]) Depth() int {
	if len(p.Sub) == 0 {
		return 1
	}
	var depth int
	for _, e := range p.Sub {
		d := e.Depth()
		if d > depth {
			depth = d
		}
	}
	return depth + 1
}
