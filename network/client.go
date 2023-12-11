package network

type Client interface {
	Recvice(Context)
	Send(Context, interface{})
	Closing(cid *ClientId)
	Closed(cid *ClientId)
}
