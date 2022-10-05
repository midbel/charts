package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/midbel/charts"
)

const (
	defaultWidth  = 1280
	defaultHeight = 360
	offsetHeight  = 300
)

var pad = charts.Padding{
	Top:    10,
	Right:  45,
	Bottom: 60,
	Left:   45,
}

func main() {
	flag.Parse()
	var (
		dtstart     = time.Date(2022, 9, 8, 0, 0, 0, 0, time.UTC)
		dtend       = time.Date(2022, 9, 28, 0, 0, 0, 0, time.UTC)
		priceDomain = charts.NumberDomain(300, 0)
		timeScale   = charts.TimeScaler(charts.TimeDomain(dtstart, dtend), charts.NewRange(0, defaultWidth-pad.Horizontal()))
		priceScale  = charts.NumberScaler(charts.NumberDomain(280, 220), charts.NewRange(0, defaultHeight-pad.Vertical()))
		types       = []string{"open", "high", "low", "close"}
	)

	serie, list, err := loadSerie(flag.Arg(0), "steelblue")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	serie.X = timeScale
	serie.Y = priceScale

	rdr := charts.StackedRenderer[string, float64]{
		Width:      0.6,
		Horizontal: true,
		Fill:       []string{"steelblue", "cornflowerblue", "darkorange", "orange"},
	}
	var dates []string
	for i := range list {
		dates = append(dates, list[i].Title)
	}
	datesScale := charts.StringScaler(dates, charts.NewRange(0, defaultWidth-pad.Horizontal()))
	typesScale := charts.StringScaler(types, charts.NewRange(0, datesScale.Space()))
	valuesScale := charts.NumberScaler(priceDomain, charts.NewRange(offsetHeight, defaultHeight-pad.Vertical()-offsetHeight))
	for i := range list {
		list[i].X = typesScale
		list[i].Y = valuesScale
	}
	day := charts.Serie[string, float64]{
		Title:    "evolution",
		Series:   list,
		X:        datesScale,
		Y:        valuesScale,
		Renderer: rdr,
	}

	ch := charts.Chart[time.Time, float64]{
		Width:   defaultWidth,
		Height:  defaultHeight,
		Padding: pad,
		Right:   getAxisY(priceScale),
		Bottom:  getAxisX(timeScale),
	}
	ch.Render(os.Stdout, serie, day)
}

func loadSerie(file, color string) (charts.Serie[time.Time, float64], []charts.Serie[string, float64], error) {
	var (
		ser  = getSerie("GOOG", color)
		list []charts.Serie[string, float64]
		err  error
	)
	ser.Points, list, err = loadPoints(file)
	return ser, list, err
}

func loadPoints(file string) ([]charts.Point[time.Time, float64], []charts.Serie[string, float64], error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()

	var (
		rs = csv.NewReader(r)
		ps []charts.Point[time.Time, float64]
		ls []charts.Serie[string, float64]
	)
	for i := 0; ; i++ {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, nil, err
		}
		pt, err := TimePoint(row[0], row[1])
		if err != nil {
			return nil, nil, err
		}
		ps = append(ps, pt)

		sub := charts.Serie[string, float64]{
			Title: row[0],
			Points: []charts.Point[string, float64]{
				charts.CategoryPoint("open", parseFloat(row[1])),
				charts.CategoryPoint("high", parseFloat(row[2])),
				charts.CategoryPoint("low", parseFloat(row[2])),
				charts.CategoryPoint("close", parseFloat(row[3])),
			},
		}
		ls = append(ls, sub)
	}
	return ps, ls, nil
}

func parseFloat(str string) float64 {
	v, _ := strconv.ParseFloat(str, 64)
	return v
}

func TimePoint(date, value string) (charts.Point[time.Time, float64], error) {
	var pt charts.Point[time.Time, float64]
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return pt, err
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return pt, err
	}
	return charts.TimePoint(t, v), nil
}

func getSerie(name, color string) charts.Serie[time.Time, float64] {
	rdr := charts.LinearRenderer[time.Time, float64]{
		Color: color,
		Text:  charts.TextBefore,
		Point: charts.GetCircle,
	}
	return charts.Serie[time.Time, float64]{
		Title:    name,
		Color:    color,
		Renderer: rdr,
	}
}

func getAxisX(scaler charts.Scaler[time.Time]) charts.Axis {
	return charts.TimeAxis{
		Ticks:          12,
		Rotate:         -45,
		Orientation:    charts.OrientBottom,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithLabelTicks: true,
		WithOuterTicks: false,
		WithBands:      true,
	}
}

func getAxisY(scaler charts.Scaler[float64]) charts.Axis {
	return charts.NumberAxis{
		Ticks: 10,
		// Rotate:         -90,
		Orientation:    charts.OrientRight,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithLabelTicks: true,
		WithOuterTicks: true,
		Format: func(f float64) string {
			return strconv.Itoa(int(f))
		},
		WithBands: false,
	}
}
