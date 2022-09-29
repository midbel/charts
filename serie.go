package charts

import (
	"github.com/midbel/svg"
)

type Serie[T, U ScalerConstraint] struct {
	Color string

	Points []Point[T, U]
	X      Scaler[T]
	Y      Scaler[U]

	Renderer  Renderer[T, U]
	WithPoint func(svg.Pos) svg.Element
	WithArea  bool
}

func (s Serie[T, U]) Render() svg.Element {
	return s.Renderer.Render(s)
}
