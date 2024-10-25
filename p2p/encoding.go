// Package p2p provides functionality for peer-to-peer communication.
package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder is an interface for decoding incoming messages.
// Implementing structs must provide a Decode method to process incoming
// data into the RPC structure.
type Decoder interface {
	// Decode decodes an incoming message from the provided io.Reader into the given RPC struct.
	//
	// Parameters:
	//   - r: The io.Reader interface for reading the incoming message.
	//   - msg: Pointer to an RPC struct to store the decoded data.
	//
	// Returns: Error if decoding fails, nil otherwise.
	Decode(io.Reader, *RPC) error
}

// GOBDecoder is a struct that implements the Decoder interface using the gob package.
// It encodes and decodes messages in binary format.
type GOBDecoder struct{}

// Decode decodes an incoming message from the provided io.Reader into the given RPC struct.
//
// Parameters:
//   - reader: The io.Reader for reading the incoming gob-encoded message.
//   - Msg: Pointer to the RPC struct where the decoded data will be stored.
//
// Returns: Error if the decoding process fails, nil otherwise.
func (dec GOBDecoder) Decode(reader io.Reader, msg *RPC) error {
	return gob.NewDecoder(reader).Decode(msg)
}

// DefaultDecoder is a struct that implements the Decoder interface with custom decoding logic.
// It can detect if an incoming message is a stream or a standard payload.
type DefaultDecoder struct{}

// Decode decodes an incoming message from the provided io.Reader into the given RPC struct.
// If the message is identified as a stream, it sets the `Stream` field of the RPC struct to true,
// skips decoding the payload, and exits.
//
// Parameters:
//   - r: The io.Reader for reading the incoming message.
//   - msg: Pointer to the RPC struct where the decoded data will be stored.
//
// Behavior:
//   - Reads the first byte of the incoming message to determine if it's a stream.
//   - If the first byte matches the IncomingStream constant, it marks the message as a stream
//     by setting `msg.Stream` to true and returns immediately without further decoding.
//   - Otherwise, it reads the rest of the message into `msg.Payload`.
//
// Returns: Error if the read operation fails, nil otherwise.
func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		return err
	}
	stream := peekBuf[0] == IncomingStream
	if stream {
		msg.Stream = true
		return nil
	}
	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]
	return nil
}
