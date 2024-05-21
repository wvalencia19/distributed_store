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

type Message struct {
	From    string
	Payload any
}

type DataMessage struct {
	Key  string
	Data []byte
}

func (f *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range f.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (f *FileServer) StoreData(key string, r io.Reader) error {
	//1 store the file to disk
	//2 broadcast this file to all known peers in the network

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: []byte("storagekey"),
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range f.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil

	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// if err := f.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// msg := &Message{
	// 	From:    "todo",
	// 	Payload: p,
	// }

	// fmt.Println(buf.Bytes())

	// return f.broadcast(msg)
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
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&m); err != nil {
				log.Println(err)
			}
			if err := f.handleMessage(&m); err != nil {
				log.Println(err)
			}
		case <-f.quitCh:
			return nil
		}
	}
}

func (f *FileServer) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case *DataMessage:
		fmt.Printf("received data %+v\n", v)
	}
	return nil
}
