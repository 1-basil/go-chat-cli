//go:build windows

package main

import (
	"net"
	"syscall"
)

// reuseListenConfig returns a net.ListenConfig that sets SO_REUSEADDR,
// allowing multiple processes on the same machine to bind to the same UDP port.
func reuseListenConfig() net.ListenConfig {
	return net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			})
		},
	}
}
