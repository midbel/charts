module github.com/midbel/charts

go 1.19

require (
	github.com/midbel/slices v0.1.1
	github.com/midbel/svg v0.0.1
	github.com/midbel/oryx v0.0.1
	golang.org/x/sync v0.0.0-20220929204114-8fcdb60fdcc0
)

replace github.com/midbel/svg v0.0.1 => ../svg
replace github.com/midbel/oryx v0.0.1 => ../oryx

replace github.com/midbel/slices v0.1.1 => ../slices
