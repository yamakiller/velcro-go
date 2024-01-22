package remote

type RPCRole int

// RPC role
const (
	Client RPCRole = iota
	Server
)
