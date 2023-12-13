package network

type ProducerWidthClientSystem func(*NetworkSystem) Client

type Client interface {
	Accept(Context)
	Ping(Context)
	Recvice(Context)
	Closed(Context)
}
