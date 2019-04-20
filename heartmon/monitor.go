package heartmon

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// NOTE: I'm not 100% sure this first sample is perfectly normal.

type MonitorReader struct {
	outfile  string
	incoming chan uint16

	stop chan struct{}

	sync.Mutex
	// There's a bit of dodginess here, since if any of these block
	// they end up blocking the whole shebang. It's why we at the very
	// least hand out very large buffered channels.
	subscriptions []chan uint16
}

// Subscribe hands out channels that will echo the results coming back from
// the heart monitor.
//
// The channel will be buffered with 2048 slots (which is still only 4K
// worth of space), which is enough for about ten seconds. In addition, if
// you are doing anything that will potentially take time, such as sending
// a packet, consider it an obligation to drain the channel for each
// packet, rather than sending one value per packet, so that you don't fall
// behind, because if you block this channel you will block the entire
// heart data collection procedure. There's a lot of buffers in the system
// but eventually they will fill up.
func (mr *MonitorReader) Subscribe() chan uint16 {
	mr.Lock()
	defer mr.Unlock()

	// enough for approximately 10 entire seconds of data
	subscription := make(chan uint16, 2048)
	mr.subscriptions = append(mr.subscriptions, subscription)
	return subscription
}

// Unsubscribe will remove the given channel from the subscription list.
//
// If you have enough subscriptions that a linear scan over an array
// is causing actual performance issues, you've probably got bigger
// problems.
func (mr *MonitorReader) Unsubscribe(c chan uint16) {
	mr.Lock()
	defer mr.Unlock()

	for idx, subscriber := range mr.subscriptions {
		if subscriber == c {
			mr.subscriptions = append(mr.subscriptions[:idx],
				mr.subscriptions[idx+1:]...)
			return
		}
	}
}

func (mr *MonitorReader) Serve() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	out := mr.outfile
	if out == "" {
		now := time.Now()
		out = fmt.Sprintf("heart_data_starting_%s", now.Format(time.RFC3339))
	}

	outF, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	writer := NewRecordWriter(outF)

	// if append ends up growing this, it's not a catastrophe; 210 is just
	// a sizing guess. The grown buffer would end up reused anyhow.
	heartReadings := make([]uint16, 0, 210)
	subscribers := []chan uint16{}
	for {
		select {
		case r := <-mr.incoming:
			heartReadings = append(heartReadings, r)

			subscribers = subscribers[:0]
			mr.Lock()
			subscribers = append(subscribers, mr.subscriptions...)
			mr.Unlock()
			for _, subscriber := range subscribers {
				// FIXME: Use a select/default mechanism here
				// to forcibly unsubscribe something that
				select {
				case subscriber <- r:
					// do nothing on purpose
				default:
					close(subscriber)
					mr.Unsubscribe(subscriber)
				}
			}

		case _, _ = <-mr.stop:
			return

		case <-ticker.C:
			if len(heartReadings) > 0 {
				writer.WriteTimestamp()
				writer.WriteRecord(Heartdata, HeartDataRecord{heartReadings})
				err := writer.buf.Flush()
				if err != nil {
					log.Printf("Error writing: %v", err)
				}
				heartReadings = heartReadings[:0]
			}
		}
	}
}

func (mr *MonitorReader) Stop() {
	close(mr.stop)
}
