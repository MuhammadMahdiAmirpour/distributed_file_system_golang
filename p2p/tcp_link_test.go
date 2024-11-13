package p2p

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define a mock failing handshake function that implements HandshakeFunc
func mockFailingHandshake(Node) error {
	return assert.AnError
}

// Define a mock successful handshake function
func mockSuccessHandshake(Node) error {
	return nil
}

func TestTCPTransport_Basic(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: mockSuccessHandshake,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Equal(t, tr.ListenAddr, ":3000")

	// Start transport in goroutine
	go func() {
		err := tr.ListenAndAccept()
		assert.Nil(t, err)
	}()

	// Give time for server to start
	time.Sleep(100 * time.Millisecond)

	// Test basic connection
	conn, err := net.Dial("tcp", ":3000")
	require.NoError(t, err)
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal("error closing connection: ", err)
		}
	}(conn)

	// Cleanup
	err = tr.Close()
	assert.Nil(t, err)
}

func TestTCPTransport_Dial(t *testing.T) {
	// Server transport
	serverOpts := TCPTransportOpts{
		ListenAddr:    ":3001",
		HandshakeFunc: mockSuccessHandshake,
		Decoder:       DefaultDecoder{},
	}
	serverTr := NewTCPTransport(serverOpts)

	// Initialize server
	err := serverTr.ListenAndAccept()
	require.NoError(t, err)

	// Give time for server to start
	time.Sleep(100 * time.Millisecond)

	// Client transport
	clientOpts := TCPTransportOpts{
		ListenAddr:    ":3002",
		HandshakeFunc: mockSuccessHandshake,
		Decoder:       DefaultDecoder{},
	}
	clientTr := NewTCPTransport(clientOpts)

	// Test dialing
	err = clientTr.Dial(":3001")
	assert.Nil(t, err)

	// Give time for connection to establish
	time.Sleep(100 * time.Millisecond)

	// Cleanup - make sure to check if listener exists before closing
	if serverTr.listener != nil {
		err = serverTr.Close()
		assert.Nil(t, err)
	}

	if clientTr.listener != nil {
		err = clientTr.Close()
		assert.Nil(t, err)
	}
}

func TestTCPTransport_HandshakeFailure(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr:    ":3003",
		HandshakeFunc: mockFailingHandshake,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)

	go func() {
		err := tr.ListenAndAccept()
		assert.Nil(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	// Connection should be closed after failed handshake
	conn, err := net.Dial("tcp", ":3003")
	require.NoError(t, err)
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal("error closing connection: ", err)
		}
	}(conn)

	// Try to read - should fail due to connection being closed after handshake failure
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	assert.Error(t, err)

	if err := tr.Close(); err != nil {
		log.Fatal("error closing connection: ", err)
	}
}
