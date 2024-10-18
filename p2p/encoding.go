// Package p2p provides functionality for peer-to-peer communication.
package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder is an interface for decoding incoming messages.
type Decoder interface {
	// Decode decodes an incoming message from the provided io.Reader into the given RPC struct.
	Decode(io.Reader, *RPC) error
}

// GOBDecoder is a struct that implements the Decoder interface using the gob package.
type GOBDecoder struct{}

// Decode decodes an incoming message from the provided io.Reader into the given RPC struct.
func (dec GOBDecoder) Decode(reader io.Reader, msg *RPC) error {
	// Use the gob package to decode the incoming message.
	return gob.NewDecoder(reader).Decode(msg)
}

// DefaultDecoder is a struct that implements the Decoder interface with custom decoding logic.
type DefaultDecoder struct{}

// Decode decodes an incoming message from the provided io.Reader into the given RPC struct.
// If the message is a stream, sets the Stream field of the RPC struct to true and returns without decoding the payload.
func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	// Read the first byte of the incoming message to determine if it's a stream.
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		return err
	}
	// Check if the message is a stream.
	stream := peekBuf[0] == IncomingStream
	if stream {
		// Set the Stream field of the RPC struct to true and return.
		msg.Stream = true
		return nil
	}
	// Read the rest of the incoming message into a buffer.
	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	// Set the Payload field of the RPC struct to the decoded message.
	msg.Payload = buf[:n]
	return nil
}
