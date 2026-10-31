package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// rawFrame is one captured frame handed to the writer goroutine. Decoding
// (base64/hex/allocations) happens on the writer side so the relay pumps only
// pay for a buffer copy.
type rawFrame struct {
	srcHost, dstHost string
	srcPort, dstPort int
	data             []byte // pooled copy; owned by the writer until row build
}

// rawFramePool reuses the per-frame copies made by LogMain/LogSecondary. The
// writer returns each buffer after building its CSV row.
var rawFramePool = sync.Pool{
	New: func() any { return make([]byte, 0, 4096) },
}

// CSVLogger appends decoded packets to a CSV file progressively: a background
// goroutine consumes captured frames from a channel, decodes and writes each
// row as it arrives, so data already captured survives an abrupt process kill.
// Enqueue is non-blocking: when the queue is full the frame is dropped (the
// capture is best-effort and must never stall the relayed game traffic).
type CSVLogger struct {
	ch      chan rawFrame
	done    chan struct{}
	wg      sync.WaitGroup
	logName string
	once    sync.Once
	dropped atomic.Int64
}

// NewCSVLogger initializes the progressive CSV logger.
func NewCSVLogger(logName string) *CSVLogger {
	l := &CSVLogger{
		ch:      make(chan rawFrame, 4096),
		done:    make(chan struct{}),
		logName: logName,
	}
	l.wg.Add(1)
	go l.writeLoop()
	return l
}

// LogMain records transactions of the primary TCP flow. Never blocks: the
// frame is copied to a pooled buffer and dropped when the writer is behind.
func (l *CSVLogger) LogMain(srcHost string, srcPort int, destHost string, destPort int, data []byte) {
	l.log(srcHost, srcPort, destHost, destPort, data)
}

// LogSecondary records secondary target responses (best-effort capture).
func (l *CSVLogger) LogSecondary(srcHost string, srcPort int, destHost string, destPort int, data []byte) {
	l.log(srcHost, srcPort, destHost, destPort, data)
}

func (l *CSVLogger) log(srcHost string, srcPort int, destHost string, destPort int, data []byte) {
	buf := rawFramePool.Get().([]byte)
	buf = append(buf[:0], data...)
	select {
	case l.ch <- rawFrame{srcHost: srcHost, srcPort: srcPort, dstHost: destHost, dstPort: destPort, data: buf}:
	default:
		rawFramePool.Put(buf)
		l.dropped.Add(1)
	}
}

// writeLoop opens the CSV once, writes the header, then decodes and appends
// each captured frame as it arrives. Rows are flushed in batches (whenever the
// queue drains) and the file is synced periodically so a hard kill loses at
// most the frames still queued in memory.
func (l *CSVLogger) writeLoop() {
	defer l.wg.Done()

	f, err := os.OpenFile(l.logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("[PROXY ERROR] Cannot open CSV %s: %v", l.logName, err)
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	w := csv.NewWriter(f)
	if err := w.Write(csvHeader); err != nil {
		log.Printf("[PROXY ERROR] CSV header: %v", err)
		return
	}
	w.Flush()

	syncTicker := time.NewTicker(time.Second)
	defer syncTicker.Stop()

	write := func(rf rawFrame) {
		p := decodePacket(rf.data, rf.srcHost, strconv.Itoa(rf.srcPort), rf.dstHost, strconv.Itoa(rf.dstPort))
		_ = w.Write(p.row())
		rawFramePool.Put(rf.data)
	}

	for {
		select {
		case rf := <-l.ch:
			write(rf)
			// Coalesce whatever else is already queued into this flush.
		drain:
			for {
				select {
				case rf := <-l.ch:
					write(rf)
				default:
					break drain
				}
			}
			w.Flush()
		case <-l.done:
			// Drain whatever is still queued before stopping.
			for {
				select {
				case rf := <-l.ch:
					write(rf)
				default:
					w.Flush()
					_ = f.Sync()
					if n := l.dropped.Load(); n > 0 {
						log.Printf("[PROXY WARNING] %s: dropped %d frame(s) the capture queue could not absorb", l.logName, n)
					}
					return
				}
			}
		case <-syncTicker.C:
			_ = f.Sync()
		}
	}
}

// Close stops the writer goroutine after flushing pending packets.
func (l *CSVLogger) Close() error {
	l.once.Do(func() {
		close(l.done)
		l.wg.Wait()
	})
	log.Printf("[PROXY] Decoded CSV written to %s", l.logName)
	return nil
}
