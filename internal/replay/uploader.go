package replay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/klauspost/compress/zlib"
	"github.com/valyala/fasthttp"
)

type UploadManager struct {
	host   string
	port   uint
	method string
	addr   string
	conn   net.Conn
	// connReader buffers reads for the response parser; it must live as long
	// as the connection so buffered bytes survive across requests.
	connReader  *bufio.Reader
	connMu      sync.Mutex
	lastConnect time.Time
	sendQueue   chan *Info
	pool        []*Info
	poolMu      sync.Mutex
	// inflight counts pages currently being pushed; it is a diagnostic gauge
	// (uploads are serial on the process loop), kept atomic so any goroutine
	// may read it without touching connMu.
	inflight atomic.Int32

	// seqMu/sequences assign each session an increasing upload sequence: the
	// replay host keys pages by matchid+sequenceid, so a constant 0 would make
	// every page of a match collide under the same key.
	seqMu     sync.Mutex
	sequences map[int64]uint32
}

type Info struct {
	SessionID  int64
	SequenceID uint32
	Data       []byte
	DataLength uint32
	RefCount   uint32
}

func NewReplayUploadManager(host string, port uint, method string) *UploadManager {
	return &UploadManager{
		host:      host,
		port:      port,
		method:    method,
		addr:      net.JoinHostPort(host, fmt.Sprintf("%d", port)),
		sendQueue: make(chan *Info, 100),
		pool:      make([]*Info, 0, 100),
	}
}

// dialTimeout bounds every connect attempt so a firewalled replay host cannot
// hang the upload loop on the OS-level connect timeout.
const dialTimeout = 5 * time.Second

func (m *UploadManager) dial() (net.Conn, error) {
	d := net.Dialer{Timeout: dialTimeout}
	return d.Dial("tcp", m.addr)
}

func (m *UploadManager) Start(ctx context.Context) error {
	conn, err := m.dial()
	if err != nil {
		return fmt.Errorf("dial replay host: %w", err)
	}

	m.connMu.Lock()
	m.conn = conn
	m.connReader = bufio.NewReader(conn)
	m.connMu.Unlock()

	go m.processLoop(ctx)

	log.Printf("ReplayUploadManager connected to %s (method: %s)", m.addr, m.method)
	return nil
}

// reconnectInterval throttles reconnect dials to one per 5s so a down replay
// host is not hammered on every queued message.
const reconnectInterval = 5 * time.Second

// connect dials a new connection when the current one is gone. It is throttled
// to one attempt every 5 seconds so a down replay host is not hammered on every
// queued message.
func (m *UploadManager) connect() error {
	m.connMu.Lock()
	defer m.connMu.Unlock()

	if m.conn != nil {
		return nil
	}
	if time.Since(m.lastConnect) < reconnectInterval {
		return fmt.Errorf("reconnect throttled")
	}
	m.lastConnect = time.Now()

	conn, err := m.dial()
	if err != nil {
		return fmt.Errorf("dial replay host: %w", err)
	}

	m.conn = conn
	m.connReader = bufio.NewReader(conn)
	log.Printf("ReplayUploadManager reconnected to %s", m.addr)
	return nil
}

// OnProxyData queues a Proxy (0x10) message body for upload as the replay data
// of the given match. The session ID is not optional bookkeeping: the replay
// host keys every page by matchid+sequenceid, so uploading without one would
// file two concurrent matches' pages under the same key.
func (m *UploadManager) OnProxyData(sessionID int64, data []byte) {
	m.PackReplayData(sessionID, data)
}

func (m *UploadManager) PackReplayData(sessionID int64, data []byte) {
	info := m.getFromPool()
	info.SessionID = sessionID
	info.SequenceID = m.nextSequence(sessionID)
	info.Data = make([]byte, len(data))
	copy(info.Data, data)
	info.DataLength = uint32(len(data))
	info.RefCount = 1

	m.OnAddReplayData(info)
}

// nextSequence returns the next upload sequence for the session (1-based).
// Pages are keyed by matchid+sequenceid on the replay host.
func (m *UploadManager) nextSequence(sessionID int64) uint32 {
	m.seqMu.Lock()
	defer m.seqMu.Unlock()
	if m.sequences == nil {
		m.sequences = make(map[int64]uint32)
	}
	m.sequences[sessionID]++
	return m.sequences[sessionID]
}

// OnAddReplayData queues a raw page for upload. Compression deliberately
// happens on the upload loop (see processLoop), NOT here: this runs on the
// game-message goroutine, and a level-9 deflate of every page would stall
// message handling.
func (m *UploadManager) OnAddReplayData(info *Info) {
	select {
	case m.sendQueue <- info:
	default:
		log.Printf("Replay send queue full, dropping")
		m.returnToPool(info)
	}
}

// zlibWriterPool reuses zlib writers across uploads. A deflate compressor
// allocates ~1MB of hash/window buffers at level 9, so recreating it per upload
// (uploads run on a timer) would churn that allocation each time. klauspost's
// Writer keeps its compressor across Reset calls, so pooled writers stay warm.
var zlibWriterPool = sync.Pool{
	New: func() any {
		w, err := zlib.NewWriterLevel(io.Discard, zlib.BestCompression)
		if err != nil {
			panic(err) // BestCompression (9) is always valid
		}
		return w
	},
}

