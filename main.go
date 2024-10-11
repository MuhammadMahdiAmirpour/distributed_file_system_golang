package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/muhammadmahdiamirpour/distributed-file-system/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "_net",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(1 * time.Second)
	go func() {
		log.Fatal(s2.Start())
	}()
	time.Sleep(1 * time.Second)
	// data := bytes.NewReader([]byte("my very big data is here"))
	// if err := s2.Store("tha.jpg", data); err != nil {
	// 	fmt.Printf("Error: %v\n", err)
	// }
	time.Sleep(5 * time.Millisecond)
	r, err := s2.Get("tha.jpg")
	if err != nil {
		log.Fatal(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
