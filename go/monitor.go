package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// NOTE: I'm not 100% sure this first sample is perfectly normal.

var port = flag.String("serial", "/dev/ttyACM0", "The serial port for the arduino")
var outfile = flag.String("outfile", "", "output file to write to")

func main() {
	flag.Parse()

	out := *outfile
	if out == "" {
		now := time.Now()
		out = fmt.Sprintf("heart_data_starting_%s", now.Format(time.RFC3339))
	}

	port, err := os.Open(*port)
	if err != nil {
		panic(err)
	}

	outF, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	writer := NewRecordWriter(outF)

	buf := bufio.NewReader(port)

	readings := make(chan uint16)
	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			by, err := buf.ReadBytes('\n')
			if err != nil {
				panic(err)
			}

			if len(by) == 1 {
				continue
			}

			by = by[:len(by)-1]

			i, err := strconv.Atoi(string(by))
			if err == nil && i < 65535 {
				readings <- uint16(i)
			} else {
				if i >= 65535 {
					continue
				}
				fmt.Println(err)
			}

		}
	}()

	heartReadings := make([]uint16, 0, 210)
	for {
		select {
		case r := <-readings:
			heartReadings = append(heartReadings, r)
		case <-ticker.C:
			if len(heartReadings) > 0 {
				writer.WriteTimestamp()
				var buf bytes.Buffer

				_ = binary.Write(&buf, binary.BigEndian, heartReadings)
				writer.WriteRecord(Heartdata, buf.Bytes())
				err := writer.buf.Flush()
				if err != nil {
					log.Printf("Error writing: %v", err)
				}
				heartReadings = heartReadings[:0]
			}
		}
	}
}
