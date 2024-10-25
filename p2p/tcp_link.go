package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents a remote node in a TCP-based network connection.
//
// Fields:
//   - Conn: The network connection interface for this peer.
//   - outbound:
//     A boolean indicating if this peer connection was established by dialing out
//     (true) or by accepting an incoming connection
//     (false).
//   - Wg: A WaitGroup for synchronizing the closing of streams associated with the peer.
type TCPPeer struct {
	net.Conn
	outbound bool
	wg       *sync.WaitGroup
}

// CloseStream signals the completion of a stream by decrementing the WaitGroup counter.
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

// NewTCPPeer creates and returns a new TCPPeer instance, initializing its connection, outbound status, and WaitGroup.
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

// Send transmits a byte slice of data to the peer over the network connection.
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// TCPTransportOpts contains configuration options for initializing a TCPTransport instance.
//
// Fields:
//   - ListenAddr: The address that the transport will bind to for listening to incoming connections.
//   - HandshakeFunc: A function used to perform any necessary handshake when establishing a peer connection.
//   - Decoder: A decoder instance to decode incoming data into RPC structs.
//   - OnNode: A callback function that is invoked when a new node (peer) is established.
type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnNode        func(Node) error
}

// TCPTransport manages TCP-based network transport for communication between nodes in a network.
//
// Fields:
//   - TCPTransportOpts: The configuration options for this transport.
//   - listener: The net.Listener instance for accepting incoming connections.
//   - rpcch: A channel for receiving RPC messages from other nodes.
//   - mu: A mutex for synchronizing access to peer connections.
//   - Peers: A map of active peer nodes, keyed by their network addresses.
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
	mu       sync.RWMutex
	peers    map[net.Addr]Node
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}

// NewTCPTransport initializes and returns a new TCPTransport instance with the specified options.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}
}

// Addr returns the address that the transport is bound to for accepting connections.
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

// Consume returns a read-only channel for receiving incoming RPC messages from peers in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// ListenAndAccept starts listening on the configured address
// and begins accepting incoming connections in a separate goroutine.
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go func() {
		t.startAcceptLoop()
	}()
	log.Printf("TCP transport listening on port: %s\n", t.ListenAddr)
	return nil
}

// Close closes the listener, stopping the transport from accepting further connections.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// handleConn manages a single peer connection, handling the handshake and reading incoming messages.
//
// Parameters:
//   - conn: The network connection to the peer.
//   - outbound: Indicates if the connection was dialed (outbound) or accepted (inbound).
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		err = conn.Close()
		if err != nil {
			return
		}
	}()
	peer := NewTCPPeer(conn, outbound)
	if err = t.HandshakeFunc(peer); err != nil {
		err = conn.Close()
		if err != nil {
			fmt.Printf("TCP handshake error: %s\n", err)
			return
		}
		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}
	if t.OnNode != nil {
		if err = t.OnNode(peer); err != nil {
			return
		}
	}
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr().String()
		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}
		fmt.Println("waiting until stream is done")
		t.rpcch <- rpc
		fmt.Println("stream done, continuing normal read loop")
	}
}

// startAcceptLoop continuously accepts incoming connections and spawns a goroutine to handle each one.
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
		go t.handleConn(conn, false)
	}
}
