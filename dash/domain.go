package dash

import (
	"errors"
	"fmt"
	"time"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/eval"
	"github.com/midbel/buddy/types"
	"github.com/midbel/charts"
)

var errScaler = errors.New("scale not set")

type Input struct {
	Type   string
	Scaler ScalerMaker
	Domain
}

func (i Input) isNumber() bool {
	return i.Type == TypeNumber
}

func (i Input) isTime() bool {
	return i.Type == TypeTime
}

func (i Input) isString() bool {
	return i.Type == TypeString
}

func (i Input) CategoryScale(rg charts.Range) (charts.Scaler[string], error) {
	if i.Scaler == nil {
		return nil, errScaler
	}
	return i.Scaler.CategoryScale(rg)
}

func (i Input) NumberScale(rg charts.Range, reverse bool) (charts.Scaler[float64], error) {
	if i.Scaler == nil {
		return nil, errScaler
	}
	return i.Scaler.NumberScale(rg, reverse)
}

func (i Input) TimeScale(rg charts.Range, format string, reverse bool) (charts.Scaler[time.Time], error) {
	if i.Scaler == nil {
		return nil, errScaler
	}
	return i.Scaler.TimeScale(rg, format, reverse)
}

type Domain struct {
	Label      string
	Ticks      int
	Format     string
	Position   string
	InnerTicks bool
	OuterTicks bool
	LabelTicks bool
	BandTicks  bool
}

func (d Domain) GetCategoryAxis(cfg Config, scale charts.Scaler[string]) (charts.Axis[string], error) {
	var (
		axe    = createAxis[string](d, scale)
		format func(string) string
	)
	if expr, err := cfg.Scripts.Resolve(d.Format); err == nil {
		format = wrapExpr[string](expr)
	} else {
		format = func(s string) string {
			return s
		}
	}
	axe.Format = format
	return axe, nil
}

func (d Domain) GetNumberAxis(cfg Config, scale charts.Scaler[float64]) (charts.Axis[float64], error) {
	var (
		axe    = createAxis[float64](d, scale)
		format func(float64) string
	)

	if expr, err := cfg.Scripts.Resolve(d.Format); err == nil {
		format = wrapExpr[float64](expr)
	} else {
		format = func(f float64) string {
			return fmt.Sprintf(d.Format, f)
		}
	}
	axe.Format = format
	return axe, nil
}

func (d Domain) GetTimeAxis(cfg Config, scale charts.Scaler[time.Time]) (charts.Axis[time.Time], error) {
	formatTime, err := makeTimeFormat(d.Format)
	if err != nil {
		return charts.Axis[time.Time]{}, err
	}
	axe := createAxis[time.Time](d, scale)
	axe.Format = formatTime
	return axe, nil
}

func createAxis[T charts.ScalerConstraint](d Domain, scale charts.Scaler[T]) charts.Axis[T] {
	return charts.Axis[T]{
		Label:          d.Label,
		Ticks:          d.Ticks,
		Scaler:         scale,
		WithInnerTicks: d.InnerTicks,
		WithOuterTicks: d.OuterTicks,
		WithLabelTicks: d.LabelTicks,
		WithBands:      d.BandTicks,
	}
}

func wrapExpr[T any](expr ast.Expression) func(value T) string {
	return func(value T) string {
		p, err := types.CreatePrimitive(value)
		if err != nil {
			return ""
		}
		env := types.EmptyEnv()
		env.Define("value", p)
		res, err := eval.Execute(expr, env)
		if err != nil {
			return ""
		}
		fmt.Printf("%T - %s\n", res, res)
		return res.String()
	}
}
