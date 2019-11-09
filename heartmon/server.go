package heartmon

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type Server struct {
	l net.Listener
}

func (s *Server) Serve() {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			return
		}

		go newInstance(conn)
	}
}

func (s *Server) Stop() {
	s.l.Close()
}

type Instance struct {
	input io.Reader
}

func NewServer(address string) (*Server, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Server{l}, nil
}

func newInstance(conn io.Reader) {
	fmt.Println("Connection started")
	now := time.Now()
	ts := now.Format(time.RFC3339)
	filename := fmt.Sprintf("heartbeat_starting_%s.hrt", ts)
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Couldn't create %s: %v", filename, err)
		return
	}

	filename2 := fmt.Sprintf("human_heartbeat_%s.txt", ts)
	f2, err := os.Create(filename2)
	if err != nil {
		log.Printf("Couldn't create %s: %v", filename2, err)
		return
	}

	hrrR, hrrW := io.Pipe()
	stderrR, stderrW := io.Pipe()

	packets := io.TeeReader(
		io.TeeReader(
			conn,
			hrrW, // write the packets to the human readable file
		),
		stderrW, // write all the packets to the standard error
		// human-readable output
	)

	go HumanReadableOutput(hrrR, f2)
	go HumanReadableOutput(stderrR, os.Stderr)
	io.Copy(f, packets)
}
