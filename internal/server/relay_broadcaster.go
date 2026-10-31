package server

import (
	"context"

	"github.com/luskaner/ageBattleServer/internal/logger"
	battlenet "github.com/luskaner/ageBattleServer/internal/net"
)

type RelayBroadcaster struct {
	broadcast func(port uint) [][]byte
	ports     []uint
	guid      string
	listener  *battlenet.UDPBroadcaster
}

func NewRelayBroadcaster(broadcast func(port uint) [][]byte, ports []uint) *RelayBroadcaster {
	return &RelayBroadcaster{
		broadcast: broadcast,
		ports:     ports,
		guid:      battlenet.GenerateGUID(),
	}
}

func (b *RelayBroadcaster) GUID() string {
	return b.guid
}

func (b *RelayBroadcaster) Start(ctx context.Context) error {
	b.listener = battlenet.NewUDPBroadcaster(b.ports, b.broadcast)

	if err := b.listener.Start(ctx); err != nil {
		return err
	}

	logger.Info("Listener", "broadcast", "relayBroadcastPort broadcasting to ports %v", b.ports)
	return nil
}

func (b *RelayBroadcaster) Stop() {
	if b.listener != nil {
		b.listener.Stop()
	}
}
