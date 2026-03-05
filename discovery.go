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
	// BroadcastInterval is how often presence is advertised.
	BroadcastInterval = 5 * time.Second
	// DiscoveryPrefix is the protocol prefix for discovery packets.
	DiscoveryPrefix = "DISCOVER:"
)

// BroadcastPresence sends a UDP broadcast every BroadcastInterval advertising
// this peer's username and TCP port.
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

	for {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Printf("broadcast error: %v", err)
		}
		time.Sleep(BroadcastInterval)
	}
}

// ListenDiscovery listens for UDP discovery broadcasts and updates the peer table.
// It ignores packets from our own username.
func ListenDiscovery(pt *PeerTable, selfUsername string) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		log.Fatalf("resolve listen addr: %v", err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatalf("listen discovery: %v", err)
	}
	defer conn.Close()

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
