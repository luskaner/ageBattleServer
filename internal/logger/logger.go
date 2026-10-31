package logger

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Level int

const (
	LevelNone   Level = 0
	LevelError  Level = 1
	LevelWarn   Level = 2
	LevelNotice Level = 3
	LevelInfo   Level = 4
	LevelDebug  Level = 5
)

var levelNames = map[Level]string{
	LevelNone:   "None",
	LevelError:  "Error",
	LevelWarn:   "Warning",
	LevelNotice: "Notice",
	LevelInfo:   "Info",
	LevelDebug:  "Debug",
}

type Module struct {
	Name  string
	Level Level
}

// flushInterval is how often the lazy flusher drains the output buffer. The
// bufio.Writer also flushes automatically once its buffer fills, so this only
// bounds the latency of partially-filled batches under light logging.
const flushInterval = 100 * time.Millisecond

// Logger writes timestamped log lines. Levels are read with atomics so the hot
// path (shouldLog, used as a per-message guard by the message dispatchers) never
// takes a lock; the output buffer and its flusher are serialized under mu.
type Logger struct {
	mu          sync.RWMutex
	globalLevel atomic.Int32
	moduleName  atomic.Pointer[string]
	moduleLevel atomic.Int32
	output      io.Writer
	buf         *bufio.Writer
	flusher     bool // true while the lazy flush goroutine is running
	prefix      string
	pidStr      string
}

func New() *Logger {
	pid := os.Getpid()
	l := &Logger{
		output: os.Stderr,
		prefix: "bs",
		pidStr: strconv.Itoa(pid),
	}
	l.globalLevel.Store(int32(LevelDebug))
	l.moduleLevel.Store(int32(LevelDebug))
	return l
}

func (l *Logger) SetGlobalLevel(level Level) {
	l.globalLevel.Store(int32(level))
}

func (l *Logger) GetGlobalLevel() Level {
	return Level(l.globalLevel.Load())
}

func (l *Logger) SetModule(name string, level Level) {
	l.moduleLevel.Store(int32(level))
	if name == "" {
		l.moduleName.Store(nil)
	} else {
		l.moduleName.Store(&name)
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.buf != nil {
		_ = l.buf.Flush()
	}
	l.output = w
	l.buf = nil
}

func (l *Logger) SetPrefix(p string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = p
}

func (l *Logger) GetPrefix() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.prefix
}

// shouldLog reports whether a message at level for module passes the global and
// module level filters. It is lock-free and safe for the hot path.
func (l *Logger) shouldLog(module string, level Level) bool {
	if module != "" {
		if mn := l.moduleName.Load(); mn != nil {
			name := *mn
			if strings.HasPrefix(module, name) || strings.HasPrefix(name, module) {
				return level <= Level(l.moduleLevel.Load())
			}
		}
	}
	return level <= Level(l.globalLevel.Load())
}

func (l *Logger) log(level Level, module, method, format string, args ...any) {
	if !l.shouldLog(module, level) {
		return
	}

	levelName, ok := levelNames[level]
	if !ok {
		levelName = "Unknown"
	}

	now := time.Now()
	ms := now.Nanosecond() / 1000000
	timestamp := fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d:%03d",
		now.Year(), int(now.Month()), now.Day(),
		now.Hour(), now.Minute(), now.Second(), ms)

	msg := fmt.Sprintf(format, args...)

	var moduleStr string
	if method != "" {
		moduleStr = module + "::" + method
	} else {
		moduleStr = module
	}

	var sb strings.Builder
	sb.WriteString(timestamp)
	sb.WriteString(" #")
	sb.WriteString(l.pidStr)
	sb.WriteString(" [")
	sb.WriteString(levelName)
	sb.WriteString("] ")
	sb.WriteString(moduleStr)
	sb.WriteString(": ")
	sb.WriteString(msg)
	sb.WriteString("\n")

	l.mu.Lock()
	if l.buf == nil {
		l.buf = bufio.NewWriter(l.output)
	}
	_, _ = l.buf.WriteString(sb.String())
	if !l.flusher {
		l.flusher = true
		go l.flushLoop()
	}
	l.mu.Unlock()
}

// flushLoop drains the output buffer on a timer while there is anything to
// write and exits once it is empty, so the goroutine only lives while the
// logger is active. It is spawned lazily by log; a write that happens after the
// loop's idle check sees flusher == false and starts a fresh loop.
func (l *Logger) flushLoop() {
	for {
		time.Sleep(flushInterval)
		l.mu.Lock()
		if l.buf != nil && l.buf.Buffered() > 0 {
			_ = l.buf.Flush()
		}
		if l.buf == nil || l.buf.Buffered() == 0 {
			l.flusher = false
			l.mu.Unlock()
			return
		}
		l.mu.Unlock()
	}
}

// Flush synchronously drains the output buffer. Callers that read the writer
// directly (tests) must call it before inspecting the output.
func (l *Logger) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.buf != nil {
		_ = l.buf.Flush()
	}
}

func (l *Logger) Error(module, method, format string, args ...any) {
	l.log(LevelError, module, method, format, args...)
}

func (l *Logger) Warning(module, method, format string, args ...any) {
	l.log(LevelWarn, module, method, format, args...)
}

func (l *Logger) Info(module, method, format string, args ...any) {
	l.log(LevelInfo, module, method, format, args...)
}

func (l *Logger) Debug(module, method, format string, args ...any) {
	l.log(LevelDebug, module, method, format, args...)
}

func (l *Logger) Log(module, method string, level Level, format string, args ...any) {
	l.log(level, module, method, format, args...)
}

func (l *Logger) ShouldLog(level Level) bool {
	return l.shouldLog("", level)
}

var defaultLogger = New()

func SetGlobalLevel(level Level) {
	defaultLogger.SetGlobalLevel(level)
}

func GetGlobalLevel() Level {
	return defaultLogger.GetGlobalLevel()
}

func SetModule(name string, level Level) {
	defaultLogger.SetModule(name, level)
}

func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

func Flush() {
	defaultLogger.Flush()
}

func Error(module, method, format string, args ...any) {
	defaultLogger.Error(module, method, format, args...)
}

func Warning(module, method, format string, args ...any) {
	defaultLogger.Warning(module, method, format, args...)
}

func Info(module, method, format string, args ...any) {
	defaultLogger.Info(module, method, format, args...)
}

func Debug(module, method, format string, args ...any) {
	defaultLogger.Debug(module, method, format, args...)
}

func Log(module, method string, level Level, format string, args ...any) {
	defaultLogger.Log(module, method, level, format, args...)
}

func ShouldLog(level Level) bool {
	return defaultLogger.shouldLog("", level)
}

func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "none":
		return LevelNone, nil
	case "error":
		return LevelError, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "notice":
		return LevelNotice, nil
	case "info":
		return LevelInfo, nil
	case "debug":
		return LevelDebug, nil
	default:
		return LevelNone, fmt.Errorf("invalid log level: %s (valid: none, error, warning, notice, info, debug)", s)
	}
}

func ParseModuleLevel(s string) (string, Level, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid module level format: %s (expected Module:Level)", s)
	}

	name := strings.TrimSpace(parts[0])
	levelStr := strings.TrimSpace(parts[1])

	level, err := ParseLevel(levelStr)
	if err != nil {
		return "", 0, err
	}

	return name, level, nil
}
