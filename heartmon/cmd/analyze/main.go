package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/thejerf/afibmon/heartmon"
	"github.com/thejerf/afibmon/heartmon/beatalyse"
)

var chunkSize = flag.Int("chunksize", 512, "size of chunks to process")

func main() {
	flag.Parse()
	filename := os.Args[1]

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Can't open file %s: %v\n", filename, err)
		os.Exit(1)
	}

	records := heartmon.NewRecordReader(f)

	data := []uint16{}

	frame := 0
	analyzer := beatalyse.New(*chunkSize)
	var startishTime *time.Time
	for {
		record, err := records.NextRecord()
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintf(os.Stderr, "Error while reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		switch r := record.(type) {
		case heartmon.TimestampRecord:
			if startishTime == nil {
				startishTime = &r.Time
			}
		case heartmon.HeartDataRecord:
			data = append(data, r.Data...)
		}

		if len(data) < *chunkSize {
			continue
		}

		// Have enough data now
		chunk := data[:*chunkSize]
		data = data[*chunkSize:]

		f, err = os.Create("plotdata.tmp")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't open plotdata.tmp: %v\n", err)
			os.Exit(1)
		}

		analyzer.Analyze(chunk, f)
		f.Close()

		cmd := exec.Command("gnuplot",
			"-e",
			fmt.Sprintf(
				`
set yr [0:10000];
set terminal png size 3000,1500;
set output "freq_frames/frame%05d.png";
set title "freq - frame %05d - %s";
plot 'plotdata.tmp' with lines
`,
				frame, frame, startishTime.Format(time.RFC1123),
			),
		)
		err = cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't call gnuplot: %v\n",
				err)
			os.Exit(1)
		}

		f, err = os.Create("plotdata.tmp")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't open plotdata.tmp: %v\n", err)
			os.Exit(1)
		}
		for idx, value := range chunk {
			fmt.Fprintf(f, "%v %v\n", idx, value)
		}
		f.Close()

		cmd = exec.Command("gnuplot",
			"-e",
			fmt.Sprintf(`
set yr [0:800]; set terminal png size 3000,1500;
set output "amp_frames/frame%05d.png";
set title "amp - frame %05d - %s";
plot 'plotdata.tmp' with lines
`,
				frame, frame, startishTime.Format(time.RFC1123),
			),
		)
		err = cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't call gnuplot: %v\n",
				err)
			os.Exit(1)
		}
		f.Close()

		startishTime = nil
		frame++

		if frame%10 == 0 {
			fmt.Println("Frame", frame)
		}
	}
}

/*
ffmpeg
ffmpeg -r 15 -i frames/frame%05d.png -vcodec libx264 -crf 25 frequency.mp4
ffmpeg -r 15 -i heartframes/frame%05d.png -vcodec libx264 -crf 25 amplitude.mp4
*/
