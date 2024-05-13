package p2p

import "net"

// RPC holds any arbitrary data is being sent over the
// each transport between two nodes in the network
type RPC struct {
	From    net.Addr
	Payload []byte
}
