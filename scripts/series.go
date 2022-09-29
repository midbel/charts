package main

import (
	//	"math"
	"os"
	"strconv"
	"time"

	"github.com/midbel/charts"
	"github.com/midbel/svg"
)

func main() {
	var (
		dtstart = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		dtend   = time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)
		xscale2 = charts.TimeScaler(charts.TimeDomain(dtstart, dtend), charts.NewRange(0, 700))
		yscale2 = charts.NumberScaler(charts.NumberDomain(100, 0), charts.NewRange(0, 520))
	)

	var serie2 charts.Serie[time.Time, float64]
	serie2.Renderer = charts.CubicRender[time.Time, float64](0.25, true)
	serie2.Color = "red"
	serie2.WithPoint = func(p svg.Pos) svg.Element {
		ci := charts.GetCirclePoint(p)
		ci.Fill = svg.NewFill(serie2.Color)
		return ci.AsElement()
	}
	serie2.X = xscale2
	serie2.Y = yscale2
	serie2.Points = []charts.Point[time.Time, float64]{
		charts.TimePoint(time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC), 34),
		charts.TimePoint(time.Date(2022, 1, 20, 0, 0, 0, 0, time.UTC), 39),
		charts.TimePoint(time.Date(2022, 2, 10, 0, 0, 0, 0, time.UTC), 40),
		charts.TimePoint(time.Date(2022, 2, 26, 0, 0, 0, 0, time.UTC), 45),
		charts.TimePoint(time.Date(2022, 3, 7, 0, 0, 0, 0, time.UTC), 43),
		charts.TimePoint(time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC), 43),
		charts.TimePoint(time.Date(2022, 6, 11, 0, 0, 0, 0, time.UTC), 67),
		charts.TimePoint(time.Date(2022, 6, 29, 0, 0, 0, 0, time.UTC), 80),
		charts.TimePoint(time.Date(2022, 7, 6, 0, 0, 0, 0, time.UTC), 89),
		//charts.TimePoint(time.Date(2022, 7, 23, 0, 0, 0, 0, time.UTC), math.NaN()),
		//charts.TimePoint(time.Date(2022, 8, 3, 0, 0, 0, 0, time.UTC), math.NaN()),
		//charts.TimePoint(time.Date(2022, 8, 15, 0, 0, 0, 0, time.UTC), math.NaN()),
		//charts.TimePoint(time.Date(2022, 8, 20, 0, 0, 0, 0, time.UTC), math.NaN()),
		charts.TimePoint(time.Date(2022, 9, 5, 0, 0, 0, 0, time.UTC), 98),
		charts.TimePoint(time.Date(2022, 9, 12, 0, 0, 0, 0, time.UTC), 78),
		//charts.TimePoint(time.Date(2022, 9, 20, 0, 0, 0, 0, time.UTC), math.NaN()),
		charts.TimePoint(time.Date(2022, 10, 3, 0, 0, 0, 0, time.UTC), 19),
		charts.TimePoint(time.Date(2022, 10, 7, 0, 0, 0, 0, time.UTC), 22),
		charts.TimePoint(time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC), 54),
		charts.TimePoint(time.Date(2022, 11, 19, 0, 0, 0, 0, time.UTC), 98),
		charts.TimePoint(time.Date(2022, 12, 6, 0, 0, 0, 0, time.UTC), 86),
		charts.TimePoint(time.Date(2022, 12, 16, 0, 0, 0, 0, time.UTC), 54),
		charts.TimePoint(time.Date(2022, 12, 25, 0, 0, 0, 0, time.UTC), 12),
		charts.TimePoint(time.Date(2022, 12, 30, 0, 0, 0, 0, time.UTC), 1),
	}

	left := charts.NumberAxis{
		Ticks:          10,
		Orientation:    charts.OrientLeft,
		Scaler:         yscale2,
		WithInnerTicks: true,
		WithLabelTicks: true,
		WithOuterTicks: true,
		Format: func(f float64) string {
			return strconv.Itoa(int(f))
		},
	}
	bot := charts.TimeAxis{
		Ticks:          10,
		Orientation:    charts.OrientBottom,
		Scaler:         xscale2,
		WithInnerTicks: true,
		WithLabelTicks: true,
		WithOuterTicks: false,
	}
	ch := charts.Chart[time.Time, float64]{
		Width:  800,
		Height: 600,
		Padding: charts.Padding{
			Top:    40,
			Bottom: 40,
			Left:   60,
			Right:  40,
		},
		Left:   left,
		Bottom: bot,
	}
	ch.Render(os.Stdout, serie2)
}
