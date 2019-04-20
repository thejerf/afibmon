package heartmon

import (
	"fmt"
	"log"
	"net/http"
)

// LiveHeartEvents receives the incoming live heart events. This is
// intended for local use, as it will push a single packet for every
// incoming heart event, which is not a great plan unless you're on
// a local network.
type LiveHeartEvents struct {
	rp *MonitorReader
}

func (lhe *LiveHeartEvents) ServeHttp(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.WriteHeader(200)

	events := lhe.rp.Subscribe()

	flusher, isFlusher := rw.(http.Flusher)

	inputs := make([]uint16, 0, 32)

	for {
		inputs = inputs[:0]

		// drain the channel, so that even if we do get a bit behind
		// we do our best to not block the monitor goroutine. This
		// technically is probably too fragile to deploy to a high-load
		// cloud server or something but in this application is probably
		// fine.
		for {
			select {
			case result, ok := <-events:
				if !ok {
					log.Printf("LiveHeartEvents forcibly unsubscribed")
					return
				}
				inputs = append(inputs, result)
			default:
				break
			}
		}

		// "lazily" use the default Go fmt.Printf format for slices
		// of integers
		_, err := rw.Write([]byte(fmt.Sprintf("data: %v\n\n", inputs)))
		if err != nil {
			return
		}
		if isFlusher {
			flusher.Flush()
		}
	}
}
