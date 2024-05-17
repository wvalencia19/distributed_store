package main

import (
	"bytes"
	"log"
	"time"

	"github.com/wvalencia19/distribuited_store/data"
	"github.com/wvalencia19/distribuited_store/p2p"
)

func makeServer(listedAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listedAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       &p2p.DefaultDecoder{},
		// OnPeer: ,
	}

	tr := p2p.NewTCPTransport(tcpTransportOpts)

	fileserverOpts := FileServerOpts{
		StorageRoot:       listedAddr + "_network",
		PathTransformFunc: data.CASPathTransformFun,
		Transport:         tr,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileserverOpts)
	tr.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(1 * time.Second)

	go s2.Start()

	time.Sleep(1 * time.Second)

	data := bytes.NewReader([]byte("my big data file here"))
	s2.StoreData("myprivatedata", data)

	select {}
}
