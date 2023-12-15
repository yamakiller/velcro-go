package balancer

import (
	"context"

	"github.com/yamakiller/velcro-go/cluster/naming"
)

// DoneInfo is callback when rpc done
type DoneInfo struct {
	Err     error
	Trailer map[string]string
}

// Balancer is picker
type Balancer interface {
	Pick(ctx context.Context, nodes []naming.Instance) (node *naming.Instance, done func(DoneInfo), e error)
	Schema() string
}

type Printable interface {
	PrintStats()
}
