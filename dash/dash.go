package dash

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/charts"
	"github.com/midbel/svg"
	"github.com/midbel/svg/layout"
)

//go:embed themes/*css
var themes embed.FS

const themedir = "themes"

var (
	DefaultWidth  = 800.0
	DefaultHeight = 600.0

	TimeFormat   = "%Y-%m-%d"
	DefaultPath  = "out.svg"
	DefaultDelim = ","
)

const (
	TypeNumber = "number"
	TypeTime   = "time"
	TypeString = "string"
)

const (
	PosTop    = "top"
	PosRight  = "right"
	PosBottom = "bottom"
	PosLeft   = "left"
)

type Renderer interface {
	layout.Renderer
	Render(io.Writer)
}

type Legend struct {
	Title    string
	Position []string
}

type Cell struct {
	Row    int
	Col    int
	Width  int
	Height int
	Config Config
}

func MakeCell(c Config) Cell {
	empty := Cell{
		Width:  1,
		Height: 1,
		Config: c,
	}
	empty.Config.Files = nil
	empty.Config.Cells = nil
	return empty
}

type Config struct {
	Title string
	Legend

	Path string

	Width  float64
	Height float64
	Rows   int
	Cols   int
	Pad    charts.Padding

	Delimiter  string
	TimeFormat string

	X     Input
	Y     Input
	Files []File

	Style   Style
	Env     *Environ[any]
	Scripts *Environ[ast.Expression]

	Cells []Cell

	Theme string
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
	cfg.X.Type = TypeNumber
	cfg.Y.Type = TypeNumber

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
	var w io.Writer = os.Stdout
	if c.Path != "" {
		f, err := os.Create(c.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	rdr.Render(w)
	return nil
}

func (c Config) render() (Renderer, error) {
	var (
		err   error
		maker Renderer
	)
	switch {
	case c.X.isNumber() && c.Y.isNumber():
		maker, err = c.numberChart()
	case c.X.isTime() && c.Y.isNumber():
		maker, err = c.timeChart()
	case c.X.isString() && c.Y.isNumber():
		maker, err = c.categoryChart()
	default:
		err = fmt.Errorf("unsupported chart type %s/%s", c.X.Type, c.Y.Type)
	}
	return maker, err
}

func (c Config) renderDashboard() error {
	var (
		err  error
		grid layout.Grid
	)
	grid = layout.Grid{
		Width:  c.Width,
		Height: c.Height,
		Rows:   c.Rows,
		Cols:   c.Cols,
	}
	if c.Rows == 0 && c.Cols == 0 {
		grid.Rows, grid.Cols = c.computeGridDimension()
	}

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

func (c Config) categoryChart() (Renderer, error) {
	var (
		xrange = c.createRangeX()
		yrange = c.createRangeY()
		chart  = createChart[string, float64](c)
		series = make([]charts.Data, len(c.Files))
	)
	xscale, err := c.X.CategoryScale(xrange)
	if err != nil {
		return nil, err
	}
	yscale, err := c.Y.NumberScale(yrange, true)
	if err != nil {
		return nil, err
	}
	for i := range c.Files {
		series[i], err = c.Files[i].CategorySerie(c.Style, xscale, yscale)
		if err != nil {
			return nil, err
		}
	}
	switch c.X.Position {
	case PosBottom:
		chart.Bottom, err = c.X.GetCategoryAxis(c, xscale)
	case PosTop:
		chart.Top, err = c.X.GetCategoryAxis(c, xscale)
	}
	if err != nil {
		return nil, err
	}
	switch c.Y.Position {
	case PosLeft:
		chart.Left, err = c.Y.GetNumberAxis(c, yscale)
	case PosRight:
		chart.Right, err = c.Y.GetNumberAxis(c, yscale)
	}
	if err != nil {
		return nil, err
	}
	return chartRenderer(chart, series), nil
}

func (c Config) timeChart() (Renderer, error) {
	var (
		xrange = c.createRangeX()
		yrange = c.createRangeY()
		chart  = createChart[time.Time, float64](c)
		series = make([]charts.Data, len(c.Files))
	)
	xscale, err := c.X.TimeScale(xrange, TimeFormat, false)
	if err != nil {
		return nil, err
	}
	yscale, err := c.Y.NumberScale(yrange, true)
	if err != nil {
		return nil, err
	}
	for i := range c.Files {
		series[i], err = c.Files[i].TimeSerie(c.Style, c.TimeFormat, xscale, yscale)
		if err != nil {
			return nil, err
		}
	}
	switch c.X.Position {
	case PosBottom:
		chart.Bottom, err = c.X.GetTimeAxis(c, xscale)
	case PosTop:
		chart.Top, err = c.X.GetTimeAxis(c, xscale)
	}
	if err != nil {
		return nil, err
	}
	switch c.Y.Position {
	case PosLeft:
		chart.Left, err = c.Y.GetNumberAxis(c, yscale)
	case PosRight:
		chart.Right, err = c.Y.GetNumberAxis(c, yscale)
	}
	if err != nil {
		return nil, err
	}
	return chartRenderer(chart, series), nil
}

func (c Config) numberChart() (Renderer, error) {
	var (
		xrange = c.createRangeX()
		yrange = c.createRangeY()
		chart  = createChart[float64, float64](c)
		series = make([]charts.Data, len(c.Files))
	)
	xscale, err := c.X.NumberScale(xrange, false)
	if err != nil {
		return nil, err
	}
	yscale, err := c.Y.NumberScale(yrange, true)
	if err != nil {
		return nil, err
	}
	for i := range c.Files {
		series[i], err = c.Files[i].NumberSerie(c.Style, xscale, yscale)
		if err != nil {
			return nil, err
		}
	}
	switch c.X.Position {
	case PosBottom:
		chart.Bottom, err = c.X.GetNumberAxis(c, xscale)
	case PosTop:
		chart.Top, err = c.X.GetNumberAxis(c, xscale)
	}
	if err != nil {
		return nil, err
	}
	switch c.Y.Position {
	case PosLeft:
		chart.Left, err = c.Y.GetNumberAxis(c, yscale)
	case PosRight:
		chart.Right, err = c.Y.GetNumberAxis(c, yscale)
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
	if dat, err := themes.ReadFile(filepath.Join(themedir, cfg.Theme) + ".css"); err == nil {
		ch.Theme = string(dat)
	} else if dat, err = os.ReadFile(cfg.Theme); err == nil {
		ch.Theme = string(dat)
	} else {
		ch.Theme = cfg.Theme
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
