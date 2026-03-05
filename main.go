package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	listen := flag.Bool("listen", false, "Start in listener mode (TCP server + UDP discovery)")
	username := flag.String("u", "", "Username (required for listen and send)")
	message := flag.String("m", "", "Message to send to the target user")
	filePath := flag.String("t", "", "File path to send to the target user")
	showUsers := flag.Bool("users", false, "Discover and show online users, then exit")
	port := flag.String("port", "8080", "TCP port for the server (default 8080)")

	flag.Parse()

	// --- Mode: List users ---
	if *showUsers {
		discoverAndShowUsers()
		return
	}

	// --- Mode: Listen ---
	if *listen {
		if *username == "" {
			fmt.Println("Error: -u <username> is required in listen mode")
			flag.Usage()
			os.Exit(1)
		}
		runListenMode(*username, *port)
		return
	}

	// --- Mode: Send message or file ---
	if *username == "" {
		fmt.Println("Error: -u <target_username> is required")
		flag.Usage()
		os.Exit(1)
	}

	if *message == "" && *filePath == "" {
		fmt.Println("Error: specify -m <message> or -t <file> to send")
		flag.Usage()
		os.Exit(1)
	}

	// We need to discover the target user's IP
	peer := discoverPeer(*username)
	if peer == nil {
		fmt.Printf("Error: user '%s' not found on the network\n", *username)
		os.Exit(1)
	}

	// Determine our own username from hostname
	senderName := getSenderName()

	if *message != "" {
		if err := SendMessage(peer.IP, peer.Port, senderName, *message); err != nil {
			log.Fatalf("send message failed: %v", err)
		}
	}

	if *filePath != "" {
		if err := SendFile(peer.IP, peer.Port, senderName, *filePath); err != nil {
			log.Fatalf("send file failed: %v", err)
		}
	}
}

// runListenMode starts the TCP server and UDP discovery goroutines, then blocks.
func runListenMode(username, port string) {
	pt := NewPeerTable()

	fmt.Printf("[gochat] starting as '%s' on TCP port %s\n", username, port)

	go StartServer(port)
	go BroadcastPresence(username, port)
	go ListenDiscovery(pt, username)

	// Periodically clean stale peers
	go func() {
		for {
			time.Sleep(15 * time.Second)
			pt.Cleanup(30 * time.Second)
		}
	}()

	// Block until Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Println("\n[gochat] shutting down")
}

// discoverPeer runs a brief UDP discovery to find a specific user.
func discoverPeer(targetUsername string) *Peer {
	pt := NewPeerTable()

	// Start listening for discovery broadcasts
	go ListenDiscovery(pt, "_sender_") // use a dummy self-name

	fmt.Printf("[discovery] looking for '%s' on the network...\n", targetUsername)

	// Wait up to 5 seconds, checking every 500ms
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		peer := pt.Get(targetUsername)
		if peer != nil {
			fmt.Printf("[discovery] found %s @ %s:%s\n", targetUsername, peer.IP, peer.Port)
			return peer
		}
	}

	return nil
}

// discoverAndShowUsers discovers peers for a few seconds and prints them.
func discoverAndShowUsers() {
	pt := NewPeerTable()

	go ListenDiscovery(pt, "_lister_")

	fmt.Println("[discovery] scanning network for 3 seconds...")
	time.Sleep(3 * time.Second)

	pt.PrintAll()
}

// getSenderName returns a sender name derived from the hostname.
func getSenderName() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}
