package charts

import (
	"fmt"

	"github.com/midbel/svg"
	// "github.com/midbel/slices"
)

type LineStyle int

const (
	StyleStraight LineStyle = 1 << iota
	StyleSolid
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
	FillStyle   string

	FontSize   float64
	FontColor  string
	FontFamily []string
	FontBold   bool
	FontItalic bool

	Padding
}

func DefaultStyle() Style {
	return Style{
		LineType:    StyleStraight,
		LineColor:   ColorBlue,
		LineWidth:   1,
		LineOpacity: 1,
		FillList:    Tableau10,
		FillOpacity: 1,
		FontSize:    FontSize,
		FontFamily:  []string{FontMonospace},
		FontColor:   ColorBlack,
	}
}

func (s Style) Rect() svg.Rect {
	var rect svg.Rect
	return rect
}

func (s Style) Text(str string) svg.Text {
	txt := svg.NewText(str)
	txt.Baseline = "middle"
	txt.Fill = svg.NewFill(s.FontColor)
	txt.Font = svg.NewFont(s.FontSize)
	txt.Font.Family = s.FontFamily
	if s.FontBold {
		txt.Font.Weight = "bold"
	}
	if s.FontItalic {
		txt.Font.Style = "italic"
	}
	return txt
}

func (s Style) LinePath() svg.Path {
	var pat svg.Path
	pat.Rendering = "geometricPrecision"
	pat.Stroke = svg.NewStroke(s.LineColor, s.LineWidth)
	pat.Stroke.Opacity = s.LineOpacity
	pat.Stroke.LineJoin = "round"
	pat.Stroke.LineCap = "round"
	pat.Fill = svg.NewFill("none")
	fmt.Printf("%+v\n", pat)

	switch s.LineType {
	case StyleStraight, StyleSolid:
	case StyleDotted:
		pat.Stroke.DashArray = append(pat.Stroke.DashArray, 1, 5)
	case StyleDashed:
		pat.Stroke.DashArray = append(pat.Stroke.DashArray, 10, 5)
	default:
	}
	return pat
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
