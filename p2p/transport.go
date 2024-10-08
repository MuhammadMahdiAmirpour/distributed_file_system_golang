package p2p

// Peer is the remote node in the network
type Peer interface {
	Close() error
}

// Transport handles communication between nodes in the network (TCP, UDP, websockets, ...)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
