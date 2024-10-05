package p2p

// HandshakeFunc is a function for handshake before starting a connection
type HandshakeFunc func(Peer) error

func NOPHHandshakeFunc(Peer) error { return nil }
