package loadbalance

import (
	"context"

	"github.com/yamakiller/velcro-go/rpc2/pkg/discovery"
)

var _ Picker = &DummyPicker{}

// DummyPicker is a picker that always returns nil on Next.
type DummyPicker struct{}

// Next implements the Picker interface.
func (np *DummyPicker) Next(ctx context.Context, request interface{}) (ins discovery.Instance) {
	return nil
}
