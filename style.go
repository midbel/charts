package charts

type LineStyle int

const (
	StyleStraight LineStyle = 1 << iota
	StyleDotted
	StyleDashed
)

const currentColour = "currentColour"

type Style struct {
	Line struct {
		Style   LineStyle
		Color   string
		Width   float64
		Opacity float64
	}
	Fill struct {
		Opacity float64
		List    []string
	}
	Text struct {
		Size     float64
		Color    string
		Families []string
		Bold     bool
		Italic   bool
	}
}
