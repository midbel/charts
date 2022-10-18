package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/charts/dsl"
)

func main() {
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer r.Close()

	if err = dsl.NewDecoder(r).Decode(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
