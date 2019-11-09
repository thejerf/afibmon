package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
)

var port = flag.String("serial", "/dev/ttyACM0", "The serial port for the arduino")
var outfile = flag.String("outfile", "", "output file to write to")

func main() {
	flag.Parse()

	readings := make(chan uint16)

	port, err := os.Open(*port)
	if err != nil {
		panic(err)
	}
	buf := bufio.NewReader(port)

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
			// every once in a while the serial port gets overloaded
			// or the heart thing gets confused and sends invalid
			// inputs that appear to be two numbers mixed together.
			// While I'd love to fix that on the Arduino side, in the
			// meantime we need to handle it. 1024 is the top end of
			// what it is supposed to emit:
			// https://www.arduino.cc/en/Reference/AnalogRead
			// and most such distortions appear to result in numbers larger
			// than that.
			// Also, as far as FFTs are concerned I'm pretty sure it's much
			// better to drop a sample than to put a spurious transient in
			// it, which will spray high frequencies everywhere.
			if err == nil && i < 1024 {
				readings <- uint16(i)
			} else {
				if i >= 65535 {
					continue
				}
				fmt.Println(err)
			}
		}
	}()

}
