package network

type NetworkModule interface {
	Open(string) error
	Stop()
}
