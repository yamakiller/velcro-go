package streaming

import (
	"context"
	"io"

	"google.golang.org/grpc/metadata"
)

// Stream both client and server stream
type Stream interface {
	// SetHeader sets the header metadata. It may be called multiple times.
	// When call multiple times, all the provided metadata will be merged.
	// All the metadata will be sent out when one of the following happens:
	//  - ServerStream.SendHeader() is called;
	//  - The first response is sent out;
	//  - An RPC status is sent out (error or success).
	SetHeader(metadata.MD) error
	// SendHeader sends the header metadata.
	// The provided md and headers set by SetHeader() will be sent.
	// It fails if called multiple times.
	SendHeader(metadata.MD) error
	// SetTrailer sets the trailer metadata which will be sent with the RPC status.
	// When called more than once, all the provided metadata will be merged.
	SetTrailer(metadata.MD)
	// Header is used for client side stream to receive header from server.
	Header() (metadata.MD, error)
	// Trailer is used for client side stream to receive trailer from server.
	Trailer() metadata.MD
	// Context the stream context.Context
	Context() context.Context
	// RecvMsg recvive message from peer
	// will block until an error or a message received
	// not concurrent-safety
	RecvMsg(m interface{}) error
	// SendMsg send message to peer
	// will block until an error or enough buffer to send
	// not concurrent-safety
	SendMsg(m interface{}) error
	// not concurrent-safety with SendMsg
	io.Closer
}

// Args endpoint request
type Args struct {
	Stream Stream
}

// Result endpoint response
type Result struct {
	Stream Stream
}
