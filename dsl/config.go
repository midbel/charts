package dsl

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/midbel/charts"
	"github.com/midbel/slices"
	"golang.org/x/sync/errgroup"
)

var (
	DefaultWidth  = 800.0
	DefaultHeight = 600.0

	TimeFormat   = "%y-%m-%d"
	DefaultPath  = "out.svg"
	DefaultDelim = ","
)

const (
	TypeNumber = "number"
	TypeTime   = "time"
	TypeString = "string"
)

const (
	StyleStraight = "straight"
	StyleDotted   = "dotted"
	StyleDashed   = "dashed"
)

const (
	RenderLine       = "line"
	RenderPie        = "pie"
	RenderBar        = "bar"
	RenderStep       = "step"
	RenderStepAfter  = "step-after"
	RenderStepBefore = "step-before"
)

type Config struct {
	Title  string
	Path   string
	Width  float64
	Height float64
	Pad    struct {
		Top    float64
		Right  float64
		Bottom float64
		Left   float64
	}
	Delimiter  string
	TimeFormat string
	Types      struct {
		X string
		Y string
	}
	Domains struct {
		X Domain
		Y Domain
	}
	Center struct {
		X string
		Y string
	}
	Legend struct {
		Title    string
		Position []string
	}
	Files []File

	Style Style
}

func Default() Config {
	cfg := Config{
		Path:       DefaultPath,
		Width:      DefaultWidth,
		Height:     DefaultHeight,
		TimeFormat: TimeFormat,
		Style:      GlobalStyle(),
	}
	cfg.Types.X = TypeNumber
	cfg.Types.Y = TypeNumber

	return cfg
}

func (c Config) Render() error {
	var err error
	switch {
	case c.Types.X == TypeNumber && c.Types.Y == TypeNumber:
		err = c.renderNumberChart()
	case c.Types.X == TypeTime && c.Types.Y == TypeNumber:
		err = c.renderTimeChart()
	default:
		err = fmt.Errorf("unsupported chart type %s/%s", c.Types.X, c.Types.Y)
	}
	return err
}

func (c Config) renderTimeChart() error {
	var (
		xrange = c.createRangeX()
		yrange = c.createRangeY()
		chart  = createChart[time.Time, float64](c)
		series = make([]charts.Data, len(c.Files))
	)
	xscale, err := c.Domains.X.makeTimeScale(xrange, false)
	if err != nil {
		return err
	}
	yscale, err := c.Domains.Y.makeNumberScale(yrange, true)
	if err != nil {
		return err
	}
	var grp errgroup.Group
	for i := range c.Files {
		s, j := c.Files[i], i
		grp.Go(func() (err error) {
			series[j], err = s.makeTimeSerie(c.Style, c.TimeFormat, xscale, yscale)
			return err
		})
	}
	if err := grp.Wait(); err != nil {
		return err
	}
	switch c.Domains.X.Position {
	case "bottom":
		chart.Bottom, err = c.Domains.X.makeTimeAxis(xscale)
	case "top":
		chart.Top, err = c.Domains.X.makeTimeAxis(xscale)
	}
	if err != nil {
		return err
	}
	switch c.Domains.Y.Position {
	case "left":
		chart.Left, err = c.Domains.Y.makeNumberAxis(yscale)
	case "right":
		chart.Right, err = c.Domains.Y.makeNumberAxis(yscale)
	}
	if err != nil {
		return err
	}
	return renderChart(c.Path, chart, series)
}

func (c Config) renderNumberChart() error {
	var (
		xrange = c.createRangeX()
		yrange = c.createRangeY()
		chart  = createChart[float64, float64](c)
		series = make([]charts.Data, len(c.Files))
	)
	xscale, err := c.Domains.X.makeNumberScale(xrange, false)
	if err != nil {
		return err
	}
	yscale, err := c.Domains.Y.makeNumberScale(yrange, true)
	if err != nil {
		return err
	}
	if pt, ok, err := c.getNumberCenter(); ok && err == nil {
		chart.Center = pt
	}
	var grp errgroup.Group
	for i := range c.Files {
		s, j := c.Files[i], i
		grp.Go(func() (err error) {
			series[j], err = s.makeNumberSerie(c.Style, xscale, yscale)
			return err
		})
	}
	if err := grp.Wait(); err != nil {
		return err
	}
	switch c.Domains.X.Position {
	case "bottom":
		chart.Bottom, err = c.Domains.X.makeNumberAxis(xscale)
	case "top":
		chart.Top, err = c.Domains.X.makeNumberAxis(xscale)
	}
	if err != nil {
		return err
	}
	switch c.Domains.Y.Position {
	case "left":
		chart.Left, err = c.Domains.Y.makeNumberAxis(yscale)
	case "right":
		chart.Right, err = c.Domains.Y.makeNumberAxis(yscale)
	}
	if err != nil {
		return err
	}
	return renderChart(c.Path, chart, series)
}

