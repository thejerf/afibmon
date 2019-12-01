package heartmon

import (
	"bufio"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

const (
	Timestamp = byte(1)
	Heartdata = byte(2)
	Error     = byte(3)
)

// This defines a simple record-based format that allows us to mark
// timestampse periodically into the output file, while allowing us room
// for future additions if necessary. Not sure what those would be, but
// best to leave the room.

type RecordWriter struct {
	buf *bufio.Writer
}

func NewRecordWriter(w io.Writer) *RecordWriter {
	return &RecordWriter{
		buf: bufio.NewWriter(w),
	}
}

func (rw *RecordWriter) WriteRecord(
	recordType byte,
	record encoding.BinaryMarshaler) error {
	_ = rw.buf.WriteByte(recordType)
	output, err := record.MarshalBinary()
	if err != nil {
		return err
	}
	_ = binary.Write(rw.buf, binary.BigEndian, uint16(len(output)))
	_, err = rw.buf.Write(output)
	return err
}

func (rw *RecordWriter) WriteTimestamp() error {
	return rw.WriteRecord(Timestamp, TimestampRecord{time.Now()})
}

type RecordReader struct {
	buf *bufio.Reader
}

func NewRecordReader(r io.Reader) *RecordReader {
	return &RecordReader{bufio.NewReader(r)}
}

// FIXME: We really ought to observe the io.EOF coming in at the correct
// location.
func (rr *RecordReader) NextRecord() (Record, error) {
	ty, err := rr.buf.ReadByte()
	if err != nil {
		return nil, err
	}
	twoB := make([]byte, 2)
	_, err = io.ReadFull(rr.buf, twoB)
	if err != nil {
		return nil, err
	}
	l := binary.BigEndian.Uint16(twoB)
	record := make([]byte, int(l))
	_, err = io.ReadFull(rr.buf, record)
	if err != nil {
		return nil, err
	}
	switch ty {
	case Timestamp:
		r := TimestampRecord{}
		err := r.UnmarshalBinary(record)
		if err != nil {
			return nil, err
		}
		return r, nil
	case Heartdata:
		r := HeartDataRecord{}
		err := r.UnmarshalBinary(record)
		if err != nil {
			return nil, err
		}
		return r, nil
	case Error:
		r := ErrorRecord{}
		err = r.UnmarshalBinary(record)
		if err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, errors.New("unknown record type")
	}
}

type Record interface {
	isRecord()
}

type TimestampRecord struct {
	Time time.Time
}

// MarshalBinary marshals the timestamp record into a binary stream.
func (tr TimestampRecord) MarshalBinary() ([]byte, error) {
	nano := tr.Time.UnixNano()
	eightbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(eightbytes, uint64(nano))
	return eightbytes, nil
}

func (tr *TimestampRecord) UnmarshalBinary(b []byte) error {
	switch len(b) {
	case 4:
		epoch := binary.BigEndian.Uint32(b)
		tr.Time = time.Unix(int64(epoch), 0)
		return nil
	case 8:
		nanoseconds := binary.BigEndian.Uint64(b)
		tr.Time = time.Unix(0, int64(nanoseconds))
		return nil
	default:
		return errors.New("Illegal size timestamp")
	}

}

func (tr TimestampRecord) isRecord() {}

type HeartDataRecord struct {
	Data []uint16
}

// MarshalBinary will write the heart data record to a []byte.
func (hdr HeartDataRecord) MarshalBinary() ([]byte, error) {
	b := make([]byte, 2*len(hdr.Data), 2*len(hdr.Data))

	for i := 0; i < len(hdr.Data); i++ {
		binary.BigEndian.PutUint16(b[2*i:2*i+2], hdr.Data[i])
	}

	return b, nil
}

func (hdr *HeartDataRecord) UnmarshalBinary(b []byte) error {
	if len(b)%2 != 0 {
		return errors.New("Illegal size heart record")
	}
	if len(b) == 0 {
		return nil
	}

	data := make([]uint16, len(b)/2, len(b)/2)
	for i := 0; i < len(b)/2; i++ {
		data[i] = binary.BigEndian.Uint16(b[i*2 : i*2+2])
	}

	hdr.Data = data
	return nil
}

func (hdr HeartDataRecord) isRecord() {}

type ErrorRecord struct {
	Error string
}

func (er ErrorRecord) MarshalBinary() ([]byte, error) {
	return []byte(er.Error), nil
}

func (er *ErrorRecord) UnmarshalBinary(b []byte) error {
	er.Error = string(b)
	return nil
}

func (er ErrorRecord) isRecord() {}

// RateDetector uses a bit of a hacky approach to detect the heart rate.
type RateDetector struct {
	WriteTimestamp bool
	rr             *RecordReader
	output         io.Writer
	alerter        *Alerter

	buffer []uint16
}

// NewRateDetector returns a new RateDetector.
func NewRateDetector(r io.Reader, w io.Writer) *RateDetector {
	alerter := NewAlerter(w)

	return &RateDetector{
		false,
		NewRecordReader(r),
		w,
		alerter,
		nil,
	}
}

// This design allows us to take a pre-existing stream of heart info and
// stream it through, reporting when all the alerts would have been.

func (rr *RateDetector) Run() {
	consequetiveBad := 0
	limit := 90

	lastTime := time.Time{}

	for {
		record, err := rr.rr.NextRecord()

		keep := 50*60 + 1

		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Fprintf(rr.output, "Can't read from stream: %v\n", err)
			return
		}

		switch r := record.(type) {
		case TimestampRecord:
			fmt.Fprintf(rr.output, "Time: %s\n",
				r.Time.Format(time.RFC1123),
			)
			lastTime = r.Time
		case ErrorRecord:
			// Reset the buffer due to error
			rr.buffer = []uint16{}

		case HeartDataRecord:
			rr.buffer = append(rr.buffer, r.Data...)
			// trim to 60 seconds + 1 sample assuming 50Hz sample rate
			samples := len(rr.buffer)
			if samples > keep {
				rr.buffer = rr.buffer[samples-keep:]
			}

			bpm := DetectHeartbeats(rr.buffer)

			fmt.Fprintf(rr.output, "Beats per minute: %d\n",
				bpm)

			if bpm > limit {
				consequetiveBad++
			} else {
				consequetiveBad = 0
			}

			if consequetiveBad > 20 {
				rr.alerter.Alert(lastTime)
			} else {
				rr.alerter.Stop(lastTime)
			}
		}
	}
}

