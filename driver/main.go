package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/muhammadmahdiamirpour/distributed-file-system/crypto"
	"github.com/muhammadmahdiamirpour/distributed-file-system/p2p"
	"github.com/muhammadmahdiamirpour/distributed-file-system/server"
	"github.com/muhammadmahdiamirpour/distributed-file-system/storage"
)

func makeServer(listenAddr string, nodes ...string) *server.FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := server.FileServerOpts{
		EncKey:            crypto.NewEncryptionKey(),
		StorageRoot:       listenAddr + "_net",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := server.NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":7000", ":3000")
	s3 := makeServer(":5000", ":3000", ":7000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(2 * time.Second)
	go func() {
		log.Fatal(s2.Start())
	}()
	time.Sleep(2 * time.Second)
	go func() {
		log.Fatal(s3.Start())
	}()
	time.Sleep(2 * time.Second)
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("pic_%d.png", i)
		data := bytes.NewReader([]byte("my very big data is here"))
		if err := s3.Store(key, data); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		err := s3.Storage.Delete(s3.ID, key)
		if err != nil {
			return
		}
		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