func (c Config) getNumberCenter() (charts.Point[float64, float64], bool, error) {
	var (
		pt  charts.Point[float64, float64]
		err error
	)
	if c.Center.X == "" || c.Center.Y == "" {
		return pt, false, err
	}
	pt.X, err = strconv.ParseFloat(c.Center.X, 64)
	if err != nil {
		return pt, false, err
	}
	pt.Y, err = strconv.ParseFloat(c.Center.Y, 64)
	if err != nil {
		return pt, false, err
	}
	return pt, true, nil
}

func (c Config) createRangeX() charts.Range {
	return charts.NewRange(0, c.Width-c.Pad.Left-c.Pad.Right)
}

func (c Config) createRangeY() charts.Range {
	return charts.NewRange(0, c.Height-c.Pad.Top-c.Pad.Bottom)
}

type Domain struct {
	Label      string
	Ticks      int
	Format     string
	Domain     []string
	Position   string
	InnerTicks bool
	OuterTicks bool
	LabelTicks bool
	BandTicks  bool
}

func (d Domain) makeNumberScale(rg charts.Range, reverse bool) (charts.Scaler[float64], error) {
	if len(d.Domain) == 0 {
		return nil, fmt.Errorf("domain not set")
	}
	fst, err := strconv.ParseFloat(slices.Fst(d.Domain), 64)
	if err != nil {
		return nil, err
	}
	lst, err := strconv.ParseFloat(slices.Lst(d.Domain), 64)
	if err != nil {
		return nil, err
	}
	if reverse {
		fst, lst = lst, fst
	}
	return charts.NumberScaler(charts.NumberDomain(fst, lst), rg), nil
}

func (d Domain) makeTimeScale(rg charts.Range, reverse bool) (charts.Scaler[time.Time], error) {
	if len(d.Domain) == 0 {
		return nil, fmt.Errorf("domain not set")
	}
	parseTime, err := makeParseTime(d.Format)
	if err != nil {
		return nil, err
	}
	fst, err := parseTime(slices.Fst(d.Domain))
	if err != nil {
		return nil, err
	}
	lst, err := parseTime(slices.Lst(d.Domain))
	if err != nil {
		return nil, err
	}
	if reverse {
		fst, lst = lst, fst
	}
	return charts.TimeScaler(charts.TimeDomain(fst, lst), rg), nil
}

func (d Domain) makeNumberAxis(scale charts.Scaler[float64]) (charts.Axis[float64], error) {
	axe := charts.Axis[float64]{
		Label:          d.Label,
		Ticks:          d.Ticks,
		Scaler:         scale,
		WithInnerTicks: d.InnerTicks,
		WithOuterTicks: d.OuterTicks,
		WithLabelTicks: d.LabelTicks,
		WithBands:      d.BandTicks,
		Format: func(f float64) string {
			return fmt.Sprintf(d.Format, f)
		},
	}
	return axe, nil
}

func (d Domain) makeTimeAxis(scale charts.Scaler[time.Time]) (charts.Axis[time.Time], error) {
	formatTime, err := makeTimeFormat(d.Format)
	if err != nil {
		return charts.Axis[time.Time]{}, err
	}
	axe := charts.Axis[time.Time]{
		Label:          d.Label,
		Ticks:          d.Ticks,
		Scaler:         scale,
		WithInnerTicks: d.InnerTicks,
		WithOuterTicks: d.OuterTicks,
		WithLabelTicks: d.LabelTicks,
		WithBands:      d.BandTicks,
		Format:         formatTime,
	}
	return axe, nil
}

type Style struct {
	Type          string
	Stroke        string
	Fill          bool
	Point         string
	Width         float64
	InnerRadius   float64
	OuterRadius   float64
	IgnoreMissing bool
	TextPosition  string
	LineStyle     string
}

