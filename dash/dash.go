package dash

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/charts"
	"github.com/midbel/svg"
	"github.com/midbel/svg/layout"
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
	PosTop = "top"
	PosRight = "right"
	PosBottom = "bottom"
	PosLeft = "left"
)

type Renderer interface {
	layout.Renderer
	Render(io.Writer)
}

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

	Style   Style
	Env     *Environ[any]
	Scripts *Environ[ast.Expression]

	Cells []Cell
}

type Cell struct {
	Row    int
	Col    int
	Width  int
	Height int
	Config Config
}

func Default() Config {
	cfg := Config{
		Path:       DefaultPath,
		Width:      DefaultWidth,
		Height:     DefaultHeight,
		TimeFormat: TimeFormat,
		Style:      GlobalStyle(),
		Scripts:    EmptyEnv[ast.Expression](),
	}
	cfg.Types.X = TypeNumber
	cfg.Types.Y = TypeNumber

	return cfg
}

func (c Config) Render() error {
	if len(c.Cells) > 0 {
		return c.renderDashboard()
	}
	rdr, err := c.render()
	if err != nil {
		return err
	}
	w, err := os.Create(c.Path)
	if err != nil {
		return err
	}
	rdr.Render(w)
	return nil
}

func (c Config) render() (Renderer, error) {
	var (
		err error
		mak Renderer
	)
	switch {
	case c.Types.X == TypeNumber && c.Types.Y == TypeNumber:
		mak, err = c.makeNumberChart()
	case c.Types.X == TypeTime && c.Types.Y == TypeNumber:
		mak, err = c.makeTimeChart()
	case c.Types.X == TypeString && c.Types.Y == TypeNumber:
		mak, err = c.makeCategoryChart()
	default:
		err = fmt.Errorf("unsupported chart type %s/%s", c.Types.X, c.Types.Y)
	}
	return mak, err
}

func (c Config) renderDashboard() error {
	var (
		err  error
		grid layout.Grid
	)
	grid = layout.Grid{
		Width:  c.Width,
		Height: c.Height,
	}
	grid.Rows, grid.Cols = c.computeGridDimension()

	for _, cs := range c.Cells {
		cell := layout.Cell{
			X: cs.Row,
			Y: cs.Col,
			W: cs.Width,
			H: cs.Height,
		}
		if cell.Item, err = cs.Config.render(); err != nil {
			return err
		}
		grid.Cells = append(grid.Cells, cell)
	}

	w, err := os.Create(c.Path)
	if err != nil {
		return err
	}
	defer w.Close()

	return grid.Render(w)
}

func (c Config) computeGridDimension() (int, int) {
	var rows, cols int
	for _, e := range c.Cells {
		r := e.Row + e.Height
		if r > rows {
			rows = r
		}
		c := e.Col + e.Width
		if c > cols {
			cols = c
		}
	}
	return rows, cols
}

func (c Config) makeCategoryChart() (Renderer, error) {
	var (
		xrange = c.createRangeX()
		yrange = c.createRangeY()
		chart  = createChart[string, float64](c)
		series = make([]charts.Data, len(c.Files))
	)
	xscale, err := c.Domains.X.makeCategoryScale(xrange)
	if err != nil {
		return nil, err
	}
	yscale, err := c.Domains.Y.makeNumberScale(yrange, true)
	if err != nil {
		return nil, err
	}
	for i := range c.Files {
		series[i], err = c.Files[i].makeCategorySerie(c.Style, xscale, yscale)
		if err != nil {
			return nil, err
		}
	}
	switch c.Domains.X.Position {
	case PosBottom:
		chart.Bottom, err = c.Domains.X.makeCategoryAxis(c, xscale)
	case PosTop:
		chart.Top, err = c.Domains.X.makeCategoryAxis(c, xscale)
	}
	if err != nil {
		return nil, err
	}
	switch c.Domains.Y.Position {
	case PosLeft:
		chart.Left, err = c.Domains.Y.makeNumberAxis(c, yscale)
	case PosRight:
		chart.Right, err = c.Domains.Y.makeNumberAxis(c, yscale)
	}
	if err != nil {
		return nil, err
	}
	return chartRenderer(chart, series), nil
}

func (c Config) makeTimeChart() (Renderer, error) {
	var (
		xrange = c.createRangeX()
		yrange = c.createRangeY()
		chart  = createChart[time.Time, float64](c)
		series = make([]charts.Data, len(c.Files))
	)
	xscale, err := c.Domains.X.makeTimeScale(xrange, false)
	if err != nil {
		return nil, err
	}
	yscale, err := c.Domains.Y.makeNumberScale(yrange, true)
	if err != nil {
		return nil, err
	}
	for i := range c.Files {
		series[i], err = c.Files[i].makeTimeSerie(c.Style, c.TimeFormat, xscale, yscale)
		if err != nil {
			return nil, err
		}
	}
	switch c.Domains.X.Position {
	case PosBottom:
		chart.Bottom, err = c.Domains.X.makeTimeAxis(c, xscale)
	case PosTop:
		chart.Top, err = c.Domains.X.makeTimeAxis(c, xscale)
	}
	if err != nil {
		return nil, err
	}
	switch c.Domains.Y.Position {
	case PosLeft:
		chart.Left, err = c.Domains.Y.makeNumberAxis(c, yscale)
	case PosRight:
		chart.Right, err = c.Domains.Y.makeNumberAxis(c, yscale)
	}
	if err != nil {
		return nil, err
	}
	return chartRenderer(chart, series), nil
}

func (c Config) makeNumberChart() (Renderer, error) {
	var (
		xrange = c.createRangeX()
		yrange = c.createRangeY()
		chart  = createChart[float64, float64](c)
		series = make([]charts.Data, len(c.Files))
	)
	xscale, err := c.Domains.X.makeNumberScale(xrange, false)
	if err != nil {
		return nil, err
	}
	yscale, err := c.Domains.Y.makeNumberScale(yrange, true)
	if err != nil {
		return nil, err
	}
	for i := range c.Files {
		series[i], err = c.Files[i].makeNumberSerie(c.Style, xscale, yscale)
		if err != nil {
			return nil, err
		}
	}
	switch c.Domains.X.Position {
	case PosBottom:
		chart.Bottom, err = c.Domains.X.makeNumberAxis(c, xscale)
	case PosTop:
		chart.Top, err = c.Domains.X.makeNumberAxis(c, xscale)
	}
	if err != nil {
		return nil, err
	}
	switch c.Domains.Y.Position {
	case PosLeft:
		chart.Left, err = c.Domains.Y.makeNumberAxis(c, yscale)
	case PosRight:
		chart.Right, err = c.Domains.Y.makeNumberAxis(c, yscale)
	}
	if err != nil {
		return nil, err
	}
	return chartRenderer(chart, series), nil
}

func (c Config) createRangeX() charts.Range {
	return charts.NewRange(0, c.Width-c.Pad.Left-c.Pad.Right)
}

func (c Config) createRangeY() charts.Range {
	return charts.NewRange(0, c.Height-c.Pad.Top-c.Pad.Bottom)
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

type chartMaker struct {
	charts.Drawner
	series []charts.Data
}

func chartRenderer(ch charts.Drawner, series []charts.Data) Renderer {
	return chartMaker{
		Drawner: ch,
		series:  series,
	}
}

func (c chartMaker) Element() svg.Element {
	return c.Drawn(c.series...)
}

func (c chartMaker) Render(w io.Writer) {
	ws := bufio.NewWriter(w)
	defer ws.Flush()

	el := c.Element()
	el.Render(ws)
}
