package main

import (
	"log"

	"github.com/wvalencia19/distribuited_store/data"
	"github.com/wvalencia19/distribuited_store/p2p"
)

func makeServer(listedAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listedAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       &p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpTransportOpts)

	fileserverOpts := FileServerOpts{
		StorageRoot:       listedAddr + "_network",
		PathTransformFunc: data.CASPathTransformFun,
		Transport:         tr,
		BootstrapNodes:    nodes,
	}

	return NewFileServer(fileserverOpts)
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()

	s2.Start()

}
