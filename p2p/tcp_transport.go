package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer is the remote node in a TCP connection
type TCPPeer struct {
	conn net.Conn

	// dial and retrieve a connection -> outbound = true
	// accept and retrieve a connection -> outbound = false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{conn: conn, outbound: outbound}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	handshakeFunc HandshakeFunc
	decoder       Decoder
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		handshakeFunc: func(Peer) error { return nil },
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}
	go func() {
		err := t.startAcceptLoop()
		if err != nil {
			fmt.Println(err)
		}
	}()
	return nil
}

func (t *TCPTransport) startAcceptLoop() error {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			return err
		}
		fmt.Printf("new incoming connection %s\n", conn)
		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	if err := t.handshakeFunc(peer); err != nil {
	}
	// buf := new(bytes.Buffer)
	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
	}
}
