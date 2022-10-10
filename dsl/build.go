package dsl

import (
	"fmt"
	"io"
	"strconv"
)

type Builder struct {
	scan *Scanner
	curr Token
	peek Token

	builders map[string]func() error
}

func New(r io.Reader) *Builder {
	var b Builder
	b.builders = map[string]func() error {
		"scale": b.buildScale,
		"serie": b.buildSerie,
		"axis": b.buildAxis,
		"chart": b.buildChart,
		"renderer": b.buildRenderer,
		"render": nil,
		"csv": nil,
		"json": nil,
		"xml": nil,
	}
	b.scan = Scan(r)
	b.next()
	b.next()
	return &b
}

func (b *Builder) Build() error {
	for !b.done() {
		if b.curr.Type == Comment {
			b.next()
			continue
		}
		build, ok := b.builders[b.curr.Literal]
		if !ok {
			return fmt.Errorf("%s: unsupported command", b.curr.Literal)
		}
		if err := build(); err != nil {
			return err
		}
	}
	b.next()
	return nil
}

func (b *Builder) buildAxis() error {
	b.next()
	b.next()
	var (
		it  = defaultAxis(b.curr.Literal)
		err error
	)
	b.next()
	b.next()
	for b.curr.Type != EOL && b.curr.Type != EOF {
		if b.curr.Type == Comment {
			b.next()
			continue
		}
		switch b.curr.Literal {
		case "type":
			it.Type, err = b.getString()
			if err == nil {
				err = validType(it.Type)
			}
		case "legend":
			it.Title, err = b.getString()
		case "ticks":
			it.Ticks, err = b.getInt()
		case "scale":
			_, err = b.getReference()
		case "label-ticks":
			it.Label, err = b.getBool()
		case "outer-ticks":
			it.Outer, err = b.getBool()
		case "inner-ticks":
			it.Inner, err = b.getBool()
		case "bands-ticks":
			it.Bands, err = b.getBool()
		case "color":
			it.Color, err = b.getString()
		default:
			err = unknownProp(b.curr.Literal, "axis")
		}
		if err := hasError(err, b.eol()); err != nil {
			return err
		}
	}
	b.next()
	return nil
}

func (b *Builder) buildScale() error {
	b.next()
	b.next()
	var (
		it  = defaultScale(b.curr.Literal)
		err error
	)
	b.next()
	b.next()
	for b.curr.Type != EOL && b.curr.Type != EOF {
		if b.curr.Type == Comment {
			b.next()
			continue
		}
		switch b.curr.Literal {
		case "type":
			it.Type, err = b.getString()
			if err == nil {
				err = validType(it.Type)
			}
		case "range":
			it.Range, err = b.getRange()
			if n := len(it.Range); n != 2 {
				err = fmt.Errorf("invalid number of values given for scale range: %d", n)
			}
		case "domain":
			it.Domain, err = b.getList()
			if n := len(it.Domain); n < 2 {
				err = fmt.Errorf("invalid number of values given for scale domain: %d", n)
			}
		default:
			err = unknownProp(b.curr.Literal, "scale")
		}
		if err := hasError(err, b.eol()); err != nil {
			return err
		}
	}
	b.next()
	return nil
}

func (b *Builder) buildSerie() error {
	b.next()
	b.next()
	var (
		it  = defaultSerie(b.curr.Literal)
		err error
	)
	_ = it
	b.next()
	b.next()
	for b.curr.Type != EOL && b.curr.Type != EOF {
		switch b.curr.Literal {
		case "x":
			_, err = b.getReference()
		case "y":
			_, err = b.getReference()
		case "renderer":
			_, err = b.getReference()
		case "values":
			_, err = b.getReference()
		default:
			err = unknownProp(b.curr.Literal, "serie")
		}
		if err := hasError(err, b.eol()); err != nil {
			return err
		}
	}
	b.next()
	return nil
}

func (b *Builder) buildRenderer() error {
	b.next()
	b.next()
	var err error
	switch b.curr.Literal {
	default:
		err = fmt.Errorf("%s: unknown renderer type", b.curr.Literal)
	case "line":
		err = b.buildLineRenderer()
	case "step":
	case "step-after":
	case "step-before":
	case "bar":
	case "pie":
		err = b.buildPieRenderer()
	}
	return err
}

func (b *Builder) buildPieRenderer() error {
	b.next()
	b.next()
	var (
		it  = defaultPieRenderer(b.curr.Literal)
		err error
	)
	b.next()
	b.next()
	for b.curr.Type != EOL && b.curr.Type != EOF {
		switch b.curr.Literal {
		case "inner-radius":
			it.Inner, err = b.getFloat()
		case "outer-radius":
			it.Outer, err = b.getFloat()
		case "colors":
			it.Colors, err = b.getList()
		default:
			err = unknownProp(b.curr.Literal, "pie")
		}
		if err := hasError(err, b.eol()); err != nil {
			return err
		}
	}
	b.next()
	return nil
}

