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
var analysis = flag.String("analysis",
	"freq_and_amp", "analysis to perform")

func main() {
	flag.Parse()
	filename := flag.Arg(0)

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Can't open file %s: %v\n", filename, err)
		os.Exit(1)
	}

	records := heartmon.NewRecordReader(f)

	data := []uint16{}

	frame := 0
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

		switch *analysis {
		case "freq_and_amp":
			err = plotFreqAndAmp(frame, chunk, *startishTime)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Couldn't render frame: %v",
					err)
				os.Exit(1)
			}

		case "amp_buckets":
			err = plotAmpBuckets(frame, chunk, *startishTime)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Couldn't render frame: %v",
					err)
				os.Exit(1)
			}

		default:
			fmt.Fprint(os.Stderr, "Unknown analysis %q\n",
				*analysis)
			os.Exit(1)
		}

		startishTime = nil
		frame++
		if frame%25 == 0 {
			fmt.Println("Frame", frame)
		}
	}
}

func plotAmpBuckets(
	frame int,
	chunk []uint16,
	startishTime time.Time,
) error {
	var err error

	analyzer := beatalyse.New(*chunkSize)

	f, err := os.Create("plotdata_amp.tmp")
	if err != nil {
		return fmt.Errorf("Can't open plotdata_amp.tmp: %v\n", err)
	}
	for idx, value := range chunk {
		fmt.Fprintf(f, "%v %v\n", idx, value)
	}
	f.Close()

	f, err = os.Create("plotdata.ratios")

	fft := analyzer.FFT(chunk)
	buckets := fft.Buckets(10)
	normalized := buckets.Normalized()

	var idx int
	var value float64
	for idx, value = range normalized {
		fmt.Fprintf(f, "%v %v\n", idx, value)
	}
	idx++
	fmt.Fprintf(f, "%v 0\n", idx)
	f.Close()

	// For this style of analysis, we want the big plot of normal
	// amplitude, with an embedded bar graph of the ratios of the various
	// FFT buckets.
	gnuplotProgram := fmt.Sprintf(
		`set terminal png size 3000,1500;
set output "frame_%05d.png";
set title "freq - frame %05d - %s";

set multiplot;

set origin 0, 0;
set size 1, 1;
set yr [0:800];
set object 1 rectangle from graph 0,0 to graph 1,1 behind fillcolor rgb'%s' fillstyle solid noborder;
plot 'plotdata_amp.tmp' with lines;

set size 0.2,0.2;
set origin 0.75,0.75;

set yr [0:3];
set style fill solid;
plot 'plotdata.ratios' with fillsteps;

unset multiplot;
`,
		frame, frame, startishTime.Format(time.RFC1123), "0xFFFFEE")
	cmd := exec.Command("gnuplot",
		"-e",
		gnuplotProgram,
	)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Couldn't call gnuplot: %v\n", err)
	}
	return nil
}

func plotFreqAndAmp(
	frame int,
	chunk []uint16,
	startishTime time.Time,
) error {
	var err error

	analyzer := beatalyse.New(*chunkSize)

	f, err := os.Create("plotdata.tmp")
	if err != nil {
		return fmt.Errorf("Can't open plotdata.tmp: %v\n", err)
	}

	fft := analyzer.FFT(chunk)
	fft.DumpText(f)
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
		return fmt.Errorf("Couldn't call gnuplot: %v\n",
			err)
	}

	f, err = os.Create("plotdata.tmp")
	if err != nil {
		return fmt.Errorf("Can't open plotdata.tmp: %v\n", err)
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
		fmt.Errorf("Couldn't call gnuplot: %v\n",
			err)
	}
	return nil
}

/*
ffmpeg commands:
for amp and freq:
ffmpeg -r 15 -i frames/frame%05d.png -vcodec libx264 -crf 25 frequency.mp4
ffmpeg -r 15 -i heartframes/frame%05d.png -vcodec libx264 -crf 25 amplitude.mp4

for amp_buckets:
ffmpeg -r 15 -i frames/frame%05d.png -vcodec libx264 -crf 25 amp_buckets.mp4

*/
