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

type FileServerOpts struct {
	ID                string
	EncKey            []byte
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
	Transport         p2p.Link
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts
	peerLock sync.Mutex
	peers    map[string]p2p.Node
	Storage  *storage.Store
	quitch   chan struct{}
}

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

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send([]byte{p2p.IncomingMessage}); err != nil {
			return err
		}
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	ID   string
	Key  string
	Size int64
}

type MessageGetFile struct {
	ID  string
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.Storage.Has(s.ID, key) {
		fmt.Printf("[%s] serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.Storage.Read(s.ID, key)
		return r, err
	}
	fmt.Printf("do not have file %s locally, fetching from network...\n", key)
	msg := Message{
		Payload: MessageGetFile{
			ID:  s.ID,
			Key: crypto.HashKey(key),
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(500 * time.Millisecond)
	for _, peer := range s.peers {
		// First, read the file size,
		// so we can limit the number of bytes that we read from the connection so it will not keep hanging
		var fileSize int64
		err := binary.Read(peer, binary.LittleEndian, &fileSize)
		if err != nil {
			return nil, err
		}
		n, err := s.Storage.WriteDecrypt(s.EncKey, s.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received (%d) bytes over the network form (%s)\n", s.Transport.Addr(), n, peer.RemoteAddr())
		peer.CloseStream()
	}
	_, r, err := s.Storage.Read(s.ID, key)
	return r, err
}

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
			Size: size + 16,
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
	mw := io.MultiWriter(peers...)
	_, err = mw.Write([]byte{p2p.IncomingStream})
	if err != nil {
		return err
	}
	n, err := crypto.CopyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] received and written (%d) bytes to disk: \n", s.Transport.Addr(), n)
	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnNode(p p2p.Node) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("connected to remote %s", p.RemoteAddr())
	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to error or user quit action")
		err := s.Transport.Close()
		if err != nil {
			fmt.Printf("error closing transport %s", err)
			return
		}
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Printf("decoding error : %s", err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("error handling message", err)
			}
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	n, err := s.Storage.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("[%s] written %d bytes to disk \n", s.Transport.Addr(), n)
	peer.CloseStream()
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.Storage.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk", s.Transport.Addr(), msg.Key)
	}
	fmt.Printf("[%s] serving file (%s) over the network", s.Transport.Addr(), msg.Key)
	fileSize, r, err := s.Storage.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}
	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Printf("close readCloser")
		defer func(rc io.ReadCloser) {
			if err := rc.Close(); err != nil {
				fmt.Printf("error closing readCloser: %s", err)
			}
		}(rc)
	}
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found in the map", from)
	}
	// First send the "incomingStream" byte to the peer and then we can send the file sized as an int64
	err = peer.Send([]byte{p2p.IncomingStream})
	if err != nil {
		return err
	}
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
