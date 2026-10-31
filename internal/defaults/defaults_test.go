package defaults

import (
	"testing"
)

func TestDefaultRelayBroadcastPort(t *testing.T) {
	checks := []struct {
		got  string
		want string
	}{
		{got: DefaultRelayBroadcastPort, want: "9999"},
		{got: SecondaryRelayBroadcastPort, want: "8888"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("expected default relay port %q, got %q", c.want, c.got)
		}
	}
}
