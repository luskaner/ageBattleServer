package net

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type connAddr struct {
	conn *net.UDPConn
	port uint
}

type UDPBroadcaster struct {
	ports    []uint
	interval time.Duration
	conns    []connAddr
	mu       sync.Mutex
	closed   bool
	buildMsg func(port uint) [][]byte
}

func NewUDPBroadcaster(ports []uint, buildMsg func(port uint) [][]byte) *UDPBroadcaster {
	return &UDPBroadcaster{
		ports:    ports,
		interval: 10 * time.Second,
		buildMsg: buildMsg,
	}
}

func (b *UDPBroadcaster) Start(ctx context.Context) error {
	conns := make([]connAddr, 0, len(b.ports))
	for _, port := range b.ports {
		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   net.ParseIP("255.255.255.255"),
			Port: int(port),
		})
		if err != nil {
			// Don't leak the sockets created for earlier ports.
			for _, ca := range conns {
				_ = ca.conn.Close()
			}
			return fmt.Errorf("udp dial: %w", err)
		}
		if err := conn.SetWriteBuffer(1024); err != nil { // 1 KiB buffer, messages are small
			_ = conn.Close()
			for _, ca := range conns {
				_ = ca.conn.Close()
			}
			return err
		}
		conns = append(conns, connAddr{conn: conn, port: port})
	}

	b.mu.Lock()
	b.conns = conns
	b.mu.Unlock()

	go b.setupConns()
	go b.broadcastLoop(ctx)

	return nil
}

// setupConns dials one socket per broadcast address of every interface that
// is up and has the broadcast flag set, and replaces the fallback
// limited-broadcast sockets when ready. It runs in the background so Start
// stays fast even when interface enumeration is slow.
func (b *UDPBroadcaster) setupConns() {
	addrs, err := broadcastAddrs(b.ports)
	if err != nil {
		log.Printf("UDPBroadcaster: failed to enumerate broadcast addresses: %v", err)
		return
	}

	conns := make([]connAddr, 0, len(addrs))
	for _, addr := range addrs {
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Printf("UDPBroadcaster: dial %s: %v", addr, err)
			continue
		}
		if err := conn.SetWriteBuffer(1024); err != nil {
			_ = conn.Close()
			continue
		}
		conns = append(conns, connAddr{conn: conn, port: uint(addr.Port)})
	}

	if len(conns) == 0 {
		return
	}

	b.mu.Lock()
	if b.closed {
		for _, conn := range conns {
			_ = conn.conn.Close()
		}
		b.mu.Unlock()
		return
	}
	for _, conn := range b.conns {
		_ = conn.conn.Close()
	}
	b.conns = conns
	b.mu.Unlock()

	log.Printf("UDPBroadcaster using broadcast addresses: %v", addrs)
}

// broadcastAddrs returns the IPv4 broadcast address of every interface that
// is up and has the broadcast flag set, crossed with every destination port.
// If no interface qualifies it falls back to the limited broadcast
// 255.255.255.255 so the sockets still work.
func broadcastAddrs(ports []uint) ([]*net.UDPAddr, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	bcastIPs := make([]net.IP, 0, len(ifaces))
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 {
			continue
		}
		ifaddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, ifaddr := range ifaddrs {
			ipnet, ok := ifaddr.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			ip := ipnet.IP.To4()
			mask := ipnet.Mask
			bcast := make(net.IP, len(ip))
			for i := range ip {
				bcast[i] = ip[i] | ^mask[i]
			}
			bcastIPs = append(bcastIPs, bcast)
		}
	}

	if len(bcastIPs) == 0 {
		bcastIPs = []net.IP{net.ParseIP("255.255.255.255")}
	}

	addrs := make([]*net.UDPAddr, 0, len(bcastIPs)*len(ports))
	for _, ip := range bcastIPs {
		for _, port := range ports {
			addrs = append(addrs, &net.UDPAddr{IP: ip, Port: int(port)})
		}
	}
	return addrs, nil
}

func (b *UDPBroadcaster) broadcastLoop(ctx context.Context) {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.broadcast()
		}
	}
}

func (b *UDPBroadcaster) broadcast() {
	if b.buildMsg == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.conns != nil && !b.closed {
		for _, ca := range b.conns {
			for _, data := range b.buildMsg(ca.port) {
				if data == nil {
					continue
				}
				_, _ = ca.conn.Write(data)
			}
		}
	}
}

func (b *UDPBroadcaster) Stop() {
	b.mu.Lock()
	b.closed = true
	for _, ca := range b.conns {
		_ = ca.conn.Close()
	}
	b.conns = nil
	b.mu.Unlock()
}
