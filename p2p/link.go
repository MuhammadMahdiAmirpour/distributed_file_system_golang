package p2p

import "net"

// Node represents a remote node in the network and extends the net.Conn interface,
// providing methods for sending data and closing streams specifically in the context
// of a distributed network.
// Methods:
//   - Send([]byte) error: Sends a byte slice of data to the node. Returns an error if the send operation fails.
//   - CloseStream(): Closes the data stream to the node, typically used when a message or transmission has been completed.
type Node interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Link is an abstraction that represents a communication channel between nodes in the network,
// allowing for different underlying protocols (e.g., TCP, UDP, websockets) to be used for node communication.
// Methods:
//   - Addr() string: Returns the local network address of the link as a string.
//   - Dial(string) error: Connects to a specified address string. Returns an error if the connection fails.
//   - ListenAndAccept() error: Starts listening for incoming connections and accepts them. Returns an error if the operation fails.
//   - Consume() <-chan RPC: Provides a read-only channel to consume incoming RPC (Remote Procedure Call) messages from connected nodes.
//   - Close() error: Closes the link and any active connections, returning an error if the close operation encounters issues.
type Link interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
