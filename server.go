package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// StartServer starts the TCP server on the given port to receive messages and files.
func StartServer(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("tcp listen error: %v", err)
	}
	fmt.Printf("[server] listening on TCP port %s\n", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("tcp accept error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read the header line: "MSG|sender|body" or "FILE|sender|filename|size"
	headerLine, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("read header error: %v", err)
		return
	}
	headerLine = strings.TrimSpace(headerLine)
	parts := strings.SplitN(headerLine, "|", 4)

	if len(parts) < 2 {
		log.Printf("invalid header: %s", headerLine)
		return
	}

	switch parts[0] {
	case "MSG":
		handleMessage(parts)
	case "FILE":
		handleFile(parts, reader)
	default:
		log.Printf("unknown type: %s", parts[0])
	}
}

func handleMessage(parts []string) {
	if len(parts) < 3 {
		log.Printf("malformed MSG header")
		return
	}
	sender := parts[1]
	body := parts[2]
	fmt.Printf("\n[message from %s]: %s\n", sender, body)
}

func handleFile(parts []string, reader *bufio.Reader) {
	if len(parts) < 4 {
		log.Printf("malformed FILE header")
		return
	}
	sender := parts[1]
	filename := filepath.Base(parts[2]) // sanitize
	sizeStr := parts[3]

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		log.Printf("invalid file size: %v", err)
		return
	}

	fmt.Printf("\n[file from %s]: receiving '%s' (%d bytes)...\n", sender, filename, size)

	// Save to current directory
	outFile, err := os.Create(filename)
	if err != nil {
		log.Printf("create file error: %v", err)
		return
	}
	defer outFile.Close()

	written, err := io.CopyN(outFile, reader, size)
	if err != nil {
		log.Printf("file receive error after %d bytes: %v", written, err)
		return
	}

	fmt.Printf("[file from %s]: saved '%s' (%d bytes)\n", sender, filename, written)
}
