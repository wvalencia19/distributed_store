package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/wvalencia19/distribuited_store/data"
	"github.com/wvalencia19/distribuited_store/p2p"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc data.PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *data.Store
	quitCh chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := data.StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          data.NewStore(storeOpts),
		quitCh:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (f *FileServer) OnPeer(p p2p.Peer) error {
	f.peerLock.Lock()
	defer f.peerLock.Unlock()

	f.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s", p.RemoteAddr())

	return nil
}

func (f *FileServer) bootstrapNetwork() error {
	for _, addr := range f.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			if err := f.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}

	return nil
}

func (f *FileServer) Start() error {
	if err := f.Transport.ListenAndAccept(); err != nil {
		return err
	}

	f.bootstrapNetwork()
	f.loop()

	return nil
}

type Payload struct {
	Key  string
	Data []byte
}

func (f *FileServer) broadcast(p *Payload) error {
	peers := []io.Writer{}

	for _, peer := range f.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}

func (f *FileServer) StoreData(key string, r io.Reader) error {
	//1 store the file to disk
	//2 broadcast this file to all known peers in the network

	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)

	if err := f.store.Write(key, tee); err != nil {
		return err
	}

	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}

	p := &Payload{
		Key:  key,
		Data: buf.Bytes(),
	}

	fmt.Println(buf.Bytes())

	return f.broadcast(p)
}
func (f *FileServer) Stop() {
	close(f.quitCh)
}

func (f *FileServer) loop() error {
	defer func() {
		log.Println("fileserver stopped due to user quit action")
		f.Transport.Close()
	}()
	for {
		select {
		case msg := <-f.Transport.Consume():
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%+v\n", p)
		case <-f.quitCh:
			return nil
		}
	}
}
