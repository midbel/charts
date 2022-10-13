package dsl

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

type Decoder struct {
	path string
	scan *Scanner
	curr Token
	peek Token
}

func NewDecoder(r io.Reader) *Decoder {
	var d Decoder
	if r, ok := r.(interface{ Name() string }); ok {
		d.path = r.Name()
	}
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
		fmt.Println(d.curr, d.peek)
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
	fmt.Printf("%+v\n", cfg)
	return cfg.Render()
}

func (d *Decoder) decodeInclude(cfg *Config) error {
	d.next()
	r, err := os.Open(d.curr.Literal)
	if err != nil {
		return err
	}
	defer r.Close()
	return NewDecoder(r).decode(cfg)
}

func (d *Decoder) decodeSet(cfg *Config) error {
	fmt.Println("enter decodeSet")
	defer fmt.Println(">> leave decodeSet")
	d.next()
	var err error
	switch d.curr.Literal {
	case "title":
		d.next()
		cfg.Title = d.curr.Literal
		d.next()
	case "size":
		d.next()
		cfg.Width, err = strconv.ParseFloat(d.curr.Literal, 64)
		if err != nil {
			break
		}
		d.next()
		if d.curr.Type != Comma {
			err = fmt.Errorf("unexpected token %s", d.curr)
		}
		d.next()
		cfg.Height, err = strconv.ParseFloat(d.curr.Literal, 64)
		if err != nil {
			return err
		}
		d.next()
	case "padding":
		d.next()
		var list []float64
		for d.curr.Type != EOL && d.curr.Type != EOF {
			f, err := strconv.ParseFloat(d.curr.Literal, 64)
			if err != nil {
				return err
			}
			list = append(list, f)
			d.next()
			switch d.curr.Type {
			case Comma:
				d.next()
			case EOF, EOL:
			default:
				return fmt.Errorf("unexpected token %s", d.curr)
			}
		}
		switch len(list) {
		case 0:
		case 1:
			cfg.Pad.Top = list[0]
			cfg.Pad.Right = list[0]
			cfg.Pad.Bottom = list[0]
			cfg.Pad.Left = list[0]
		case 2:
			cfg.Pad.Top = list[0]
			cfg.Pad.Right = list[1]
			cfg.Pad.Bottom = list[0]
			cfg.Pad.Left = list[1]
		case 3:
			cfg.Pad.Top = list[0]
			cfg.Pad.Right = list[1]
			cfg.Pad.Bottom = list[2]
			cfg.Pad.Left = list[1]
		case 4:
			cfg.Pad.Top = list[0]
			cfg.Pad.Right = list[1]
			cfg.Pad.Bottom = list[2]
			cfg.Pad.Left = list[3]
		default:
			err = fmt.Errorf("too many values given for padding")
		}
	case "xdata":
		d.next()
		cfg.Types.X = d.curr.Literal
		d.next()
	case "xcenter":
		d.next()
		cfg.Center.X = d.curr.Literal
		d.next()
	case "xdomain":
		d.next()
		for d.curr.Type != EOL && d.curr.Type != EOF {
			cfg.Domains.X.Domain = append(cfg.Domains.X.Domain, d.curr.Literal)
			d.next()
			switch d.curr.Type {
			case Comma:
				d.next()
			case EOF, EOL:
			default:
				return fmt.Errorf("unexpected token %s", d.curr)
			}
		}
	case "ydata":
		d.next()
		cfg.Types.Y = d.curr.Literal
		d.next()
	case "ycenter":
		d.next()
		cfg.Center.Y = d.curr.Literal
		d.next()
	case "ydomain":
		d.next()
		for d.curr.Type != EOL && d.curr.Type != EOF {
			cfg.Domains.Y.Domain = append(cfg.Domains.Y.Domain, d.curr.Literal)
			d.next()
			switch d.curr.Type {
			case Comma:
				d.next()
			case EOF, EOL:
			default:
				return fmt.Errorf("unexpected token %s", d.curr)
			}
		}
	case "xticks":
		return d.decodeTicks(&cfg.Domains.X)
	case "yticks":
		return d.decodeTicks(&cfg.Domains.Y)
	case "timefmt":
		d.next()
		cfg.TimeFormat = d.curr.Literal
		d.next()
	default:
		err = fmt.Errorf("%s unsupported/unknown option", d.curr.Literal)
	}
	if d.curr.Type != EOL && d.curr.Type != EOF {
		return fmt.Errorf("expected end of line, got %s", d.curr)
	}
	d.next()
	return err
}

func (d *Decoder) decodeTicks(dom *Domain) error {
	d.next()
	if d.peek.Type == EOL || d.peek.Type == EOF {
		count, err := strconv.Atoi(d.curr.Literal)
		if err != nil {
			return err
		}
		dom.Ticks = count
		d.next()
		if d.curr.Type != EOL && d.curr.Type != EOF {
			return fmt.Errorf("unexpected token %s", d.curr)
		}
		d.next()
		return nil
	}
	if d.peek.Type == Keyword && d.peek.Literal == kwWith {
		d.next()
		d.next()
		if d.curr.Type != Lparen {
			return fmt.Errorf("unexepected token %s", d.curr)
		}
		d.next()
		for d.curr.Type != Rparen && !d.done() {
			
		}
		return nil
	}
	switch d.curr.Literal {
	case "count":
		d.next()
		count, err := strconv.Atoi(d.curr.Literal)
		if err != nil {
			return err
		}
		dom.Ticks = count
		d.next()
	case "position":
		d.next()
		dom.Position = d.curr.Literal
		d.next()
	case "label":
		d.next()
		dom.Label = d.curr.Literal
		d.next()
	case "format":
		d.next()
		dom.Format = d.curr.Literal
		d.next()
	case "inner-ticks":
		d.next()
		ok, err := strconv.ParseBool(d.curr.Literal)
		if err != nil {
			return err
		}
		dom.InnerTicks = ok
		d.next()
	case "outer-ticks":
		d.next()
		ok, err := strconv.ParseBool(d.curr.Literal)
		if err != nil {
			return err
		}
		dom.OuterTicks = ok
		d.next()
	case "label-ticks":
		d.next()
		ok, err := strconv.ParseBool(d.curr.Literal)
		if err != nil {
			return err
		}
		dom.LabelTicks = ok
		d.next()
	case "band-ticks":
		d.next()
		ok, err := strconv.ParseBool(d.curr.Literal)
		if err != nil {
			return err
		}
		dom.BandTicks = ok
		d.next()
	default:
		return fmt.Errorf("%s unsupported/unknown option for ticks", d.curr.Literal)
	}
	if d.curr.Type != EOL && d.curr.Type != EOF {
		return fmt.Errorf("expected end of line, got %s", d.curr)
	}
	d.next()
	return nil
}

func (d *Decoder) decodeLoad(cfg *Config) error {
	d.next()
	var fi File
	fi.Path = d.curr.Literal
	d.next()
	if d.curr.Type != Keyword && d.curr.Literal != kwUsing {
		return fmt.Errorf("unexpected word %s", d.curr)
	}
	d.next()
	return nil
}

func (d *Decoder) decodeRender(cfg *Config) error {
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
