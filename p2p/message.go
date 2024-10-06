package p2p

import "net"

// RPC is a container of the data that is being transferred over the network
type RPC struct {
	From    net.Addr
	Payload []byte
}
