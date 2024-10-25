package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/muhammadmahdiamirpour/distributed-file-system/crypto"
	"github.com/muhammadmahdiamirpour/distributed-file-system/p2p"
	"github.com/muhammadmahdiamirpour/distributed-file-system/storage"
)

// FileServerOpts defines options used for configuring the FileServer instance.
type FileServerOpts struct {
	ID                string                    // Unique identifier for the server node
	EncKey            []byte                    // Encryption key for file storage and transmission
	StorageRoot       string                    // Root path for file storage
	PathTransformFunc storage.PathTransformFunc // Function to transform file paths based on the key
	Transport         p2p.Link                  // Transport layer for peer-to-peer communication
	BootstrapNodes    []string                  // List of nodes for initial network bootstrap
}

// FileServer represents the main server responsible for managing files in a distributed manner.
type FileServer struct {
	FileServerOpts                     // Embeds options to make configuration easier
	peerLock       sync.Mutex          // Mutex to ensure thread-safe access to peers
	peers          map[string]p2p.Node // Map of connected peers with peer address as a key
	Storage        *storage.Store      // Storage layer to manage local file storage
	quitch         chan struct{}       // Channel to signal termination of the server
}

// NewFileServer initializes and returns a new FileServer instance.
// It sets up storage with the provided options and generates a unique ID if not supplied.
func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := storage.StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	if len(opts.ID) == 0 {
		opts.ID = crypto.GenerateID()
	}
	return &FileServer{
		FileServerOpts: opts,
		Storage:        storage.NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Node),
	}
}

// broadcast sends a message to all connected peers in the network.
func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		// Send message type byte indicating the incoming message
		if err := peer.Send([]byte{p2p.IncomingMessage}); err != nil {
			return err
		}
		// Send the actual message payload
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// Message defines a generic message with a payload that can hold any data type.
type Message struct {
	Payload any
}

// MessageStoreFile represents a message for storing a file with ID, encryption key, and size.
type MessageStoreFile struct {
	ID   string // Unique identifier for the message
	Key  string // Encrypted key for the file
	Size int64  // Size of the file to store
}

// MessageGetFile represents a request message to get a file with ID and encryption key.
type MessageGetFile struct {
	ID  string // Identifier for the file
	Key string // Encrypted key to retrieve the file
}

// Get retrieves a file by key.
// If it exists locally, it is read from local storage.
// If not, it broadcasts a network request to retrieve the file from peers.
func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.Storage.Has(s.ID, key) {
		fmt.Printf("[%s] serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.Storage.Read(s.ID, key)
		return r, err
	}
	fmt.Printf("File %s not found locally, fetching from network...\n", key)
	msg := Message{
		Payload: MessageGetFile{
			ID:  s.ID,
			Key: crypto.HashKey(key),
		},
	}
	// Broadcast request to peers
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	// Wait briefly to allow peers to respond
	time.Sleep(500 * time.Millisecond)
	for _, peer := range s.peers {
		// Receive the file size before reading the data
		var fileSize int64
		err := binary.Read(peer, binary.LittleEndian, &fileSize)
		if err != nil {
			return nil, err
		}
		// Limit the bytes read based on the file size to avoid hanging
		n, err := s.Storage.WriteDecrypt(s.EncKey, s.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received (%d) bytes over the network from (%s)\n", s.Transport.Addr(), n, peer.RemoteAddr())
		peer.CloseStream()
	}
	_, r, err := s.Storage.Read(s.ID, key)
	return r, err
}

// Store saves a file locally and broadcasts a storage message to the network.
func (s *FileServer) Store(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)
	size, err := s.Storage.Write(s.ID, key, tee)
	if err != nil {
		return err
	}
	msg := Message{
		Payload: MessageStoreFile{
			ID:   s.ID,
			Key:  crypto.HashKey(key),
			Size: size + 16, // Account for extra encryption padding
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return err
	}
	time.Sleep(5 * time.Millisecond)
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	// Send the file to all connected peers
	mw := io.MultiWriter(peers...)
	_, err = mw.Write([]byte{p2p.IncomingStream})
	if err != nil {
		return err
	}
	n, err := crypto.CopyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.Addr(), n)
	return nil
}

// Stop stops the FileServer by closing the quitch channel.
func (s *FileServer) Stop() {
	close(s.quitch)
}

// OnNode handles a new peer connection by adding it to the peer list.
func (s *FileServer) OnNode(p p2p.Node) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("connected to remote %s", p.RemoteAddr())
	return nil
}

// loop is the main event loop for processing incoming messages and terminating when quitch is closed.
func (s *FileServer) loop() {
	defer func() {
		log.Println("File server stopped")
		err := s.Transport.Close()
		if err != nil {
			fmt.Printf("Error closing transport: %s", err)
			return
		}
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Printf("decoding error: %s", err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("Error handling message", err)
			}
		case <-s.quitch:
			return
		}
	}
}

// handleMessage handles incoming messages and dispatches them based on a message type.
func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

// handleMessageStoreFile handles a request to store a file and writes it locally.
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found", from)
	}
	n, err := s.Storage.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("[%s] written %d bytes to disk\n", s.Transport.Addr(), n)
	peer.CloseStream()
	return nil
}

// handleMessageGetFile handles a request to retrieve a file, sending it to the requesting peer.
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.Storage.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] file (%s) not found on disk", s.Transport.Addr(), msg.Key)
	}
	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)
	fileSize, r, err := s.Storage.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}
	if rc, ok := r.(io.ReadCloser); ok {
		defer func(rc io.ReadCloser) {
			err := rc.Close()
			if err != nil {
				log.Fatal("error closing the connection in function handleMessageGetFile")
			}
		}(rc)
	}
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found", from)
	}
	// First send the "incomingStream" byte to the peer, and then we can send the file size as an int64
	err = peer.Send([]byte{p2p.IncomingStream})
	if err != nil {
		return err
	}
	// Send file size before sending data
	err = binary.Write(peer, binary.LittleEndian, fileSize)
	if err != nil {
		return err
	}
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] written %d bytes over the networks to %s\n", s.Transport.Addr(), n, from)
	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Printf("[%s] attempting to connect with remote: %s\n", s.Transport.Addr(), addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error", err)
			}
		}(addr)
	}
	return nil
}

func (s *FileServer) Start() error {
	fmt.Printf("[%s] starting fileserver...\n", s.Transport.Addr())
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	err := s.bootstrapNetwork()
	if err != nil {
		return err
	}
	s.loop()
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
