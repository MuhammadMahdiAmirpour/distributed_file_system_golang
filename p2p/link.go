package p2p

import "net"

// Node is the remote node in the network
type Node interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Link handles communication between nodes in the network (TCP, UDP, websockets, ...)
type Link interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
