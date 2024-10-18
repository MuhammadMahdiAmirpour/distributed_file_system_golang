package p2p

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC is a container of the data that is being transferred over the network
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
