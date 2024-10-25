package p2p

// HandshakeFunc defines a function type for performing a handshake operation before establishing a connection with a peer.
// Parameters:
//   - Node: The peer node with which the handshake will be performed.
//
// Returns:
//   - An error if the handshake fails, otherwise nil.
type HandshakeFunc func(Node) error

// NOPHandshakeFunc is a no-operation (NOP) handshake function that simply returns nil.
// This can be used when no handshake logic is required before a connection, acting as a placeholder function.
// Returns:
//   - nil, indicating no error.
func NOPHandshakeFunc(Node) error { return nil }
