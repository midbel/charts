package main

import (
	"os"
	_ "time"

	"github.com/midbel/charts"
)

func main() {
	var (
		xscale1 = charts.NumberScaler(charts.NumberDomain(0, 720), charts.NewRange(0, 700))
		yscale1 = charts.NumberScaler(charts.NumberDomain(720, 0), charts.NewRange(0, 520))
	)

	var serie1 charts.Serie[float64, float64]
	serie1.Renderer = charts.StepBeforeRender[float64, float64]()
	serie1.Color = "blue"
	serie1.WithArea = true
	serie1.WithPoint = true
	serie1.X = xscale1
	serie1.Y = yscale1
	serie1.Points = []charts.Point[float64, float64]{
		charts.NumberPoint(100, 245),
		charts.NumberPoint(150, 567),
		charts.NumberPoint(324, 98),
		charts.NumberPoint(461, 19),
		charts.NumberPoint(511, 563),
		charts.NumberPoint(541, 463),
		charts.NumberPoint(571, 113),
		charts.NumberPoint(591, 703),
		charts.NumberPoint(645, 301),
		charts.NumberPoint(716, 341),
	}

	/*	var (
			dtstart = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
			dtend   = time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)
		)

		var (
			xscale2 = charts.TimeScaler(charts.TimeDomain(dtstart, dtend), charts.NewRange(0, 720))
			yscale2 = charts.NumberScaler(charts.NumberDomain(100, 0), charts.NewRange(0, 520))
		)

		var serie2 charts.Serie[time.Time, float64]
		serie2.Renderer = charts.StepAfterRender[time.Time, float64]()
		serie2.Color = "red"
		serie2.WithArea = true
		serie2.WithPoint = true
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
			charts.TimePoint(time.Date(2022, 9, 5, 0, 0, 0, 0, time.UTC), 98),
			charts.TimePoint(time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC), 19),
			charts.TimePoint(time.Date(2022, 11, 19, 0, 0, 0, 0, time.UTC), 98),
			charts.TimePoint(time.Date(2022, 12, 6, 0, 0, 0, 0, time.UTC), 86),
			charts.TimePoint(time.Date(2022, 12, 16, 0, 0, 0, 0, time.UTC), 54),
			charts.TimePoint(time.Date(2022, 12, 25, 0, 0, 0, 0, time.UTC), 12),
			charts.TimePoint(time.Date(2022, 12, 30, 0, 0, 0, 0, time.UTC), 1),
		}*/

	left := charts.NumberAxis{
		Ticks:       10,
		Orientation: charts.OrientLeft,
		Scaler:      yscale1,
	}
	/*	right := charts.NumberAxis{
		Ticks:  10,
		Scaler: yscale2,
	}*/
	bot := charts.NumberAxis{
		Ticks:       10,
		Orientation: charts.OrientBottom,
		Scaler:      xscale1,
	}
	/*	top := charts.TimeAxis{
		Ticks:  10,
		Scaler: xscale2,
	}*/
	ch := charts.Chart[float64, float64]{
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
		/*		Right:  right,
				Top:    top,*/
	}
	ch.Render(os.Stdout, serie1)
}
