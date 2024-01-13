package client

import (
	"context"

	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
)

type LongConnPool interface {
	Get(ctx context.Context, address string) (*LongConn, error)

	Put(*LongConn) error

	Discard(*LongConn) error

	Close() error
}

type DefaultConnPool struct {
	pls          *intrusive.Linked
	openingConns int32
	address      string
}
