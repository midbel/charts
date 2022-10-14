package dsl

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	DefaultShell     = "sh"
	DefaultShellArgs = "-c"
)

type Decoder struct {
	path string

	shell string

	scan *Scanner
	curr Token
	peek Token
}

func NewDecoder(r io.Reader) *Decoder {
	var d Decoder
	if r, ok := r.(interface{ Name() string }); ok {
		d.path = filepath.Dir(r.Name())
	}
	d.shell = DefaultShell
	d.scan = Scan(r)
	d.next()
	d.next()
	return &d
}

func (d *Decoder) Decode() error {
	cfg := Default()
	return d.decode(&cfg)
}

func (d *Decoder) decode(cfg *Config) error {
	for !d.done() {
		if d.curr.Type == Comment {
			d.next()
			continue
		}
		if d.curr.Type != Keyword {
			return fmt.Errorf("expected keyword but got %q", d.curr.Literal)
		}
		if d.curr.Literal == kwRender {
			break
		}
		var err error
		switch d.curr.Literal {
		case kwSet:
			err = d.decodeSet(cfg)
		case kwLoad:
			err = d.decodeLoad(cfg)
		case kwInclude:
			err = d.decodeInclude(cfg)
		default:
			err = fmt.Errorf("unexpected keyword %s", d.curr.Literal)
		}
		if err != nil {
			return err
		}
	}
	if d.curr.Type != Keyword && d.curr.Literal != kwRender {
		return fmt.Errorf("expected keyword but got %q", d.curr.Literal)
	}
	if err := d.decodeRender(cfg); err != nil {
		return err
	}
	if d.curr.Type != EOF {
		return fmt.Errorf("unexpected token %s", d.curr)
	}
	return cfg.Render()
}

func (d *Decoder) decodeRender(cfg *Config) error {
	d.next()
	switch d.curr.Type {
	case Literal:
		cfg.Path, _ = d.getString()
	case EOL, EOF:
	default:
		return fmt.Errorf("unexpected token %s", d.curr)
	}
	return d.eol()
}

func (d *Decoder) decodeInclude(cfg *Config) error {
	d.next()
	r, err := os.Open(d.curr.Literal)
	if err != nil {
		return err
	}
	defer r.Close()
	if err := NewDecoder(r).decode(cfg); err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeSet(cfg *Config) error {
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
			err = fmt.Errorf("invalid number values given for size")
		}
	case "padding":
		list, err := d.getFloatList()
		if err != nil {
			return err
		}
		switch len(list) {
		case 1:
			cfg.Pad.Top = list[0]
			cfg.Pad.Right = list[0]
			cfg.Pad.Bottom = list[0]
			cfg.Pad.Left = list[0]
		case 2:
			cfg.Pad.Top, cfg.Pad.Bottom = list[0], list[0]
			cfg.Pad.Right, cfg.Pad.Left = list[1], list[1]
		case 3:
			cfg.Pad.Top = list[0]
			cfg.Pad.Bottom = list[2]
			cfg.Pad.Right, cfg.Pad.Left = list[1], list[1]
		case 4:
			cfg.Pad.Top = list[0]
			cfg.Pad.Right = list[1]
			cfg.Pad.Bottom = list[2]
			cfg.Pad.Left = list[3]
		default:
			err = fmt.Errorf("invalid number values given for padding")
		}
	case "xdata":
		cfg.Types.X, err = d.getType()
	case "xcenter":
		cfg.Center.X, err = d.getString()
	case "xdomain":
		cfg.Domains.X.Domain, err = d.getStringList()
	case "ydata":
		cfg.Types.Y, err = d.getType()
	case "ycenter":
		cfg.Center.Y, err = d.getString()
	case "ydomain":
		cfg.Domains.Y.Domain, err = d.getStringList()
	case "xticks":
		return d.decodeTicks(&cfg.Domains.X)
	case "yticks":
		return d.decodeTicks(&cfg.Domains.Y)
	case "style":
		return d.decodeStyle(&cfg.Style)
	case "timefmt":
		cfg.TimeFormat, err = d.getString()
	case "delimiter":
		cfg.Delimiter, err = d.getString()
	case "legend":
		return d.decodeLegend(cfg)
	default:
		err = fmt.Errorf("%s unsupported/unknown option", cmd)
	}
	if err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeLegend(cfg *Config) error {
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
		err = fmt.Errorf("%s unsupported/unknown option for legend", d.curr.Literal)
	}	
	if err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeStyle(style *Style) error {
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
	case kwWith:
		err = d.decodeWith(func() error {
			return d.decodeStyle(style)
		})
	default:
		err = fmt.Errorf("%s unsupported/unknown option for style", d.curr.Literal)
	}
	if err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeTicks(dom *Domain) error {
	if d.peek.Type == EOL || d.peek.Type == EOF {
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
		err = fmt.Errorf("%s unsupported/unknown option for ticks", cmd)
	}
	if err != nil {
		return err
	}
	return d.eol()
}

