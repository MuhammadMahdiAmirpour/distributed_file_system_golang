package p2p

import "net"

// Peer is the remote node in the network
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Transport handles communication between nodes in the network (TCP, UDP, websockets, ...)
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
