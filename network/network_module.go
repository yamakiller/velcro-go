package network

type NetworkModule interface {
	Open(string) error
	Network() string
	Stop()
}
