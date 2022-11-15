package decode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/midbel/buddy/parse"
	"github.com/midbel/charts"
	"github.com/midbel/charts/dash"
	"github.com/midbel/slices"
)

var (
	DefaultShell     = "sh"
	DefaultShellArgs = "-c"
)

type Decoder struct {
	file  string
	path  string
	cwd   string
	shell string

	env   *dash.Environ[[]string]
	files *dash.Environ[dash.File]

	scan *Scanner
	curr Token
	peek Token
}

func NewDecoder(r io.Reader) *Decoder {
	d := Decoder{
		cwd:   ".",
		env:   dash.EmptyEnv[[]string](),
		files: dash.EmptyEnv[dash.File](),
		shell: DefaultShell,
		scan:  Scan(r),
	}
	if r, ok := r.(interface{ Name() string }); ok {
		d.file = r.Name()
		d.path = filepath.Dir(d.file)
	}
	if cwd, err := os.Getwd(); err == nil {
		d.cwd = cwd
	}
	d.next()
	d.next()
	return &d
}

func (d *Decoder) Decode() (*dash.Config, error) {
	cfg := dash.Default()
	return &cfg, d.decode(&cfg)
}

func (d *Decoder) decode(cfg *dash.Config) error {
	accept := func(tok Token) bool {
		return tok.Type == Keyword && tok.Literal != kwRender
	}
	err := d.decodeBody(cfg, accept)
	if err != nil {
		return err
	}
	if err := d.expectKw(kwRender); err != nil {
		return err
	}
	if err := d.decodeRender(cfg); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) decodeRender(cfg *dash.Config) error {
	d.next()
	switch d.curr.Type {
	case Literal:
		cfg.Path, _ = d.getString()
	case EOL, EOF, Comment:
	default:
		return d.decodeError("literal or end of line expected")
	}
	return d.eol()
}

func (d *Decoder) decodeBody(cfg *dash.Config, accept func(Token) bool) error {
	d.skipEOL()
	for accept(d.curr) && !d.done() {
		if err := d.expect(Keyword, "keyword expected"); err != nil {
			return err
		}
		var err error
		switch d.curr.Literal {
		case kwSet:
			err = d.decodeSet(cfg)
		case kwLoad:
			err = d.decodeLoad(cfg)
		case kwInclude:
			err = d.decodeInclude(cfg)
		case kwDefine:
			err = d.decodeDefine(cfg)
		case kwAt:
			err = d.decodeAt(cfg)
		case kwDeclare:
			err = d.decodeDeclare()
		case kwUse:
			err = d.decodeUse(cfg)
		default:
			err = d.decodeError(fmt.Sprintf("unexpected %q keyword", d.curr.Literal))
		}
		if err != nil {
			return err
		}
		d.skipEOL()
	}
	if accept(d.curr) {
		return d.decodeError("file can not be decoded")
	}
	return nil
}

func (d *Decoder) decodeUse(cfg *dash.Config) error {
	d.next()
	if err := d.expect(Literal, "literal expected"); err != nil {
		return err
	}
	fi, err := d.files.Resolve(d.curr.Literal)
	if err != nil {
		return fmt.Errorf("%s: file not registered", d.curr.Literal)
	}
	d.next()
	if err := d.expectKw(kwWith); err == nil {
		err = d.decodeStyle(&fi.Style)
		if err != nil {
			return err
		}
	}
	cfg.Inputs = append(cfg.Inputs, fi)
	return nil
}

func (d *Decoder) decodeAt(cfg *dash.Config) error {
	d.next()
	var (
		cell = dash.MakeCell(*cfg)
		err  error
	)
	if cell.Row, err = d.getInt(); err != nil {
		return err
	}
	if err := d.expect(Comma, "expected ','"); err != nil {
		return err
	}
	d.next()
	if cell.Col, err = d.getInt(); err != nil {
		return err
	}
	if d.is(Comma) {
		d.next()
		if cell.Width, err = d.getInt(); err != nil {
			return err
		}
		if err := d.expect(Comma, "expected ','"); err != nil {
			return err
		}
		d.next()
		if cell.Height, err = d.getInt(); err != nil {
			return err
		}
	}

	d.wrap()
	defer d.unwrap()
	switch {
	case d.isKw(kwInclude):
		err = d.decodeInclude(&cell.Config)
	case d.is(Lparen):
		d.next()
		accept := func(tok Token) bool {
			return tok.Type != Rparen
		}
		err = d.decodeBody(&cell.Config, accept)
		if err == nil {
			d.next()
			d.skipEOL()
		}
	default:
		err = d.decodeError("expected 'include' or '('")
	}
	if err == nil {
		cfg.Cells = append(cfg.Cells, cell)
	}
	return err
}

