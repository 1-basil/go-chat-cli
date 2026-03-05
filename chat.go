package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

// RunChat starts an interactive chat session.
// It launches a TCP server to receive messages and reads stdin to send them.
func RunChat(username, peerIP, peerPort, listenPort string) {
	fmt.Printf("[gochat] chat started as '%s' → %s:%s\n", username, peerIP, peerPort)
	fmt.Println("[gochat] type messages and press Enter to send. Ctrl+C to quit.")
	fmt.Println()

	// Start TCP server to receive messages
	go StartServer(listenPort)

	// Start UDP discovery in background (so others can find us)
	pt := NewPeerTable()
	go BroadcastPresence(username, listenPort)
	go ListenDiscovery(pt, username)

	// Handle Ctrl+C gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		fmt.Println("\n[gochat] chat ended")
		os.Exit(0)
	}()

	// Read stdin and send messages
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		err := SendMessage(peerIP, peerPort, username, line)
		if err != nil {
			fmt.Printf("[error] %v\n", err)
		}
	}
}
