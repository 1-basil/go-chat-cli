package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

// SendMessage sends a text message to a peer via TCP.
func SendMessage(ip, port, sender, message string) error {
	addr := net.JoinHostPort(ip, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer conn.Close()

	header := fmt.Sprintf("MSG|%s|%s\n", sender, message)
	_, err = conn.Write([]byte(header))
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	fmt.Printf("[sent] message to %s\n", addr)
	return nil
}

// SendFile sends a file to a peer via TCP.
func SendFile(ip, port, sender, filePath string) error {
	// Open and stat the file
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	addr := net.JoinHostPort(ip, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer conn.Close()

	// Send header
	header := fmt.Sprintf("FILE|%s|%s|%d\n", sender, info.Name(), info.Size())
	_, err = conn.Write([]byte(header))
	if err != nil {
		return fmt.Errorf("send file header: %w", err)
	}

	// Stream file bytes
	written, err := io.Copy(conn, f)
	if err != nil {
		return fmt.Errorf("send file data: %w", err)
	}

	fmt.Printf("[sent] file '%s' (%d bytes) to %s\n", info.Name(), written, addr)
	return nil
}
