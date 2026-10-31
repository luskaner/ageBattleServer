package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	l := New()
	if l == nil {
		t.Fatal("New() returned nil")
	}
	if l.GetGlobalLevel() != LevelDebug {
		t.Errorf("expected default level %d, got %d", LevelDebug, l.GetGlobalLevel())
	}
}

func TestSetGlobalLevel(t *testing.T) {
	l := New()
	l.SetGlobalLevel(LevelWarn)
	if l.GetGlobalLevel() != LevelWarn {
		t.Errorf("expected level %d, got %d", LevelWarn, l.GetGlobalLevel())
	}
}

func TestGlobalLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetGlobalLevel(LevelWarn)

	l.Debug("Test", "method", "debug message")
	l.Info("Test", "method", "info message")
	l.Warning("Test", "method", "warn message")
	l.Error("Test", "method", "error message")
	l.Flush()

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered at Warn level")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered at Warn level")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warning message should NOT be filtered at Warn level")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should NOT be filtered at Warn level")
	}
}

func TestModuleLevelOverride(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetGlobalLevel(LevelError)
	l.SetModule("Listener", LevelDebug)

	l.Debug("Listener", "connect", "listener debug")
	l.Debug("Server", "Start", "server debug")
	l.Flush()

	output := buf.String()
	if !strings.Contains(output, "listener debug") {
		t.Error("Listener debug should be logged (module override)")
	}
	if strings.Contains(output, "server debug") {
		t.Error("Server debug should be filtered (global level)")
	}
}

