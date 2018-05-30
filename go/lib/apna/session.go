package main

import (
	"net"
)

// Session for a APNA connection
type Session struct {
	serverAddr  *net.Addr
	clientAddr  *net.Addr
	serverEphID *EphID
	clientEphID *EphID
}
