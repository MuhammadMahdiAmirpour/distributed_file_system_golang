package p2p

// HandshakeFunc is a function for handshake before starting a connection
type HandshakeFunc func(Node) error

func NOPHandshakeFunc(Node) error { return nil }
