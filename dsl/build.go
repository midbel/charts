package dsl

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/midbel/charts"
	"github.com/midbel/slices"
)

const (
	kwWith   = "with"
	kwScale  = "scale"
	kwAxis   = "axis"
	kwChart  = "chart"
	kwSerie  = "serie"
	kwLine   = "line"
	kwPie    = "pie"
	kwCsv    = "csv"
	kwRender = "render"
)

const (
	powScale int = iota
	powAxis
	powRender
	powData
	powSerie
	powChart
	powExec
)

var priorities = map[string]int{
	kwScale:  powScale,
	kwAxis:   powAxis,
	kwLine:   powRender,
	kwPie:    powRender,
	kwCsv:    powData,
	kwSerie:  powSerie,
	kwChart:  powChart,
	kwRender: powExec,
}

const (
	typeNumber = "number"
	typeTime   = "time"
	typeString = "string"
)

const (
	defaultWidth  = 800.0
	defaultHeight = 600.0
)

type Builder struct {
	scan *Scanner
	curr Token
	peek Token

	scales    map[string]any
	axis      map[string]any
	charts    map[string]any
	renderers map[string]any
	series    map[string]any

	charttype struct {
		X string
		Y string
	}
}

func New(r io.Reader) *Builder {
	b := Builder{
		scan:      Scan(r),
		scales:    make(map[string]any),
		axis:      make(map[string]any),
		charts:    make(map[string]any),
		renderers: make(map[string]any),
		series:    make(map[string]any),
	}
	b.next()
	b.next()
	return &b
}

