package p2p

import "net"

// Represents the remote node
type Peer interface {
	net.Conn
	Close() error
}

// Transport is anything that handles the communication
// between the nodes io the network. This can be of the
// corm: TCP, UDP, websockets..
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
