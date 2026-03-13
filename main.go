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
	chat := flag.Bool("chat", false, "Start interactive chat mode")
	username := flag.String("u", "", "Username (required for listen and send)")
	message := flag.String("m", "", "Message to send to the target user")
	filePath := flag.String("t", "", "File path to send to the target user")
	showUsers := flag.Bool("users", false, "Discover and show online users, then exit")
	targetIP := flag.String("ip", "", "Direct IP address (skip discovery, use with -u)")
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

	// --- Mode: Interactive chat ---
	if *chat {
		if *username == "" {
			fmt.Println("Error: -u <username> is required in chat mode")
			flag.Usage()
			os.Exit(1)
		}
		if *targetIP == "" {
			fmt.Println("Error: -ip <address> is required in chat mode")
			flag.Usage()
			os.Exit(1)
		}
		RunChat(*username, *targetIP, *port, *port)
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

	var peerIP, peerPort string

	if *targetIP != "" {
		// Direct IP mode — skip discovery
		peerIP = *targetIP
		peerPort = *port
		fmt.Printf("[direct] connecting to %s:%s\n", peerIP, peerPort)
	} else {
		// Discovery mode
		peer := discoverPeer(*username)
		if peer == nil {
			fmt.Printf("Error: user '%s' not found on the network\n", *username)
			fmt.Println("Tip: use -ip <address> to connect directly")
			os.Exit(1)
		}
		peerIP = peer.IP
		peerPort = peer.Port
	}

	// Determine our own username from hostname
	senderName := getSenderName()

	if *message != "" {
		if err := SendMessage(peerIP, peerPort, senderName, *message); err != nil {
			log.Fatalf("send message failed: %v", err)
		}
	}

	if *filePath != "" {
		if err := SendFile(peerIP, peerPort, senderName, *filePath); err != nil {
			log.Fatalf("send file failed: %v", err)
		}
	}
}

// runListenMode starts the TCP server and UDP discovery goroutines, then blocks.
func runListenMode(username, port string) {
	pt := NewPeerTable()

	fmt.Println(`
  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  ┃                                                       ┃
  ┃    ░██████╗░░█████╗░░█████╗░██╗░░██╗░█████╗░████████╗ ┃
  ┃    ██╔════╝░██╔══██╗██╔══██╗██║░░██║██╔══██╗╚══██╔══╝ ┃
  ┃    ██║░░███╗██║░░██║██║░░╚═╝███████║███████║░░░██║░░░ ┃
  ┃    ██║░░░██║██║░░██║██║░░██╗██╔══██║██╔══██║░░░██║░░░ ┃
  ┃    ╚██████╔╝╚█████╔╝╚█████╔╝██║░░██║██║░░██║░░░██║░░░ ┃
  ┃    ░╚═════╝░░╚════╝░░╚════╝░╚═╝░░╚═╝╚═╝░░╚═╝░░░╚═╝░░░ ┃
  ┃                 ░█████╗░██╗░░░░░██╗                    ┃
  ┃                 ██╔══██╗██║░░░░░██║                    ┃
  ┃                 ██║░░╚═╝██║░░░░░██║                    ┃
  ┃                 ██║░░██╗██║░░░░░██║                    ┃
  ┃                 ╚█████╔╝███████╗██║                    ┃
  ┃                 ░╚════╝░╚══════╝╚═╝                    ┃
  ┃─────────────────────────────────────────────────────── ┃
  ┃       ⚡ P2P Chat & File Transfer on your LAN ⚡       ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
	`)

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

	fmt.Println("[discovery] scanning network for 5 seconds...")
	time.Sleep(5 * time.Second)
	fmt.Println("[discovery] scan complete.")

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
