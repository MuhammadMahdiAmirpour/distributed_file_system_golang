package p2p

const (
	// IncomingMessage is a constant used to indicate that the incoming data is a discrete message.
	// It is represented by the byte value 0x1.
	IncomingMessage = 0x1

	// IncomingStream is a constant used to signify that the incoming data should be treated as a continuous stream.
	// It is represented by the byte value 0x2.
	IncomingStream = 0x2
)

// RPC is a structure representing a data container for transferring information over the network between nodes.
// Fields:
//   - From string: Identifies the sender of the RPC message, typically as an address string.
//   - Payload []byte: The byte slice containing the actual data or message content being transferred.
//   - Stream bool: A boolean flag indicating whether the RPC is a continuous stream (true) or a single message (false).
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
