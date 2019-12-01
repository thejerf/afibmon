package main

import (
	"fmt"
	"os"

	"github.com/thejerf/afibmon/heartmon"
)

func main() {
	filename := os.Args[1]

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Can't open file %q: %v\n", filename, err)
		os.Exit(1)
	}

	rr := heartmon.NewRateDetector(f, os.Stdout)
	rr.Run()
}
