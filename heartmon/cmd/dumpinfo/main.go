package main

import (
	"fmt"
	"io"
	"os"

	"github.com/thejerf/afibmon/heartmon"
)

func main() {
	filename := os.Args[1]

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Can't open file %s: %v\n", filename, err)
		os.Exit(1)
	}

	records := heartmon.NewRecordReader(f)

	data := []uint16{}

	for {
		record, err := records.NextRecord()
		if err != nil {
			if err == io.EOF {
				for idx, datum := range data {
					fmt.Println(idx, datum)
				}
				return
			}

			fmt.Fprintf(os.Stderr, "Error while reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		switch r := record.(type) {
		case heartmon.TimestampRecord:
			// discard for now
		case heartmon.HeartDataRecord:
			data = append(data, r.Data...)
		}
	}

}
