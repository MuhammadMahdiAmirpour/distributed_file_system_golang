package p2p

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDefaultDecoder tests the DefaultDecoder implementation.
func TestDefaultDecoder(t *testing.T) {
	decoder := DefaultDecoder{}
	var rpc RPC

	// Test decoding a standard message.
	message := append([]byte{IncomingMessage}, []byte("test message")...)
	err := decoder.Decode(bytes.NewReader(message), &rpc)
	assert.Nil(t, err)
	assert.False(t, rpc.Stream) // Not a stream
	assert.Equal(t, []byte("test message"), rpc.Payload)

	// Test decoding a stream message.
	stream := append([]byte{IncomingStream}, []byte("stream data")...)
	err = decoder.Decode(bytes.NewReader(stream), &rpc)
	assert.Nil(t, err)
	assert.True(t, rpc.Stream) // It should be marked as a stream
}

// TestGOBDecoder tests the GOBDecoder implementation.
func TestGOBDecoder(t *testing.T) {
	decoder := GOBDecoder{}
	var rpc RPC

	// Prepare a GOB-encoded RPC message.
	originalRPC := RPC{
		From:    "127.0.0.1",
		Payload: []byte("gob encoded message"),
		Stream:  false,
	}
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(&originalRPC)
	assert.Nil(t, err)

	// Test decoding the GOB message.
	err = decoder.Decode(&buffer, &rpc)
	assert.Nil(t, err)
	assert.Equal(t, originalRPC.From, rpc.From)
	assert.Equal(t, originalRPC.Payload, rpc.Payload)
	assert.Equal(t, originalRPC.Stream, rpc.Stream)
}
