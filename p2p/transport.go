package p2p

// Peer is the remote node in the network
type Peer interface {
}

// Transport handles communication between nodes in the network (TCP, UDP, websockets, ...)
type Transport interface {
	ListenAndAccept() error
}
