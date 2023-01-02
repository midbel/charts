package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("a", ":8080", "listening address")
	flag.Parse()
	var (
		dir = http.Dir(flag.Arg(0))
		err = http.ListenAndServe(*addr, http.FileServer(dir))
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