func TestParseModuleLevel(t *testing.T) {
	tests := []struct {
		input     string
		wantName  string
		wantLevel Level
		wantError bool
	}{
		{"Listener:debug", "Listener", LevelDebug, false},
		{"Server:info", "Server", LevelInfo, false},
		{"Client:none", "Client", LevelNone, false},
		{"invalid", "", 0, true},
		{"name:abc", "", 0, true},
		{"name:invalid", "", 0, true},
	}

	for _, tt := range tests {
		name, level, err := ParseModuleLevel(tt.input)
		if tt.wantError {
			if err == nil {
				t.Errorf("ParseModuleLevel(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseModuleLevel(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if name != tt.wantName {
			t.Errorf("ParseModuleLevel(%q) name = %q, want %q", tt.input, name, tt.wantName)
		}
		if level != tt.wantLevel {
			t.Errorf("ParseModuleLevel(%q) level = %d, want %d", tt.input, level, tt.wantLevel)
		}
	}
}

func TestFormatOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetPrefix("test")

	l.Info("Server", "Start", "test message %d", 42)
	l.Flush()

	output := buf.String()
	if !strings.Contains(output, "[Info]") {
		t.Errorf("Expected [Info] in output, got: %s", output)
	}
	if !strings.Contains(output, "Server::Start") {
		t.Errorf("Expected Server::Start in output, got: %s", output)
	}
	if !strings.Contains(output, "test message 42") {
		t.Errorf("Expected formatted message in output, got: %s", output)
	}
}

func TestShouldLog(t *testing.T) {
	l := New()
	l.SetGlobalLevel(LevelInfo)

	if !l.ShouldLog(LevelError) {
		t.Error("Error should be logged at Info level")
	}
	if !l.ShouldLog(LevelInfo) {
		t.Error("Info should be logged at Info level")
	}
	if l.ShouldLog(LevelDebug) {
		t.Error("Debug should NOT be logged at Info level")
	}
}

func TestLevelClamping(t *testing.T) {
	l := New()
	l.SetGlobalLevel(LevelDebug)

	if l.GetGlobalLevel() != LevelDebug {
		t.Errorf("Level should be Debug (5), got %d", l.GetGlobalLevel())
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input     string
		wantLevel Level
		wantError bool
	}{
		{"none", LevelNone, false},
		{"error", LevelError, false},
		{"warn", LevelWarn, false},
		{"warning", LevelWarn, false},
		{"notice", LevelNotice, false},
		{"info", LevelInfo, false},
		{"debug", LevelDebug, false},
		{"DEBUG", LevelDebug, false},
		{"Debug", LevelDebug, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		level, err := ParseLevel(tt.input)
		if tt.wantError {
			if err == nil {
				t.Errorf("ParseLevel(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseLevel(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if level != tt.wantLevel {
			t.Errorf("ParseLevel(%q) level = %d, want %d", tt.input, level, tt.wantLevel)
		}
	}
}

func TestGetPrefix(t *testing.T) {
	l := New()
	if l.GetPrefix() != "bs" {
		t.Errorf("default prefix = %q, want bs", l.GetPrefix())
	}
	l.SetPrefix("custom")
	if l.GetPrefix() != "custom" {
		t.Errorf("prefix after SetPrefix = %q, want custom", l.GetPrefix())
	}
}

func TestLogLevelNotice(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.Log("Server", "Tick", LevelNotice, "notice message")
	l.Flush()
	if !strings.Contains(buf.String(), "[Notice]") {
		t.Errorf("expected [Notice] in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "Server::Tick") {
		t.Errorf("expected Server::Tick in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "notice message") {
		t.Errorf("expected message in output, got: %s", buf.String())
	}
}

func TestLogUnknownLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetGlobalLevel(Level(99))
	l.Log("Server", "Tick", Level(99), "unknown level message")
	l.Flush()
	if !strings.Contains(buf.String(), "[Unknown]") {
		t.Errorf("expected [Unknown] in output, got: %s", buf.String())
	}
}

func TestLogWithoutMethod(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.Info("Server", "", "no method message")
	l.Flush()
	if !strings.Contains(buf.String(), "Server: no method message") {
		t.Errorf("expected bare module prefix, got: %s", buf.String())
	}
}

func TestLogFiltered(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetGlobalLevel(LevelError)
	l.Log("Server", "Tick", LevelInfo, "filtered")
	if buf.Len() != 0 {
		t.Errorf("expected no output for filtered level, got: %s", buf.String())
	}
}

func TestModulePrefixMatchReverse(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetGlobalLevel(LevelError)
	l.SetModule("S", LevelDebug)

	l.Debug("Server", "Start", "server debug")
	l.Flush()

	if !strings.Contains(buf.String(), "server debug") {
		t.Error("module prefix should match when module name is shorter than the set module")
	}
}

func TestModuleCheckBypassedWhenModuleEmpty(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetGlobalLevel(LevelError)
	l.SetModule("Listener", LevelError)

	l.Debug("", "Start", "empty module debug")

	if buf.Len() != 0 {
		t.Errorf("empty module should fall back to global level and be filtered, got: %s", buf.String())
	}
}

func TestModuleLevelBlocks(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetGlobalLevel(LevelDebug)
	l.SetModule("Server", LevelError)

	l.Debug("Server", "Start", "blocked by module level")

	if buf.Len() != 0 {
		t.Errorf("expected no output when module level filters, got: %s", buf.String())
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	defer SetGlobalLevel(LevelDebug)
	defer SetModule("", 0)
	defer SetOutput(os.Stderr)

	var buf bytes.Buffer
	SetOutput(&buf)
	SetGlobalLevel(LevelInfo)
	if GetGlobalLevel() != LevelInfo {
		t.Errorf("GetGlobalLevel() = %d, want Info", GetGlobalLevel())
	}
	if !ShouldLog(LevelError) {
		t.Error("ShouldLog(Error) should be true at Info level")
	}
	if ShouldLog(LevelDebug) {
		t.Error("ShouldLog(Debug) should be false at Info level")
	}

	SetModule("Server", LevelDebug)
	Error("Server", "E", "pkg error")
	Warning("Server", "W", "pkg warning")
	Info("Server", "I", "pkg info")
	Debug("Server", "D", "pkg debug")
	Log("Server", "L", LevelNotice, "pkg notice")
	Flush()

	out := buf.String()
	for _, want := range []string{"pkg error", "pkg warning", "pkg info", "pkg debug", "pkg notice"} {
		if !strings.Contains(out, want) {
			t.Errorf("package-level output missing %q, got: %s", want, out)
		}
	}
	if !strings.Contains(out, "[Error]") || !strings.Contains(out, "[Debug]") {
		t.Errorf("expected level names in output, got: %s", out)
	}
}

func BenchmarkShouldLog(b *testing.B) {
	l := New()
	l.SetGlobalLevel(LevelInfo)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if l.ShouldLog(LevelDebug) {
			b.Fatal("debug should be filtered at info level")
		}
	}
}

func BenchmarkDebugFiltered(b *testing.B) {
	var sink bytes.Buffer
	l := New()
	l.SetOutput(&sink)
	l.SetGlobalLevel(LevelError)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("Client", "handleMessage", "filtered message")
	}
}

func BenchmarkDebugLogged(b *testing.B) {
	var sink bytes.Buffer
	l := New()
	l.SetOutput(&sink)
	defer l.Flush()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("Client", "handleMessage", "type=0x09 length=%d body=%x from=%s", 12, []byte{0x01, 0x02, 0x03}, "127.0.0.1:1")
	}
}
