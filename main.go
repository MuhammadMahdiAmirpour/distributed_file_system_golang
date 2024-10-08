package main

import (
	"log"
	"time"

	"github.com/muhammadmahdiamirpour/distributed-file-system/p2p"
)

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: onPeer func
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		StorageRoot:       "3000_root_dir",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
	}
	s := NewFileServer(fileServerOpts)
	go func() {
		time.Sleep(3 * time.Second)
		s.Stop()
	}()
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
