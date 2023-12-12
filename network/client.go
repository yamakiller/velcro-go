package network

type ProducerWidthClientSystem func(*NetworkSystem) Client

type Client interface {
	Accept(Context)
	Recvice(Context)
	Closed(Context)
}
