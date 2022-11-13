package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/charts/decode"
)

func main() {
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer r.Close()

	cfg, err := decode.NewDecoder(r).Decode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to decode input file: %s", err)
		fmt.Fprintln(os.Stderr)
		os.Exit(2)
	}
	if err = cfg.Render(); err != nil {
		fmt.Fprintf(os.Stderr, "fail to render chart from input file: %s", err)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
}
