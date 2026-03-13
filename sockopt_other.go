//go:build !windows

package main

import (
	"net"
	"syscall"
)

// reuseListenConfig returns a net.ListenConfig that sets SO_REUSEADDR, SO_REUSEPORT, and SO_BROADCAST.
// This allows multiple processes to share the port and enabling broadcast sending.
func reuseListenConfig() net.ListenConfig {
	return net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var opErr error
			err := c.Control(func(fd uintptr) {
				// Standard reuse
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				// Unix-specific reuse port (helpful for macOS)
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
				// Permission to send broadcast
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
			})
			if err != nil {
				return err
			}
			return opErr
		},
	}
}