func GlobalStyle() Style {
	return Style{
		Type:   RenderLine,
		Stroke: "black",
		Fill:   false,
	}
}

func (s Style) getTextPosition() charts.TextPosition {
	var pos charts.TextPosition
	switch s.TextPosition {
	case "text-before":
		pos = charts.TextBefore
	case "text-after":
		pos = charts.TextAfter
	default:
	}
	return pos
}

func (s Style) getPointFunc() charts.PointFunc {
	switch s.Point {
	case "circle":
		return charts.GetCircle
	case "square":
		return charts.GetSquare
	default:
		return nil
	}
}

func (s Style) getLineStyle() charts.LineStyle {
	var i charts.LineStyle
	switch s.LineStyle {
	case "", StyleStraight:
		i = charts.StyleStraight
	case StyleDotted:
		i = charts.StyleDotted
	case StyleDashed:
		i = charts.StyleDashed
	}
	return i
}

func (s Style) makeTimeRenderer(g Style) (charts.Renderer[time.Time, float64], error) {
	return createRenderer[time.Time, float64](s.merge(g))
}

func (s Style) makeNumberRenderer(g Style) (charts.Renderer[float64, float64], error) {
	return createRenderer[float64, float64](s.merge(g))
}

func (s Style) merge(g Style) Style {
	if s.Type == "" {
		s.Type = g.Type
	}
	if s.Stroke == "" {
		s.Stroke = g.Stroke
	}
	if s.Point == "" {
		s.Point = g.Point
	}
	if s.InnerRadius == 0 && g.InnerRadius != 0 {
		s.InnerRadius = g.InnerRadius
	}
	if s.OuterRadius == 0 && g.OuterRadius != 0 {
		s.OuterRadius = g.OuterRadius
	}
	if s.TextPosition == "" {
		s.TextPosition = g.TextPosition
	}
	return s
}

type File struct {
	Path       string
	Ident      string
	X          int
	Y          int
	TimeFormat string
	Starts     int
	Ends       int
	Style
}

func (f File) Name() string {
	if f.Ident != "" {
		return f.Ident
	}
	return strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
}

func (f File) makeTimeSerie(g Style, timefmt string, x charts.Scaler[time.Time], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := f.makeTimeRenderer(g)
	if err != nil {
		return nil, err
	}
	if f.TimeFormat == "" {
		f.TimeFormat = timefmt
	}
	parseTime, err := makeParseTime(f.TimeFormat)
	if err != nil {
		return nil, err
	}

	points, err := loadTimePoints(f, parseTime)
	if err != nil {
		return nil, err
	}
	ser := charts.Serie[time.Time, float64]{
		Title:    f.Name(),
		Renderer: rdr,
		Points:   points,
		X:        x,
		Y:        y,
	}
	return ser, err
}

func (f File) makeNumberSerie(g Style, x charts.Scaler[float64], y charts.Scaler[float64]) (charts.Data, error) {
	rdr, err := f.makeNumberRenderer(g)
	if err != nil {
		return nil, err
	}
	points, err := loadNumberPoints(f)
	if err != nil {
		return nil, err
	}
	ser := charts.Serie[float64, float64]{
		Title:    f.Name(),
		Renderer: rdr,
		Points:   points,
		X:        x,
		Y:        y,
	}
	return ser, nil
}

type PointFunc[T, U charts.ScalerConstraint] func([]string) (charts.Point[T, U], error)

