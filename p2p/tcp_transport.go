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

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	mu       sync.RWMutex
	peers    map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
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
		fmt.Printf("new incoming connection %s\n", conn.RemoteAddr())
		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	if err := t.HandshakeFunc(peer); err != nil {
		err := conn.Close()
		if err != nil {
			fmt.Printf("TCP handshake error: %s\n", err)
			return
		}
		return
	}
	// buf := new(bytes.Buffer)
	msg := &Temp{}
	for {
		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
		fmt.Printf("message: %+v\n", msg)
	}
}
