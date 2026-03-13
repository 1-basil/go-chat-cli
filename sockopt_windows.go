//go:build windows

package main

import (
	"net"
	"syscall"
)

// reuseListenConfig returns a net.ListenConfig that sets SO_REUSEADDR and SO_BROADCAST,
// allowing multiple processes to share the port and allowing broadcast packets to be sent.
func reuseListenConfig() net.ListenConfig {
	return net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
			})
		},
	}
}
