package heartmon

import (
	"bufio"
	"encoding"
	"encoding/binary"
	"errors"
	"io"
	"time"
)

const (
	Timestamp = byte(1)
	Heartdata = byte(2)
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
	default:
		return nil, errors.New("unknown record type")
	}
}

type Record interface {
	isRecord()
}

type TimestampRecord struct {
	time time.Time
}

// MarshalBinary marshals the timestamp record into a binary stream.
func (tr TimestampRecord) MarshalBinary() ([]byte, error) {
	nano := tr.time.UnixNano()
	eightbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(eightbytes, uint64(nano))
	return eightbytes, nil
}

func (tr *TimestampRecord) UnmarshalBinary(b []byte) error {
	if len(b) != 8 {
		return errors.New("Illegal size timestamp")
	}
	nanoseconds := binary.BigEndian.Uint64(b)
	tr.time = time.Unix(0, int64(nanoseconds))
	return nil
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
