package main

import (
	"bufio"
	"encoding/binary"
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

func (rw *RecordWriter) WriteRecord(recordType byte, record []byte) error {
	_ = rw.buf.WriteByte(recordType)
	_ = binary.Write(rw.buf, binary.BigEndian, uint16(len(record)))
	_, err := rw.buf.Write(record)
	return err
}

func (rw *RecordWriter) WriteTimestamp() error {
	now := time.Now()
	nano := now.UnixNano()

	eightbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(eightbytes, uint64(nano))

	return rw.WriteRecord(Timestamp, eightbytes)
}
