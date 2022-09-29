package charts

import (
	"time"
)

type ScalerConstraint interface {
	~float64 | ~string | time.Time
}

type Domain[T ScalerConstraint] interface {
	Diff(T) float64
	Extend() float64
	Values(int) []T
}

type numberDomain struct {
	fst float64
	lst float64
}

func NumberDomain(f, t float64) Domain[float64] {
	return numberDomain{
		fst: f,
		lst: t,
	}
}

func (n numberDomain) Diff(v float64) float64 {
	return v - n.fst
}

func (n numberDomain) Extend() float64 {
	return n.lst - n.fst
}

func (n numberDomain) Values(c int) []float64 {
	var (
		all  = make([]float64, c)
		step = n.Extend() / float64(c)
	)
	for i := 0; i < c; i++ {
		all[i] = n.fst + float64(i)*step
	}
	all = append(all, n.lst)
	return all
}

type timeDomain struct {
	fst time.Time
	lst time.Time
}

func TimeDomain(f, t time.Time) Domain[time.Time] {
	return timeDomain{
		fst: f,
		lst: t,
	}
}

func (t timeDomain) Diff(v time.Time) float64 {
	diff := v.Sub(t.fst)
	return float64(diff)
}

func (t timeDomain) Extend() float64 {
	diff := t.lst.Sub(t.fst)
	return float64(diff)
}

func (t timeDomain) Values(c int) []time.Time {
	var (
		all  = make([]time.Time, c)
		step = t.Extend() / float64(c)
	)
	for i := 0; i < c; i++ {
		all[i] = t.fst.Add(time.Duration(float64(i) * step))
	}
	all = append(all, t.lst)
	return all
}

type Range struct {
	F float64
	T float64
}

func NewRange(f, t float64) Range {
	return Range{
		F: f,
		T: t,
	}
}

func (r Range) Len() float64 {
	return r.T - r.F
}

func (r Range) Max() float64 {
	return r.T
}

func (r Range) Min() float64 {
	return r.F
}

type Scaler[T ScalerConstraint] interface {
	Scale(T) float64
	Space() float64
	Values(int) []T
	Max() float64
	Min() float64
}

type numberScaler struct {
	Range
	Domain[float64]
}

func NumberScaler(dom Domain[float64], rg Range) Scaler[float64] {
	return numberScaler{
		Range:  rg,
		Domain: dom,
	}
}

func (n numberScaler) Scale(v float64) float64 {
	return n.Diff(v) * n.Space()
}

func (n numberScaler) Space() float64 {
	return n.Len() / n.Extend()
}

type timeScaler struct {
	Range
	Domain[time.Time]
}

func TimeScaler(dom Domain[time.Time], rg Range) Scaler[time.Time] {
	return timeScaler{
		Range:  rg,
		Domain: dom,
	}
}

func (s timeScaler) Scale(v time.Time) float64 {
	return s.Diff(v) * s.Space()
}

func (s timeScaler) Space() float64 {
	return s.Len() / s.Extend()
}

type stringScaler struct {
	Range
	Strings []string
}

func StringScaler(str []string, rg Range) Scaler[string] {
	return stringScaler{
		Range:   rg,
		Strings: str,
	}
}

func (s stringScaler) Scale(v string) float64 {
	var x int
	for i := range s.Strings {
		if s.Strings[i] == v {
			x = i
			break
		}
	}
	return float64(x) * s.Space()
}

func (s stringScaler) Space() float64 {
	return s.Len() / float64(len(s.Strings))
}

func (s stringScaler) Values(c int) []string {
	return nil
}
