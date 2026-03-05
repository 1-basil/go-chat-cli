package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"syscall"
	"time"
)

const (
	// DiscoveryPort is the UDP port used for peer discovery.
	DiscoveryPort = 9999
	// BroadcastInterval is how often presence is advertised.
	BroadcastInterval = 5 * time.Second
	// DiscoveryPrefix is the protocol prefix for discovery packets.
	DiscoveryPrefix = "DISCOVER:"
)

// BroadcastPresence sends a UDP broadcast every BroadcastInterval advertising
// this peer's username and TCP port. It sends an initial burst of 3 packets
// for faster discovery on startup.
func BroadcastPresence(username, tcpPort string) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", DiscoveryPort))
	if err != nil {
		log.Fatalf("resolve broadcast addr: %v", err)
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		log.Fatalf("dial broadcast: %v", err)
	}
	defer conn.Close()

	msg := fmt.Sprintf("%s%s:%s", DiscoveryPrefix, username, tcpPort)

	// Initial burst — send 3 times at 500ms intervals for fast discovery
	for i := 0; i < 3; i++ {
		conn.Write([]byte(msg))
		time.Sleep(500 * time.Millisecond)
	}

	// Then broadcast at regular interval
	for {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Printf("broadcast error: %v", err)
		}
		time.Sleep(BroadcastInterval)
	}
}

// ListenDiscovery listens for UDP discovery broadcasts and updates the peer table.
// It uses SO_REUSEADDR so multiple processes on the same machine can share the port.
func ListenDiscovery(pt *PeerTable, selfUsername string) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			})
		},
	}

	pc, err := lc.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		log.Fatalf("listen discovery: %v", err)
	}
	defer pc.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("discovery read error: %v", err)
			continue
		}

		raw := string(buf[:n])
		if !strings.HasPrefix(raw, DiscoveryPrefix) {
			continue
		}

		// Parse "DISCOVER:<username>:<port>"
		payload := strings.TrimPrefix(raw, DiscoveryPrefix)
		parts := strings.SplitN(payload, ":", 2)
		if len(parts) != 2 {
			continue
		}

		peerUser := parts[0]
		peerPort := parts[1]

		// Ignore our own broadcasts
		if peerUser == selfUsername {
			continue
		}

		// Extract IP from sender address
		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			continue
		}
		peerIP := udpAddr.IP.String()

		// Only log the first time we see this peer
		existing := pt.Get(peerUser)
		if existing == nil {
			fmt.Printf("[discovery] found peer: %s @ %s:%s\n", peerUser, peerIP, peerPort)
		}

		pt.Set(peerUser, peerIP, peerPort)
	}
}
