package charts

import (
	"github.com/midbel/svg"
)

type Serie[T, U ScalerConstraint] struct {
	WithPoint bool
	WithArea  bool
	Color     string

	Points []Point[T, U]
	X      Scaler[T]
	Y      Scaler[U]

	Renderer Renderer[T, U]
}

func (s Serie[T, U]) Render() svg.Element {
	return s.Renderer.Render(s)
}
