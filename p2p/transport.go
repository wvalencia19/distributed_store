package p2p

// Represents the remote node
type Peer interface {
}

// Transport is anything that handles the communication
// between the nodes io the network. This can be of the
// corm: TCP, UDP, websockets..
type Transport interface {
}