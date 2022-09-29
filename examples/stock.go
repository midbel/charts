package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/midbel/charts"
)

const (
	defaultWidth  = 800
	defaultHeight = 600
)

var pad = charts.Padding{
	Top:    20,
	Right:  60,
	Bottom: 40,
	Left:   60,
}

func main() {
	flag.Parse()
	var (
		dtstart    = time.Date(2018, 9, 1, 0, 0, 0, 0, time.UTC)
		dtend      = time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)
		timeScale  = charts.TimeScaler(charts.TimeDomain(dtstart, dtend), charts.NewRange(0, defaultWidth-pad.Horizontal()))
		priceScale = charts.NumberScaler(charts.NumberDomain(350, 0), charts.NewRange(0, defaultHeight-pad.Vertical()))
		series     []charts.Serie[time.Time, float64]
		colors     = []string{"red", "green", "blue", "slategrey"}
	)
	for i, file := range flag.Args() {
		s, err := loadSerie(file, colors[i%len(colors)])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		s.X = timeScale
		s.Y = priceScale
		series = append(series, s)
	}

	ch := charts.Chart[time.Time, float64]{
		Width:   defaultWidth,
		Height:  defaultHeight,
		Padding: pad,
		Left:    getLeftAxis(priceScale),
		Bottom:  getBottomAxis(timeScale),
	}
	ch.Render(os.Stdout, series...)
}

func loadSerie(file, color string) (charts.Serie[time.Time, float64], error) {
	name := strings.TrimRight(filepath.Base(file), filepath.Ext(file))
	ser := charts.Serie[time.Time, float64]{
		Title:    name,
		Color:    color,
		Renderer: charts.LinearRender[time.Time, float64](false),
	}

	r, err := os.Open(file)
	if err != nil {
		return ser, err
	}
	defer r.Close()

	rs := csv.NewReader(r)
	rs.Read()
	for i := 0; ; i++ {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return ser, err
		}
		pt, err := TimePoint(row[0], row[1])
		if err != nil {
			return ser, err
		}
		ser.Points = append(ser.Points, pt)
	}
	return ser, nil
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

func getBottomAxis(scaler charts.Scaler[time.Time]) charts.Axis {
	return charts.TimeAxis{
		Ticks:          5,
		Orientation:    charts.OrientBottom,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithLabelTicks: true,
		WithOuterTicks: false,
	}
}

func getLeftAxis(scaler charts.Scaler[float64]) charts.Axis {
	return charts.NumberAxis{
		Ticks:          10,
		Orientation:    charts.OrientLeft,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithLabelTicks: true,
		WithOuterTicks: true,
		Format: func(f float64) string {
			return strconv.Itoa(int(f))
		},
	}
}
