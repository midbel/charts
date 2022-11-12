package dash

import (
	"fmt"
	"time"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/eval"
	"github.com/midbel/buddy/types"
	"github.com/midbel/charts"
)

type Domain struct {
	Label      string
	Ticks      int
	Format     string
	Domain     ScalerMaker
	Position   string
	InnerTicks bool
	OuterTicks bool
	LabelTicks bool
	BandTicks  bool
}

func (d Domain) makeCategoryScale(rg charts.Range) (charts.Scaler[string], error) {
	if d.Domain == nil {
		return nil, fmt.Errorf("domain not set")
	}
	return d.Domain.makeCategoryScale(rg)
}

func (d Domain) makeNumberScale(rg charts.Range, reverse bool) (charts.Scaler[float64], error) {
	if d.Domain == nil {
		return nil, fmt.Errorf("domain not set")
	}
	return d.Domain.makeNumberScale(rg, reverse)
}

func (d Domain) makeTimeScale(rg charts.Range, reverse bool) (charts.Scaler[time.Time], error) {
	if d.Domain == nil {
		return nil, fmt.Errorf("domain not set")
	}
	return d.Domain.makeTimeScale(rg, d.Format, reverse)
}

func (d Domain) makeCategoryAxis(cfg Config, scale charts.Scaler[string]) (charts.Axis[string], error) {
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

func (d Domain) makeNumberAxis(cfg Config, scale charts.Scaler[float64]) (charts.Axis[float64], error) {
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

func (d Domain) makeTimeAxis(cfg Config, scale charts.Scaler[time.Time]) (charts.Axis[time.Time], error) {
	formatTime, err := makeTimeFormat(d.Format)
	if err != nil {
		return charts.Axis[time.Time]{}, err
	}
	axe := createAxis[time.Time](d, scale)
	axe.Format = formatTime
	return axe, nil
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
			return fmt.Sprintf("error: %s", err)
		}
		return res.String()
	}
}