var stateNormal = 0
var stateLow = 1

func DetectHeartbeats(ecg []uint16) int {
	// now, do our crappy heartbeat detection:
	// 1. Take simple derivative of the heartbeat rate.
	// 2. Look for anything that is < -100 with a > 100 value 1 or
	//    2 in front of it.
	// 3. Call that a heartbeat.
	// Based on just eyeballing it, it's not necessarily that bad.
	//
	// It turns out that this is decent at detecting normal heartbeats, but
	// if you apply it to the moments I seem to be fibrellating, it tends
	// to false positive high. However, under the circumstances... that's
	// not all bad....

	state := stateNormal

	derivative := make([]int16, len(ecg)-1)
	heartbeats := 0
	for idx := range ecg {
		if idx == 0 {
			continue
		}

		derivative[idx-1] = int16(ecg[idx]) - int16(ecg[idx-1])

		switch state {
		case stateNormal:
			if derivative[idx-1] < -50 {
				heartbeats++
				state = stateLow
			}

		case stateLow:
			if derivative[idx-1] > 25 {
				state = stateNormal
			}
		}
	}

	return heartbeats
}

type Alerter struct {
	playing   bool
	logstream io.Writer

	cmd *exec.Cmd
}

func NewAlerter(out io.Writer) *Alerter {
	return &Alerter{logstream: out}
}

func (a *Alerter) Alert(now time.Time) {
	if a.cmd != nil {
		return
	}

	fmt.Fprintf(a.logstream, "Starting alert at %s\n", now)

	a.cmd = exec.Command("mplayer", "/home/jerf/StormSounds.m4a")
	err := a.cmd.Start()
	if err != nil {
		fmt.Fprintf(a.logstream, "Couldn't start audio: %v\n", err)
		a.cmd = nil
	}
}

func (a *Alerter) Stop(now time.Time) {
	if a.cmd == nil {
		return
	}

	fmt.Fprintf(a.logstream, "Stopping alert at %s\n", now)

	err := a.cmd.Process.Kill()
	if err != nil {
		fmt.Fprintf(a.logstream, "Can't kill mplayer: %s\n", err)
	}
	a.cmd.Wait()
	a.cmd = nil
}
