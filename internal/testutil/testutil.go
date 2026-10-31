package testutil

import (
	"testing"
	"time"
)

// CheckDeadlines verifies that the three deadline setters accept a value.
func CheckDeadlines(t *testing.T, setDeadline, setReadDeadline, setWriteDeadline func(time.Time) error) {
	t.Helper()
	now := time.Now()
	if err := setDeadline(now); err != nil {
		t.Errorf("SetDeadline error: %v", err)
	}
	if err := setReadDeadline(now); err != nil {
		t.Errorf("SetReadDeadline error: %v", err)
	}
	if err := setWriteDeadline(now); err != nil {
		t.Errorf("SetWriteDeadline error: %v", err)
	}
}