func (m *UploadManager) convertToBase64(data []byte) ([]byte, error) {
	var compressed bytes.Buffer
	w := zlibWriterPool.Get().(*zlib.Writer)
	w.Reset(&compressed)

	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		zlibWriterPool.Put(w)
		return nil, err
	}
	if err := w.Close(); err != nil {
		zlibWriterPool.Put(w)
		return nil, err
	}
	zlibWriterPool.Put(w)

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(compressed.Bytes())))
	base64.StdEncoding.Encode(encoded, compressed.Bytes())
	return encoded, nil
}

func (m *UploadManager) processLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case info := <-m.sendQueue:
			// Compress off the game-message path: this loop owns the upload,
			// so a slow deflate only delays uploads, never gameplay.
			encoded, err := m.convertToBase64(info.Data)
			if err != nil {
				log.Printf("ConvertBinaryToString failed: %v", err)
				m.returnToPool(info)
				continue
			}
			info.Data = encoded
			info.DataLength = uint32(len(encoded))
			m.pushReplayData(info)
		}
	}
}

// sendRequest writes a single replay request to the persistent connection and
// consumes the response (with keep-alive, unread responses accumulate in kernel
// buffers until the connection wedges). Reconnects (with throttling) when there
// is no usable connection.
//
// Only connect/write failures are reported to the caller for retry: once the
// request bytes are out, a lost or truncated RESPONSE must not trigger a
// re-post — the upload already reached the host, and racing closers (Stop,
// dropConn from another path) would otherwise turn into duplicate deliveries.
func (m *UploadManager) sendRequest(req *fasthttp.Request) error {
	if err := m.connect(); err != nil {
		return err
	}

	m.connMu.Lock()
	conn := m.conn
	reader := m.connReader
	m.connMu.Unlock()

	if conn == nil || reader == nil {
		return fmt.Errorf("no replay connection")
	}

	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := req.WriteTo(conn); err != nil {
		_ = conn.SetWriteDeadline(time.Time{})
		return err
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err := resp.Read(reader); err != nil {
		_ = conn.SetReadDeadline(time.Time{})
		// The request is already out; assume delivered and recycle the
		// connection lazily on the next failure instead of re-posting.
		log.Printf("Replay response not read (assuming delivered): %v", err)
		return nil
	}
	_ = conn.SetReadDeadline(time.Time{})
	if code := resp.StatusCode(); code != fasthttp.StatusOK {
		log.Printf("Replay upload returned status %d", code)
	}
	return nil
}

// dropConn closes and clears the current connection so the next send reconnects.
func (m *UploadManager) dropConn() {
	m.connMu.Lock()
	defer m.connMu.Unlock()

	if m.conn != nil {
		_ = m.conn.Close()
		m.conn = nil
	}
	m.connReader = nil
}

func (m *UploadManager) pushReplayData(info *Info) {
	m.inflight.Add(1)

	defer func() {
		m.inflight.Add(-1)
		m.returnToPool(info)
	}()

	// Origin-form request line + Host header, serialized straight onto the
	// persistent connection with fasthttp (pooled, no per-request allocs).
	requestURI := fmt.Sprintf(
		"/game/replay/setReplayData?matchid=%d&sequenceid=%d",
		info.SessionID, info.SequenceID,
	)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetRequestURI(requestURI)
	req.Header.SetHost(m.host)
	// The upload is a raw POST: the replay host is told the method, the path,
	// the host and the length, and nothing else. A Content-Type would be an
	// extra header the contract does not have, and the serializer substitutes
	// one whenever it is missing, so the substitution is what gets disabled.
	req.Header.SetNoDefaultContentType(true)
	req.SetBody(info.Data)

	if err := m.sendRequest(req); err != nil {
		log.Printf("SockSend failed: %v", err)
		// Drop the dead connection and retry once on a fresh one. Replays are
		// idempotent (keyed by matchid+sequenceid), so re-posting the same data
		// after a reconnect is safe.
		m.dropConn()
		if err := m.sendRequest(req); err != nil {
			log.Printf("SockSend retry failed: %v", err)
		}
	}
}

func (m *UploadManager) getFromPool() *Info {
	m.poolMu.Lock()
	defer m.poolMu.Unlock()

	if len(m.pool) > 0 {
		info := m.pool[len(m.pool)-1]
		m.pool = m.pool[:len(m.pool)-1]
		return info
	}

	return &Info{}
}

func (m *UploadManager) returnToPool(info *Info) {
	m.poolMu.Lock()
	defer m.poolMu.Unlock()

	if len(m.pool) < 100 {
		info.Data = nil
		m.pool = append(m.pool, info)
	}
}

func (m *UploadManager) Stop() {
	m.connMu.Lock()
	defer m.connMu.Unlock()

	if m.conn != nil {
		_ = m.conn.Close()
		m.conn = nil
	}
	m.connReader = nil
}
