package main

import (
	"fmt"
	"log"

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
	}
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
			fmt.Println(msg)
		case <-f.quitCh:
			return nil
		}
	}
}