func (d *Decoder) decodeDefine(cfg *dash.Config) error {
	d.next()
	ident, err := d.getString()
	if err != nil {
		return err
	}
	if err := d.expect(Expr, "expected expression"); err != nil {
		return err
	}
	expr, err := parse.Parse(strings.NewReader(d.curr.Literal))
	if err != nil {
		return err
	}
	d.next()
	cfg.Scripts.Define(ident, expr)
	return d.eol()
}

func (d *Decoder) decodeDeclare() error {
	d.next()
	if err := d.expect(Literal, "literal expected"); err != nil {
		return err
	}
	ident := d.curr.Literal
	d.next()
	if d.is(Data) {
		return d.eol()
	}
	values, err := d.getStringList()
	if err != nil {
		return err
	}
	d.env.Define(ident, values)
	return d.eol()
}

func (d *Decoder) decodeInclude(cfg *dash.Config) error {
	accept := func(tok Token) bool {
		return tok.Type != EOF
	}
	decodeFile := func(file string) error {
		r, err := os.Open(file)
		if err != nil {
			return err
		}
		defer r.Close()

		err = NewDecoder(r).decodeBody(cfg, accept)
		if err != nil {
			return err
		}
		return err
	}
	d.next()
	list := []string{
		filepath.Join(d.path, d.curr.Literal),
		filepath.Join(d.cwd, d.curr.Literal),
	}
	d.next()
	var derr DecodeError
	for _, file := range list {
		err := decodeFile(file)
		if errors.As(err, &derr) {
			return err
		}
		if err == nil {
			break
		}
	}
	return d.eol()
}

