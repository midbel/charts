package charts

import (
	"fmt"
)

type LineStyle int

const (
	StyleStraight LineStyle = 1 << iota
	StyleDotted
	StyleDashed
)

const (
	FontSize      = 12.0
	FontMonospace = "monospace"

	ColorBlack = "black"
	ColorBlue  = "blue"
)

const currentColour = "currentColour"

type Style struct {
	LineType    LineStyle
	LineColor   string
	LineWidth   float64
	LineOpacity float64

	FillOpacity float64
	FillList    []string

	FontSize     float64
	FontColor    string
	FontFamilies []string
	FontBold     bool
	FontItalic   bool

	Padding
}

func DefaultStyle() Style {
	return Style{
		LineStyle:    StyleStraight,
		LineColor:    ColorBlue,
		LineWidth:    1,
		LineOpacity:  1,
		FillList:     Tableau10,
		FillOpacity:  1,
		FontSize:     FontSize,
		FontFamilies: []string{FontMonospace},
		FontColor:    ColorBlack,
	}
}

type Padding struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

func PaddingFromList(list []float64) (Padding, error) {
	var pad Padding
	switch len(list) {
	case 1:
		pad.Top = list[0]
		pad.Right = list[0]
		pad.Bottom = list[0]
		pad.Left = list[0]
	case 2:
		pad.Top, pad.Bottom = list[0], list[0]
		pad.Right, pad.Left = list[1], list[1]
	case 3:
		pad.Top = list[0]
		pad.Bottom = list[2]
		pad.Right, pad.Left = list[1], list[1]
	case 4:
		pad.Top = list[0]
		pad.Right = list[1]
		pad.Bottom = list[2]
		pad.Left = list[3]
	default:
		return pad, fmt.Errorf("padding: expected 1, 2, 3 or 4 values! got %d", len(list))
	}
	return pad, nil
}

func (p Padding) Horizontal() float64 {
	return p.Left + p.Right
}

func (p Padding) Vertical() float64 {
	return p.Top + p.Bottom
}