func (b *Builder) Build() error {
	var list []*command
	for !b.done() {
		cmd, err := b.prepare()
		if err != nil {
			return err
		}
		list = append(list, cmd)
	}
	if len(list) == 0 {
		return nil
	}
	sort.Slice(list, func(i, j int) bool {
		return priorities[list[i].Name] < priorities[list[i].Name]
	})
	for i := range list {
		fmt.Println(list[i].Name, list[i].Ident)
		if err := b.execute(list[i]); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) execute(cmd *command) error {
	sort.Slice(cmd.Options, func(i, j int) bool {
		return cmd.Options[i].Name > cmd.Options[j].Name
	})
	var err error
	switch cmd.Name {
	default:
		err = fmt.Errorf("%s: unknown command", cmd.Name)
	case kwScale:
		err = b.executeScale(cmd.Ident, cmd.Options)
	case kwAxis:
		err = b.executeAxis(cmd.Ident, cmd.Options)
	case kwChart:
		err = b.executeChart(cmd.Ident, cmd.Options)
	case kwSerie:
		err = b.executeSerie(cmd.Ident, cmd.Options)
	case kwLine:
		err = b.executeLine(cmd.Ident, cmd.Options)
	case kwPie:
		err = b.executePie(cmd.Ident, cmd.Options)
	case kwCsv:
		err = b.executeCsv(cmd.Ident, cmd.Options)
	case kwRender:
		err = b.executeRender(cmd.Ident, cmd.Options)
	}
	return err
}

func (b *Builder) prepare() (*command, error) {
	cmd := command{
		Name: b.curr.Literal,
	}
	b.next()
	b.next()
	cmd.Ident = b.curr.Literal
	b.next()
	b.next()
	switch b.curr.Type {
	case EOL:
		b.next()
		return &cmd, nil
	case Literal:
		if b.curr.Literal != kwWith {
			return nil, fmt.Errorf("missing with keyword")
		}
		b.next()
		b.next()
	default:
		return nil, fmt.Errorf("unexpected token %s", b.curr)
	}
	for b.curr.Type != EOL && b.curr.Type != EOF {
		opt := option{
			Name: b.curr.Literal,
		}
		b.next()
		b.next()
		var values []any
		for b.curr.Type != Blank && b.curr.Type != EOL && b.curr.Type != EOF {
			switch b.curr.Type {
			case Literal, Reference:
				values = append(values, b.curr.Literal)
			case Number:
				n, err := strconv.ParseFloat(b.curr.Literal, 64)
				if err != nil {
					return nil, err
				}
				values = append(values, n)
			case Boolean:
				b, err := strconv.ParseBool(b.curr.Literal)
				if err != nil {
					return nil, err
				}
				values = append(values, b)
			case Date:
				t, err := time.Parse("2006-01-02", b.curr.Literal)
				if err != nil {
					return nil, err
				}
				values = append(values, t)
			case Command:
			default:
				return nil, fmt.Errorf("unexpected token %s", b.curr)
			}
			b.next()
			if b.curr.Type == Comma {
				b.next()
			}
		}
		opt.Value = values
		if len(values) == 1 {
			opt.Value = values[0]
		}
		if b.curr.Type == Blank {
			b.next()
		}
		cmd.Options = append(cmd.Options, opt)
	}
	b.next()
	return &cmd, nil
}

func (b *Builder) next() {
	b.curr = b.peek
	b.peek = b.scan.Scan()
}

func (b *Builder) done() bool {
	return b.curr.Type == EOF
}

func (b *Builder) executeRender(ident string, options []option) error {
	ch, ok := b.charts[getValue("chart", "", options)]
	if !ok {
		return fmt.Errorf("chart not found")
	}
	w, err := os.Create(ident)
	if err != nil {
		return err
	}
	defer w.Close()

	r, ok := ch.(interface {
		Render(io.Writer, ...charts.Data)
	})
	if !ok {
		return nil
	}
	r.Render(w)
	return nil
}

func (b *Builder) executeScale(ident string, options []option) error {
	r, err := findOption("range", options)
	if err != nil {
		return fmt.Errorf("range is not defined")
	}
	rg := createRange(r.Value)
	fmt.Printf("%+v\n", rg)

	x, err := findOption("domain", options)
	if err != nil {
		return fmt.Errorf("domain is not defined")
	}
	var scale any
	switch t := getType(options); t {
	case typeNumber, "":
		d := createNumberDomain(x.Value)
		scale = charts.NumberScaler(d, rg)
	case typeTime:
		d := createTimeDomain(x.Value)
		scale = charts.TimeScaler(d, rg)
	case typeString:
	default:
		return fmt.Errorf("%s: invalid value for type option", t)
	}
	b.scales[ident] = scale
	return nil
}

func (b *Builder) executeSerie(ident string, options []option) error {
	return nil
}

func (b *Builder) executeAxis(ident string, options []option) error {
	var axis any
	switch t := getType(options); t {
	case typeNumber:
		x := charts.Axis[float64]{
			Label:          getValue("legend", "", options),
			Ticks:          int(getValue("ticks", 10.0, options)),
			WithInnerTicks: getValue("inner-ticks", true, options),
			WithOuterTicks: getValue("outer-ticks", false, options),
			WithLabelTicks: getValue("label-ticks", true, options),
			WithBands:      getValue("bands-ticks", false, options),
			Format: func(f float64) string {
				return strconv.FormatFloat(f, 'f', 2, 64)
			},
		}
		ident := getValue("scale", "", options)
		if ident == "" {
			return fmt.Errorf("no scale defined")
		}
		s, ok := b.scales[ident]
		if !ok {
			return fmt.Errorf("scale %s not defined", ident)
		}
		x.Scaler, ok = s.(charts.Scaler[float64])
		if !ok {
			return fmt.Errorf("%s can not be used as scaler for axis - wrong type", ident)
		}
		fmt.Printf("axis(float): %+v\n", x)
		axis = x
	case typeTime:
		x := charts.Axis[time.Time]{
			Label:          getValue("legend", "", options),
			Ticks:          int(getValue("ticks", 10.0, options)),
			WithInnerTicks: getValue("inner-ticks", true, options),
			WithOuterTicks: getValue("outer-ticks", false, options),
			WithLabelTicks: getValue("label-ticks", true, options),
			WithBands:      getValue("bands-ticks", false, options),
			Format: func(t time.Time) string {
				return t.Format("2006-01-02")
			},
		}
		ident := getValue("scale", "", options)
		if ident == "" {
			return fmt.Errorf("no scale defined")
		}
		s, ok := b.scales[ident]
		if !ok {
			return fmt.Errorf("scale %s not defined", ident)
		}
		x.Scaler, ok = s.(charts.Scaler[time.Time])
		if !ok {
			return fmt.Errorf("%s can not be used as scaler for axis - wrong type", ident)
		}
		fmt.Printf("axis(time): %+v\n", x)
		axis = x
	case typeString:
	default:
		return fmt.Errorf("%s: invalid value for type option", t)
	}
	b.axis[ident] = axis
	return nil
}

func (b *Builder) executeChart(ident string, options []option) error {
	left, bottom := b.guessChartType(options)
	pad, _ := findOption("padding", options)
	var chart any
	switch {
	case left == typeNumber && bottom == typeTime:
		c := charts.Chart[time.Time, float64]{
			Width:   getValue("width", defaultWidth, options),
			Height:  getValue("height", defaultHeight, options),
			Title:   getValue("legend", "", options),
			Padding: createPadding(pad.Value),
		}
		var (
			ok     bool
			left   = getValue("left-axis", "", options)
			bottom = getValue("bottom-axis", "", options)
		)
		c.Left, ok = b.axis[left].(charts.Axis[float64])
		if !ok {
			return fmt.Errorf("%s can not be used as left axis - wrong type", left)
		}
		c.Bottom, ok = b.axis[bottom].(charts.Axis[time.Time])
		if !ok {
			return fmt.Errorf("%s can not be used as bottom axis - wrong type", bottom)
		}
		chart = c
	case left == typeNumber && bottom == typeNumber:
	}
	b.charts[ident] = chart
	return nil
}

func (b *Builder) guessChartType(options []option) (string, string) {
	idleft := getValue("left-axis", "", options)
	idbot := getValue("bottom-axis", "", options)

	var left, bottom string
	switch b.axis[idleft].(type) {
	case charts.Axis[time.Time]:
		left = typeTime
	case charts.Axis[float64]:
		left = typeNumber
	}

	switch b.axis[idbot].(type) {
	case charts.Axis[time.Time]:
		bottom = typeTime
	case charts.Axis[float64]:
		bottom = typeNumber
	}
	return left, bottom
}

func (b *Builder) executePie(ident string, options []option) error {
	// r := charts.PieRenderer{
	// 	Fill:        getValue("color", []string{}, options),
	// 	InnerRadius: getValue("inner-radius", 0, options),
	// 	OuterRadius: getValue("outer-radius", 0, options),
	// }
	// b.renderers = r
	return nil
}

func (b *Builder) executeLine(ident string, options []option) error {
	// r := charts.LinearRenderer{
	// 	Fill:          getValue("color", "black", options),
	// 	Skip:          int(getValue("skip", 0, options)),
	// 	IgnoreMissing: getValue("ignore-missing", false, options),
	// }
	// b.renderers = r
	return nil
}

func (b *Builder) executeCsv(ident string, options []option) error {
	return nil
}

type value interface {
	string | int | bool | time.Time | []string | []int | []time.Time
}

type command struct {
	Name    string
	Ident   string
	Options []option
}

type option struct {
	Name  string
	Value any
}

func findOption(name string, options []option) (option, error) {
	i := sort.Search(len(options), func(i int) bool {
		return options[i].Name <= name
	})
	var o option
	if i < len(options) && options[i].Name == name {
		return options[i], nil
	}
	return o, fmt.Errorf("%s: option not defined", name)
}

func getValue[T any](name string, value T, options []option) T {
	o, err := findOption(name, options)
	if err != nil {
		return value
	}
	ret, ok := o.Value.(T)
	if !ok {
		ret = value
	}
	return ret
}

func getType(options []option) string {
	return getValue("type", typeNumber, options)
}

func createPadding(value any) charts.Padding {
	var (
		pad    charts.Padding
		vs, ok = value.([]any)
	)
	if !ok {
		return pad
	}
	switch len(vs) {
	default:
	case 1:
		f, _ := vs[0].(float64)
		pad.Left = f
		pad.Right = f
		pad.Top = f
		pad.Bottom = f
	case 2:
		horiz, _ := vs[1].(float64)
		verti, _ := vs[2].(float64)
		pad.Top, pad.Bottom = verti, verti
		pad.Left, pad.Right = horiz, horiz
	case 3:
		pad.Top, _ = vs[0].(float64)
		pad.Right, _ = vs[1].(float64)
		pad.Bottom, _ = vs[2].(float64)
		pad.Left = pad.Right
	case 4:
		pad.Top, _ = vs[0].(float64)
		pad.Right, _ = vs[1].(float64)
		pad.Bottom, _ = vs[2].(float64)
		pad.Left, _ = vs[3].(float64)
	}
	return pad
}

func createRange(value any) charts.Range {
	var (
		vs, _  = value.([]any)
		fst, _ = slices.Fst(vs).(float64)
		lst, _ = slices.Lst(vs).(float64)
	)
	return charts.NewRange(fst, lst)
}

func createTimeDomain(value any) charts.Domain[time.Time] {
	var (
		vs, _  = value.([]any)
		fst, _ = slices.Fst(vs).(time.Time)
		lst, _ = slices.Lst(vs).(time.Time)
	)
	return charts.TimeDomain(fst, lst)
}

func createNumberDomain(value any) charts.Domain[float64] {
	var (
		vs, _  = value.([]any)
		fst, _ = slices.Fst(vs).(float64)
		lst, _ = slices.Lst(vs).(float64)
	)
	return charts.NumberDomain(fst, lst)
}