func (d *Decoder) decodeLoad(cfg *Config) error {
	d.next()
	var (
		fi  File
		err error
	)
	if fi.Path, err = d.getString(); err != nil {
		return err
	}
	if d.curr.Type != Keyword && d.curr.Literal != kwUsing {
		return fmt.Errorf("unexpected token %s", d.curr)
	}
	d.next()
	if d.peek.Type == Comma {
		if fi.X, err = d.getInt(); err != nil {
			return err
		}
		d.next()
	}
	if fi.Y, err = d.getInt(); err != nil {
		return err
	}
	if d.curr.Type == Keyword && d.curr.Literal == kwWith {
		err = d.decodeStyle(&fi.Style)
	} else {
		err = d.eol()
	}
	if err == nil {
		cfg.Files = append(cfg.Files, fi)
	}
	return err
}

func (d *Decoder) decodeWith(decode func() error) error {
	if d.curr.Type != Lparen {
		return fmt.Errorf("unexpected token %s", d.curr)
	}
	d.next()
	d.skipEOL()
	for d.curr.Type != Rparen && !d.done() {
		if err := decode(); err != nil {
			return err
		}
	}
	if d.curr.Type != Rparen {
		return fmt.Errorf("unexpected token %s", d.curr)
	}
	d.next()
	return nil
}

func (d *Decoder) next() {
	d.curr = d.peek
	d.peek = d.scan.Scan()
}

func (d *Decoder) done() bool {
	return d.curr.Type == EOF
}

func (d *Decoder) eol() error {
	if d.curr.Type != EOL && d.curr.Type != EOF {
		return fmt.Errorf("expected end of line, got %s", d.curr)
	}
	d.next()
	return nil
}

func (d *Decoder) skipEOL() {
	for d.curr.Type == EOL {
		d.next()
	}
}

func (d *Decoder) getRenderType() (string, error) {
	str, err := d.getString()
	if err != nil {
		return str, err
	}
	switch str {
	case RenderLine, RenderStep, RenderStepAfter, RenderStepBefore:
		return str, nil
	default:
		return "", fmt.Errorf("%s: unknown type provided", str)
	}
}

func (d *Decoder) getType() (string, error) {
	str, err := d.getString()
	if err != nil {
		return str, err
	}
	switch str {
	case TypeNumber, TypeTime, TypeString:
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
		return "", fmt.Errorf("expected literal, got %s", d.curr)
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
	for d.curr.Type != EOL && d.curr.Type != EOF {
		str, err := d.getString()
		if err != nil {
			return nil, err
		}
		list = append(list, str)
		switch d.curr.Type {
		case Comma:
			if d.peek.Type == EOL || d.peek.Type == EOF {
				return nil, fmt.Errorf("unexpected token %s", d.curr)
			}
			d.next()
		case EOF, EOL:
		default:
			return nil, fmt.Errorf("unexpected token %s", d.curr)
		}
	}
	return list, nil
}

func (d *Decoder) getFloatList() ([]float64, error) {
	var list []float64
	for d.curr.Type != EOL && d.curr.Type != EOF {
		f, err := d.getFloat()
		if err != nil {
			return nil, err
		}
		list = append(list, f)
		switch d.curr.Type {
		case Comma:
			if d.peek.Type == EOL || d.peek.Type == EOF {
				return nil, fmt.Errorf("unexpected token %s", d.curr)
			}
			d.next()
		case EOF, EOL:
		default:
			return nil, fmt.Errorf("unexpected token %s", d.curr)
		}
	}
	return list, nil
}
