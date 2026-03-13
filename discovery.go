package main

import (
	"context"
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

// getBroadcastAddrs returns all broadcast addresses for this machine's
// network interfaces, plus the global broadcast 255.255.255.255.
func getBroadcastAddrs() []string {
	addrs := []string{"255.255.255.255", "127.0.0.1"}

	ifaces, err := net.Interfaces()
	if err != nil {
		return addrs
	}

	for _, iface := range ifaces {
		// Skip down or non-broadcast interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 {
			continue
		}

		ifAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, a := range ifAddrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil {
				continue
			}

			// Calculate directed broadcast: IP | ~Mask
			mask := ipnet.Mask
			if len(mask) == 16 {
				mask = mask[12:]
			}
			if len(mask) != 4 {
				continue
			}

			broadcast := make(net.IP, 4)
			for i := 0; i < 4; i++ {
				broadcast[i] = ip4[i] | ^mask[i]
			}
			bcast := broadcast.String()
			if bcast != "255.255.255.255" {
				addrs = append(addrs, bcast)
			}
		}
	}

	return addrs
}

// BroadcastPresence sends a UDP broadcast every BroadcastInterval advertising
// this peer's username and TCP port. Sends to all broadcast addresses for
// maximum reach.
func BroadcastPresence(username, tcpPort string) {
	bcastAddrs := getBroadcastAddrs()
	msg := []byte(fmt.Sprintf("%s%s:%s", DiscoveryPrefix, username, tcpPort))

	// Resolve all broadcast targets
	targets := make([]*net.UDPAddr, 0, len(bcastAddrs))
	for _, addr := range bcastAddrs {
		ua, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", addr, DiscoveryPort))
		if err == nil {
			targets = append(targets, ua)
		}
	}

	if len(targets) == 0 {
		log.Fatalf("no broadcast addresses found")
	}

	fmt.Printf("[discovery] broadcasting to %d address(es)\n", len(targets))

	// Use reuseListenConfig so we get SO_BROADCAST permission
	lc := reuseListenConfig()
	conn, err := lc.ListenPacket(context.Background(), "udp4", ":0")
	if err != nil {
		log.Fatalf("broadcast socket: %v", err)
	}
	defer conn.Close()

	sendAll := func() {
		for _, t := range targets {
			_, err := conn.WriteTo(msg, t)
			if err != nil {
				// Don't fatal, just log. Some interfaces might fail while others work.
				log.Printf("[discovery] broadcast to %s failed: %v", t.String(), err)
			}
		}
	}

	// Initial burst — send 5 times at 500ms intervals for fast discovery
	for i := 0; i < 5; i++ {
		sendAll()
		time.Sleep(500 * time.Millisecond)
	}

	// Then send at regular interval
	for {
		sendAll()
		time.Sleep(BroadcastInterval)
	}
}

// ListenDiscovery listens for UDP broadcast discovery packets and updates
// the peer table. Uses SO_REUSEADDR (via reuseListenConfig) so multiple
// processes on the same machine can share the port.
func ListenDiscovery(pt *PeerTable, selfUsername string) {
	lc := reuseListenConfig()
	pc, err := lc.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		fmt.Printf("[discovery] FATAL: could not start discovery listener on port %d: %v\n", DiscoveryPort, err)
		return
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