func loadPoints[T, U charts.ScalerConstraint](file string, get PointFunc[T, U]) ([]charts.Point[T, U], error) {
	r, err := readFrom(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var (
		rs   = csv.NewReader(r)
		list []charts.Point[T, U]
	)
	rs.Read()
	for {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		pt, err := get(row)
		if err != nil {
			return nil, err
		}
		list = append(list, pt)
	}
	return list, nil
}

func loadNumberPoints(f File) ([]charts.Point[float64, float64], error) {
	get := func(row []string) (charts.Point[float64, float64], error) {
		var (
			pt  charts.Point[float64, float64]
			err error
		)
		pt.X, err = strconv.ParseFloat(row[f.X], 64)
		if err != nil {
			return pt, err
		}
		pt.Y, err = strconv.ParseFloat(row[f.Y], 64)
		if err != nil {
			return pt, err
		}
		return pt, nil
	}
	return loadPoints[float64, float64](f.Path, get)
}

func loadTimePoints(f File, parseTime func(string) (time.Time, error)) ([]charts.Point[time.Time, float64], error) {
	get := func(row []string) (charts.Point[time.Time, float64], error) {
		var (
			pt  charts.Point[time.Time, float64]
			err error
		)
		pt.X, err = parseTime(row[f.X])
		if err != nil {
			return pt, err
		}
		pt.Y, err = strconv.ParseFloat(row[f.Y], 64)
		if err != nil {
			return pt, err
		}
		return pt, nil
	}
	return loadPoints[time.Time, float64](f.Path, get)
}

func readFrom(location string) (io.ReadCloser, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http", "https":
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("request does not end with success result code")
		}
		return res.Body, nil
	case "", "file":
		return os.Open(u.Path)
	default:
		return nil, fmt.Errorf("%s: unsupported scheme", u.Scheme)
	}
}

func renderChart[T, U charts.ScalerConstraint](file string, chart charts.Chart[T, U], series []charts.Data) error {
	if len(series) == 0 {
		return nil
	}
	w, err := os.Create(file)
	if err != nil {
		return err
	}
	defer w.Close()
	chart.Render(w, series...)
	return nil
}

func createRenderer[T, U charts.ScalerConstraint](style Style) (charts.Renderer[T, U], error) {
	var rdr charts.Renderer[T, U]
	switch style.Type {
	case RenderLine:
		rdr = charts.LinearRenderer[T, U]{
			Color:         style.Stroke,
			IgnoreMissing: style.IgnoreMissing,
			Text:          style.getTextPosition(),
			Point:         style.getPointFunc(),
			Style:         style.getLineStyle(),
		}
	case RenderStep:
		rdr = charts.StepRenderer[T, U]{
			Color:         style.Stroke,
			IgnoreMissing: style.IgnoreMissing,
			Text:          style.getTextPosition(),
			Point:         style.getPointFunc(),
			Style:         style.getLineStyle(),
		}
	case RenderStepAfter:
		rdr = charts.StepAfterRenderer[T, U]{
			Color:         style.Stroke,
			IgnoreMissing: style.IgnoreMissing,
			Text:          style.getTextPosition(),
			Point:         style.getPointFunc(),
			Style:         style.getLineStyle(),
		}
	case RenderStepBefore:
		rdr = charts.StepBeforeRenderer[T, U]{
			Color:         style.Stroke,
			IgnoreMissing: style.IgnoreMissing,
			Text:          style.getTextPosition(),
			Point:         style.getPointFunc(),
			Style:         style.getLineStyle(),
		}
	case RenderBar:
		rdr = charts.BarRenderer[T, U]{
			Fill:  charts.Tableau10,
			Width: style.Width,
		}
	case RenderPie:
		rdr = charts.PieRenderer[T, U]{
			Fill:        charts.Tableau10,
			InnerRadius: style.InnerRadius,
			OuterRadius: style.OuterRadius,
		}
	case "sunburst":
	case "stacked":
		rdr = charts.StackedRenderer[T, U]{
			Fill:      charts.Tableau10,
			Width:     style.Width,
			Normalize: false,
		}
	case "group":
	default:
		return nil, fmt.Errorf("%s: can not use for number chart", style.Type)
	}
	return rdr, nil
}

func createChart[T, U charts.ScalerConstraint](cfg Config) charts.Chart[T, U] {
	ch := charts.Chart[T, U]{
		Title:  cfg.Title,
		Width:  cfg.Width,
		Height: cfg.Height,
		Padding: charts.Padding{
			Top:    cfg.Pad.Top,
			Right:  cfg.Pad.Right,
			Bottom: cfg.Pad.Bottom,
			Left:   cfg.Pad.Left,
		},
	}
	ch.Legend.Title = cfg.Legend.Title
	for _, p := range cfg.Legend.Position {
		switch p {
		case "top":
			ch.Legend.Orient |= charts.OrientTop
		case "bottom":
			ch.Legend.Orient |= charts.OrientBottom
		case "right":
			ch.Legend.Orient |= charts.OrientRight
		case "left":
			ch.Legend.Orient |= charts.OrientLeft
		default:
			// pass or returns an error???
		}
	}
	return ch
}
