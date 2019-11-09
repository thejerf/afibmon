package heartmon

import (
	"fmt"
	"io"
)

// This will output the records in a more human-friendly format for debug
// logging and verifying that connections are still alive.
func HumanReadableOutput(r io.Reader, w io.Writer) {
	rr := NewRecordReader(r)

	for {
		record, err := rr.NextRecord()
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Fprintf(w, "Can't read from stream: %v\n", err)
			return
		}

		switch r := record.(type) {
		case TimestampRecord:
			fmt.Fprintf(w, "Time: %s\n", r.Time)
		case HeartDataRecord:
			fmt.Fprintf(w, "Heart data: %v\n", r.Data)
		case ErrorRecord:
			fmt.Fprintf(w, "***\n*** ERROR: %v\n***\n", r.Error)
		}
	}
}
