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
	Top:    10,
	Right:  45,
	Bottom: 60,
	Left:   40,
}

func main() {
	skip := flag.Int("s", 0, "keep N values")
	flag.Parse()
	var (
		dtstart    = time.Date(2018, 9, 1, 0, 0, 0, 0, time.UTC)
		dtend      = time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)
		timeScale  = charts.TimeScaler(charts.TimeDomain(dtstart, dtend), charts.NewRange(0, defaultWidth-pad.Horizontal()))
		priceScale = charts.NumberScaler(charts.NumberDomain(350, 0), charts.NewRange(0, defaultHeight-pad.Vertical()))
		series     []charts.Data
		colors     = []string{"red", "green", "blue", "slategrey"}
	)
	for i, file := range flag.Args() {
		s, err := loadSerie(file, colors[i%len(colors)], *skip)
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
		Left:    getAxisY(priceScale),
		Bottom:  getAxisX(timeScale),
	}
	ch.Render(os.Stdout, series...)
}

func loadSerie(file, color string, skip int) (charts.Serie[time.Time, float64], error) {
	var (
		name = strings.TrimRight(filepath.Base(file), filepath.Ext(file))
		ser  = getSerie(name, color, skip)
		err  error
	)
	ser.Points, err = loadPoints(file, skip)
	return ser, err
}

func loadPoints(file string, skip int) ([]charts.Point[time.Time, float64], error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var (
		rs = csv.NewReader(r)
		ps []charts.Point[time.Time, float64]
	)
	rs.Read()
	for i := 0; ; i++ {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if skip > 0 && i%skip != 0 {
			continue
		}
		pt, err := TimePoint(row[0], row[1])
		if err != nil {
			return nil, err
		}
		ps = append(ps, pt)
	}
	return ps, nil
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

func getSerie(name, color string, skip int) charts.Serie[time.Time, float64] {
	rdr := charts.LinearRenderer[time.Time, float64]{
		Color: color,
		Skip:  skip,
		Text:  charts.TextAfter,
	}
	if skip > 10 {
		rdr.Point = charts.GetCircle
	}
	return charts.Serie[time.Time, float64]{
		Title:    name,
		Color:    color,
		Renderer: rdr,
	}
}

func getAxisX(scaler charts.Scaler[time.Time]) charts.Axis {
	return charts.TimeAxis{
		Ticks:          7,
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
		Orientation:    charts.OrientLeft,
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
