package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	// DiscoveryPort is the UDP port used for peer discovery.
	DiscoveryPort = 9999
	// MulticastGroup is the multicast address all peers join.
	// 239.x.x.x is the "administratively scoped" range, safe for LAN use.
	MulticastGroup = "239.0.0.1"
	// BroadcastInterval is how often presence is advertised.
	BroadcastInterval = 5 * time.Second
	// DiscoveryPrefix is the protocol prefix for discovery packets.
	DiscoveryPrefix = "DISCOVER:"
)

// BroadcastPresence sends a UDP multicast packet every BroadcastInterval
// advertising this peer's username and TCP port.
func BroadcastPresence(username, tcpPort string) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", MulticastGroup, DiscoveryPort))
	if err != nil {
		log.Fatalf("resolve multicast addr: %v", err)
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		log.Fatalf("dial multicast: %v", err)
	}
	defer conn.Close()

	msg := fmt.Sprintf("%s%s:%s", DiscoveryPrefix, username, tcpPort)

	// Initial burst — send 3 times at 500ms intervals for fast discovery
	for i := 0; i < 3; i++ {
		conn.Write([]byte(msg))
		time.Sleep(500 * time.Millisecond)
	}

	// Then send at regular interval
	for {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Printf("multicast send error: %v", err)
		}
		time.Sleep(BroadcastInterval)
	}
}

// ListenDiscovery listens for UDP multicast discovery packets and updates
// the peer table. Uses net.ListenMulticastUDP which is cross-platform and
// handles SO_REUSEADDR internally.
func ListenDiscovery(pt *PeerTable, selfUsername string) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", MulticastGroup, DiscoveryPort))
	if err != nil {
		log.Fatalf("resolve multicast addr: %v", err)
	}

	// nil interface = join on all available interfaces
	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Fatalf("listen discovery: %v", err)
	}
	defer conn.Close()

	// Set a generous read buffer
	conn.SetReadBuffer(1024)

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
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

		peerIP := remoteAddr.IP.String()

		// Only log the first time we see this peer
		existing := pt.Get(peerUser)
		if existing == nil {
			fmt.Printf("[discovery] found peer: %s @ %s:%s\n", peerUser, peerIP, peerPort)
		}

		pt.Set(peerUser, peerIP, peerPort)
	}
}
