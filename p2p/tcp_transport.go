package p2p

import (
	"errors"
	"fmt"
	"log"
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

// Close implements Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
	mu       sync.RWMutex
	peers    map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume returns read-only channel for reading the incoming messages received from another peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go func() {
		t.startAcceptLoop()
	}()
	log.Printf("TCP trnasport listening on port: %s\n", t.ListenAddr)
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		fmt.Printf("new incoming connection %+v\n", conn)
		go t.handleConn(conn)
	}
}

// Close implements Transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		err = conn.Close()
		if err != nil {
			return
		}
	}()
	peer := NewTCPPeer(conn, true)
	if err = t.HandshakeFunc(peer); err != nil {
		err = conn.Close()
		if err != nil {
			fmt.Printf("TCP handshake error: %s\n", err)
			return
		}
		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}
	rpc := RPC{}
	for {
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
