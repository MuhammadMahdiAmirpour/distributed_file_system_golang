package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
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
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes, // BootstrapNodes to connect with other nodes
	}

	s := server.NewFileServer(fileServerOpts)

	tcpTransport.OnNode = s.OnNode

	return s
}

func waitForNodes(nodes []string) {
	for _, node := range nodes {
		for {
			_, err := net.Dial("tcp", node)
			if err == nil {
				fmt.Printf("Successfully connected to %s\n", node)
				break
			}
			fmt.Printf("Waiting for %s to be available...\n", node)
			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	listenAddr := os.Getenv("NODE_PORT")
	bootstrapNodesEnv := os.Getenv("BOOTSTRAP_NODES")
	nodeName := os.Getenv("NODE_NAME")

	var bootstrapNodes []string
	if bootstrapNodesEnv != "" {
		bootstrapNodes = strings.Split(bootstrapNodesEnv, ",")
	}

	// Wait for bootstrap nodes to be available
	waitForNodes(bootstrapNodes)

	s := makeServer(listenAddr, bootstrapNodes...)

	go func() {
		log.Fatal(s.Start())
	}()

	if nodeName == "node3" {
		time.Sleep(5 * time.Second) // Wait for other nodes to start
		runDriverCode(s)
	}
	select {} // Block indefinitely
}

func runDriverCode(s *server.FileServer) {
	for i := 0; i < 20; i++ {
		fmt.Println("----------------------------------------------------------------------------------")
		fmt.Printf("iteration: %d\n", i)
		fmt.Println("----------------------------------------------------------------------------------")
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my very big data file here!"))
		err := s.Store(key, data)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			continue
		}

		if err := s.Storage.Delete(s.ID, key); err != nil {
			return
		}

		// Read the file back
		r, err := s.Get(key)
		if err != nil {
			log.Printf("Error reading file: %v\n", err)
			continue
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Printf("Error reading file contents: %v\n", err)
			continue
		}

		fmt.Println(string(b))

		// Delete the file after successful read
		if err := s.Storage.Delete(s.ID, key); err != nil {
			log.Printf("Error deleting file: %v\n", err)
		}
	}
}
