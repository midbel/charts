package charts

type Style struct {
	Line struct {
		Style   LineStyle
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
	}
}
