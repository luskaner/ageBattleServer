package server

import (
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// TestConnectPeerRetriesOnAnErasedSession hands the connect path a session that
// was erased after the lookup (its last member closed in between). The peer must
// land in the live session under the same key, where the next connect finds it.
func TestConnectPeerRetriesOnAnErasedSession(t *testing.T) {
	srv := createTestServer(t)
	key := session.SessionKey{TitleID: 1, SessionID: 77}
	stale := srv.sessionManager.GetOrCreateSession(key)
	last := NewClient(&mockNetConn{}, srv)
	if err := srv.sessionManager.AddPeer(stale, stale.NewPeer(stale.NextPeerID(), last)); err != nil {
		t.Fatal(err)
	}
	srv.sessionManager.HandleDisconnect(last, nil)

	joiner := NewClient(&mockNetConn{}, srv)
	sess, ok := srv.connectPeer(stale, key, joiner, 0x1122)
	if !ok {
		t.Fatal("connectPeer refused the joiner")
	}
	if live := srv.sessionManager.GetSession(key); sess != live || live == nil {
		t.Fatal("the joiner is not in the session the map holds under its key")
	}
	if sess.PeerIDForConn(joiner) == 0 {
		t.Fatal("the joiner has no peer in the live session")
	}
}

// TestReconnectOnTheConnectionThatHoldsThePeerIsIgnored sends a reconnect on the
// connection that already holds the peer. The original drops a second 0x15 on a
// joined connection; here it must neither close the connection nor remove the
// player.
func TestReconnectOnTheConnectionThatHoldsThePeerIsIgnored(t *testing.T) {
	srv := createTestServer(t)
	client := NewClient(&mockNetConn{}, srv)
	client.handleMessage(protocol.MsgLobbyConnect, lobbyConnectWithOverride(0, 0))
	client.handleMessage(protocol.MsgLobbyConnect, lobbyConnectWithOverride(1, 0))
	if clientClosed(client) {
		t.Fatal("the connection that holds the peer was closed")
	}
	sess := srv.sessionManager.FindSessionByPeer(client)
	if sess == nil || sess.PeerIDForConn(client) == 0 {
		t.Fatal("the player was removed from its session")
	}
}