func (d *Decoder) decodeSet(cfg *dash.Config) error {
	d.next()
	var (
		err error
		cmd = d.curr.Literal
	)
	d.next()
	switch cmd {
	case "title":
		cfg.Title, err = d.getString()
	case "size":
		list, err := d.getFloatList()
		if err != nil {
			return err
		}
		switch len(list) {
		case 1:
			cfg.Width, cfg.Height = list[0], list[0]
		case 2:
			cfg.Width, cfg.Height = list[0], list[1]
		default:
			err = fmt.Errorf("invalid number of values given for chart size")
		}
	case "grid":
		list, err := d.getIntList()
		if err != nil {
			return err
		}
		switch len(list) {
		case 1:
			cfg.Rows, cfg.Cols = list[0], list[0]
		case 2:
			cfg.Rows, cfg.Cols = list[0], list[1]
		default:
			err = fmt.Errorf("invalid number of values given for grid dimension")
		}
	case "padding":
		list, err := d.getFloatList()
		if err != nil {
			return err
		}
		cfg.Pad, err = charts.PaddingFromList(list)
	case "xdata":
		cfg.X.Type, err = d.getType()
	case "xdomain":
		cfg.X.Scaler, err = d.decodeScaler()
	case "ydata":
		cfg.Y.Type, err = d.getType()
	case "ydomain":
		cfg.Y.Scaler, err = d.decodeScaler()
	case "xticks":
		return d.decodeTicks(&cfg.X.Domain)
	case "yticks":
		return d.decodeTicks(&cfg.Y.Domain)
	case "style":
		return d.decodeStyle(&cfg.Style)
	case "timefmt":
		cfg.TimeFormat, err = d.getString()
	case "delimiter":
		cfg.Delimiter, err = d.getString()
	case "legend":
		return d.decodeLegend(cfg)
	case "theme":
		cfg.Theme, err = d.getString()
	default:
		err = d.optionError("set")
	}
	if err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeScaler() (dash.ScalerMaker, error) {
	if !d.peekIs(Keyword) {
		list, err := d.getStringList()
		if err != nil {
			return nil, err
		}
		return dash.ScaleFromList(list), nil
	}
	path, err := d.getString()
	if err != nil {
		return nil, err
	}
	if err := d.expectKw(kwUsing); err != nil {
		return nil, err
	}
	d.next()
	var idx dash.Indexer
	switch d.peek.Type {
	case Sum:
		var list []int
		ix, err := d.getInt()
		if err != nil {
			return nil, err
		}
		list = append(list, ix)
		for d.curr.Type == Sum {
			d.next()
			ix, err := d.getInt()
			if err != nil {
				return nil, err
			}
			list = append(list, ix)
		}
		idx = dash.SelectSum(list)
	case RangeSum:
		fst, err := d.getInt()
		if err != nil {
			return nil, err
		}
		d.next()
		lst, err := d.getInt()
		if err != nil {
			return nil, err
		}
		idx = dash.SelectSum(dash.ExpandRange(fst, lst))
	case EOL, EOF:
		x, err := d.getInt()
		if err != nil {
			return nil, err
		}
		idx = dash.SelectSingle(x)
	default:
		return nil, d.decodeError("expected ':', ':+' or end of line")
	}
	return dash.ScaleFromFile(path, idx), nil
}

func (d *Decoder) decodeLegend(cfg *dash.Config) error {
	var (
		cmd = d.curr.Literal
		err error
	)
	d.next()
	switch cmd {
	case "title":
		cfg.Legend.Title, err = d.getString()
	case "position":
		cfg.Legend.Position, err = d.getStringList()
		if len(cfg.Legend.Position) > 2 && err == nil {
			err = fmt.Errorf("too many values given for legend position")
		}
	case kwWith:
		err = d.decodeWith(func() error {
			return d.decodeLegend(cfg)
		})
	default:
		err = d.optionError("legend")
	}
	if err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeStyle(style *dash.Style) error {
	var (
		cmd = d.curr.Literal
		err error
	)
	d.next()
	switch cmd {
	case "type":
		style.Type, err = d.getRenderType()
	case "color":
		style.Stroke, err = d.getString()
	case "fill":
		style.Fill, err = d.getBool()
	case "ignore-missing":
		style.IgnoreMissing, err = d.getBool()
	case "text-position":
		style.TextPosition, err = d.getString()
	case "inner-radius":
		style.InnerRadius, err = d.getFloat()
	case "outer-radius":
		style.OuterRadius, err = d.getFloat()
	case "line-style":
		style.LineStyle, err = d.getLineStyle()
	case "width":
		style.Width, err = d.getFloat()
	case kwWith:
		err = d.decodeWith(func() error {
			return d.decodeStyle(style)
		})
	default:
		err = d.optionError("style")
	}
	if err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeTicks(dom *dash.Domain) error {
	if d.peekIs(EOL) || d.peekIs(EOF) {
		count, err := d.getInt()
		if err != nil {
			return err
		}
		dom.Ticks = count
		return d.eol()
	}
	var (
		cmd = d.curr.Literal
		err error
	)
	d.next()
	switch cmd {
	case "count":
		dom.Ticks, err = d.getInt()
	case "position":
		dom.Position, err = d.getString()
	case "label":
		dom.Label, err = d.getString()
	case "format":
		dom.Format, err = d.getString()
	case "inner-ticks":
		dom.InnerTicks, err = d.getBool()
	case "outer-ticks":
		dom.OuterTicks, err = d.getBool()
	case "label-ticks":
		dom.LabelTicks, err = d.getBool()
	case "band-ticks":
		dom.BandTicks, err = d.getBool()
	case kwWith:
		err = d.decodeWith(func() error {
			return d.decodeTicks(dom)
		})
	default:
		err = d.optionError("ticks")
	}
	if err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeLoad(cfg *dash.Config) error {
	d.next()
	var (
		fi  dash.File
		err error
	)
	if fi.Path, err = d.getString(); err != nil {
		return err
	}
	if err := d.expectKw(kwLimit); err == nil {
		d.next()
		if d.peekIs(Comma) {
			fi.Offset, err = d.getInt()
			if err != nil {
				return err
			}
		}
		fi.Count, err = d.getInt()
		if err != nil {
			return err
		}
	}
	if err := d.expectKw(kwUsing); err != nil {
		return err
	}
	d.next()
	if d.peekIs(Comma) {
		if fi.X, err = d.getInt(); err != nil {
			return err
		}
		d.next()
	}
	if fi.Y, err = d.decodeSelect(); err != nil {
		return err
	}
	if err := d.expectKw(kwWith); err == nil {
		err = d.decodeStyle(&fi.Style)
	} else {
		err = d.eol()
	}
	if err == nil {
		d.files.Define(fi.Name(), fi)
		cfg.Inputs = append(cfg.Inputs, fi)
	}
	return err
}

func (d *Decoder) decodeSelect() (dash.Selector, error) {
	getRange := func() ([]int, error) {
		fst, err := d.getInt()
		if err != nil {
			return nil, err
		}
		d.next()
		lst, err := d.getInt()
		if err != nil {
			return nil, err
		}
		return dash.ExpandRange(fst, lst), nil
	}
	getList := func(want rune) ([]int, error) {
		var list []int
		i, err := d.getInt()
		if err != nil {
			return nil, err
		}
		list = append(list, i)
		for d.curr.Type == want {
			d.next()
			i, err := d.getInt()
			if err != nil {
				return nil, err
			}
			list = append(list, i)
		}
		return list, nil
	}
	var xs []dash.Selector
	for !d.is(EOL) && !d.is(EOF) && !d.is(Keyword) {
		switch d.peek.Type {
		case Comma:
			rg, err := getList(Comma)
			if err != nil {
				return nil, err
			}
			xs = append(xs, dash.SelectMulti(rg))
		case Sum:
			rg, err := getList(Sum)
			if err != nil {
				return nil, err
			}
			xs = append(xs, dash.SelectSum(rg))
		case Range:
			rg, err := getRange()
			if err != nil {
				return nil, err
			}
			xs = append(xs, dash.SelectMulti(rg))
		case RangeSum:
			rg, err := getRange()
			if err != nil {
				return nil, err
			}
			xs = append(xs, dash.SelectSum(rg))
		case Keyword, EOL, EOF:
			i, err := d.getInt()
			if err != nil {
				return nil, err
			}
			xs = append(xs, dash.SelectMulti([]int{i}))
		default:
			return nil, d.decodeError("expected ',', ':', ':+', keyword or end of line")
		}
		switch d.curr.Type {
		case Comma:
			d.next()
		case EOL, EOF, Keyword:
		default:
			return nil, d.decodeError("expected ',', keyword or end of line")
		}
	}
	if len(xs) == 1 {
		return slices.Fst(xs), nil
	}
	return dash.Combined(xs...), nil
}

func (d *Decoder) decodeWith(decode func() error) error {
	if err := d.expect(Lparen, "expected '('"); err != nil {
		return err
	}
	d.next()
	d.skipEOL()
	for !d.is(Rparen) && !d.done() {
		if err := decode(); err != nil {
			return err
		}
	}
	if err := d.expect(Rparen, "expected ')'"); err != nil {
		return err
	}
	d.next()
	return nil
}

func (d *Decoder) wrap() {
	d.env = d.env.Wrap()
	d.files = d.files.Wrap()
}

func (d *Decoder) unwrap() {
	d.env = d.env.Unwrap()
	d.files = d.files.Unwrap()
}

func (d *Decoder) is(kind rune) bool {
	return d.curr.Type == kind
}

func (d *Decoder) peekIs(kind rune) bool {
	return d.peek.Type == kind
}

func (d *Decoder) isKw(kw string) bool {
	return d.is(Keyword) && d.curr.Literal == kw
}

func (d *Decoder) expectKw(kw string) error {
	if err := d.expect(Keyword, fmt.Sprintf("expected %q keyword", kw)); err != nil {
		return err
	}
	if d.curr.Literal != kw {
		return d.decodeError(fmt.Sprintf("%q expected, got %s", kw, d.curr.Literal))
	}
	return nil
}

func (d *Decoder) expect(kind rune, msg string) error {
	if d.is(kind) {
		return nil
	}
	return d.decodeError(msg)
}

func (d *Decoder) next() {
	d.curr = d.peek
	d.peek = d.scan.Scan()
}

func (d *Decoder) done() bool {
	return d.curr.Type == EOF
}

func (d *Decoder) eol() error {
	if !d.is(EOL) && !d.is(EOF) && !d.is(Comment) {
		return d.decodeError("expected end of line or end of file")
	}
	d.next()
	return nil
}

func (d *Decoder) optionError(item string) error {
	return OptionError{
		Position: d.curr.Position,
		File:     d.file,
		Option:   d.curr.Literal,
		Section:  item,
	}
}

func (d *Decoder) decodeError(msg string) error {
	return DecodeError{
		Position: d.curr.Position,
		File:     d.file,
		Message:  msg,
	}
}

func (d *Decoder) skipEOL() {
	for d.is(EOL) || d.is(Comment) {
		d.next()
	}
}

func (d *Decoder) getLineStyle() (string, error) {
	str, err := d.getString()
	if err != nil {
		return str, err
	}
	switch str {
	case dash.StyleStraight, dash.StyleDotted, dash.StyleDashed:
		return str, nil
	default:
		return "", fmt.Errorf("%s: unknown line style provided", str)
	}
}

func (d *Decoder) getRenderType() (string, error) {
	str, err := d.getString()
	if err != nil {
		return str, err
	}
	switch str {
	case dash.RenderLine, dash.RenderStep, dash.RenderStepAfter, dash.RenderStepBefore:
	case dash.RenderBar, dash.RenderPie, dash.RenderStack, dash.RenderNormStack, dash.RenderGroup:
	default:
		return "", fmt.Errorf("%s: unknown type provided", str)
	}
	return str, nil
}

func (d *Decoder) getType() (string, error) {
	str, err := d.getString()
	if err != nil {
		return str, err
	}
	switch str {
	case dash.TypeNumber, dash.TypeTime, dash.TypeString:
		return str, nil
	default:
		return "", fmt.Errorf("%s: unknown type provided", str)
	}
}

func (d *Decoder) getString() (string, error) {
	var str string
	switch d.curr.Type {
	case Literal:
		str = d.curr.Literal
	case Variable:
		vs, err := d.env.Resolve(d.curr.Literal)
		if err != nil {
			return "", err
		}
		str = slices.Fst(vs)
	case Command:
		var (
			out bytes.Buffer
			err bytes.Buffer
		)
		cmd := exec.Command(d.shell, "-c", d.curr.Literal)
		cmd.Stdout = &out
		cmd.Stderr = &err
		if errc := cmd.Run(); errc != nil {
			return "", fmt.Errorf("%w: %s", errc, err.String())
		}
		str = strings.TrimSpace(out.String())
	default:
		return "", d.decodeError("expected literal, variable or command")
	}
	defer d.next()
	return str, nil
}

func (d *Decoder) getBool() (bool, error) {
	str, err := d.getString()
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(str)
}

func (d *Decoder) getInt() (int, error) {
	str, err := d.getString()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func (d *Decoder) getFloat() (float64, error) {
	str, err := d.getString()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(str, 64)
}

func (d *Decoder) getStringList() ([]string, error) {
	var list []string
	for !d.is(EOL) && !d.is(EOF) {
		str, err := d.getString()
		if err != nil {
			return nil, err
		}
		list = append(list, str)
		if err := d.nextListItem(); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (d *Decoder) getIntList() ([]int, error) {
	var list []int
	for !d.is(EOL) && !d.is(EOF) {
		i, err := d.getInt()
		if err != nil {
			return nil, err
		}
		list = append(list, i)
		if err := d.nextListItem(); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (d *Decoder) getFloatList() ([]float64, error) {
	var list []float64
	for !d.is(EOL) && !d.is(EOF) {
		f, err := d.getFloat()
		if err != nil {
			return nil, err
		}
		list = append(list, f)
		if err := d.nextListItem(); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (d *Decoder) nextListItem() error {
	switch d.curr.Type {
	case Comma:
		if d.peekIs(EOL) || d.peekIs(EOF) {
			return d.decodeError("end of line not expected after ',")
		}
		d.next()
	case EOF, EOL:
	default:
		return d.decodeError("expected ',' or end of line")
	}
	return nil
}