func (b *Builder) buildLineRenderer() error {
	b.next()
	b.next()
	var (
		it  = defaultLineRenderer(b.curr.Literal)
		err error
	)
	b.next()
	b.next()
	for b.curr.Type != EOL && b.curr.Type != EOF {
		switch b.curr.Literal {
		case "ignore-missing":
			it.IgnoreMissing, err = b.getBool()
		case "color":
			it.Color, err = b.getString()
		case "point":
			it.Point, err = b.getString()
		default:
			err = unknownProp(b.curr.Literal, "line")
		}
		if err := hasError(err, b.eol()); err != nil {
			return err
		}
	}
	b.next()
	return nil
}

func (b *Builder) buildChart() error {
	b.next()
	b.next()
	var (
		it  = defaultChart(b.curr.Literal)
		err error
	)
	b.next()
	b.next()
	for b.curr.Type != EOL && b.curr.Type != EOF {
		switch b.curr.Literal {
		case "width":
			it.Width, err = b.getFloat()
		case "height":
			it.Height, err = b.getFloat()
		case "title":
			it.Title, err = b.getString()
		case "left-axis":
			_, err = b.getReference()
		case "right-axis":
			_, err = b.getReference()
		case "top-axis":
			_, err = b.getReference()
		case "bottom-axis":
			_, err = b.getReference()
		case "legend":
			_, err = b.getReference()
		case "center":
		default:
			err = unknownProp(b.curr.Literal, "chart")
		}
		if err := hasError(err, b.eol()); err != nil {
			return err
		}
	}
	b.next()
	return nil
}

func (b *Builder) getReference() (string, error) {
	b.next()
	if b.curr.Type != Equal {
		return "", fmt.Errorf("syntax error! missing equal after name")
	}
	b.next()
	switch b.curr.Type {
	case Reference:
		defer b.next()
		return b.curr.Literal, nil
	case Command:
		defer b.next()
		return b.curr.Literal, nil
	default:
		return "", unexpectedToken(b.curr)
	}
}

func (b *Builder) getString() (string, error) {
	b.next()
	if b.curr.Type != Equal {
		return "", fmt.Errorf("syntax error! missing equal after name")
	}
	b.next()
	if b.curr.Type != Literal {
		return "", fmt.Errorf("syntax error! literal expected, got %s", b.curr)
	}
	defer b.next()
	return b.curr.Literal, nil
}

func (b *Builder) getBool() (bool, error) {
	b.next()
	if b.curr.Type != Equal {
		return false, fmt.Errorf("syntax error! missing equal after name")
	}
	b.next()
	if b.curr.Type != Boolean {
		return false, fmt.Errorf("syntax error! boolean expected, got %s", b.curr)
	}
	defer b.next()
	switch b.curr.Literal {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return strconv.ParseBool(b.curr.Literal)
	}
}

func (b *Builder) getInt() (int, error) {
	i, err := b.getFloat()
	return int(i), err
}

func (b *Builder) getFloat() (float64, error) {
	b.next()
	if b.curr.Type != Equal {
		return 0, fmt.Errorf("syntax error! missing equal after name")
	}
	b.next()
	if b.curr.Type != Number {
		return 0, fmt.Errorf("syntax error! expected number, got %s", b.curr)
	}
	defer b.next()
	return strconv.ParseFloat(b.curr.Literal, 64)
}

func (b *Builder) getRange() ([]float64, error) {
	b.next()
	if b.curr.Type != Equal {
		return nil, fmt.Errorf("syntax error! missing equal after name")
	}
	b.next()
	var list []float64
	for b.curr.Type != Blank && b.curr.Type != EOL && b.curr.Type != EOF {
		if b.curr.Type != Number {
			return nil, fmt.Errorf("syntax error! expected number, got %s", b.curr)
		}
		f, err := strconv.ParseFloat(b.curr.Literal, 64)
		if err != nil {
			return nil, err
		}
		list = append(list, f)
		b.next()
		switch b.curr.Type {
		case Comma:
			b.next()
		case Blank, EOL, EOF:
		default:
			return nil, unexpectedToken(b.curr)
		}
	}
	return list, nil
}

func (b *Builder) getList() ([]string, error) {
	b.next()
	if b.curr.Type != Equal {
		return nil, fmt.Errorf("syntax error! missing equal after name")
	}
	b.next()
	var list []string
	for b.curr.Type != Blank && b.curr.Type != EOL && b.curr.Type != EOF {
		list = append(list, b.curr.Literal)
		b.next()
		switch b.curr.Type {
		case Comma:
			b.next()
		case Blank, EOL, EOF:
		default:
			return nil, unexpectedToken(b.curr)
		}
	}
	return list, nil
}

func (b *Builder) done() bool {
	return b.curr.Type == EOF
}

func (b *Builder) eol() error {
	switch b.curr.Type {
	case Blank, Comment:
		b.next()
	case EOL, EOF:
	default:
		return unexpectedToken(b.curr)
	}
	return nil
}

func (b *Builder) next() {
	b.curr = b.peek
	b.peek = b.scan.Scan()
}

func hasError(err ...error) error {
	for _, e := range err {
		if e != nil {
			return e
		}
	}
	return nil
}

func unexpectedToken(tok Token) error {
	return fmt.Errorf("syntax error! unexpected token %s", tok)
}

func unknownProp(prop, typ string) error {
	return fmt.Errorf("%s: unknown property for %s node", prop, typ)
}
