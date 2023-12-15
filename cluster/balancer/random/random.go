package random

import (
	"context"
	"math/rand"

	"github.com/yamakiller/velcro-go/cluster/balancer"
	"github.com/yamakiller/velcro-go/cluster/naming"
)

var (
	_ balancer.Balancer = &Picker{}

	Name = "random"
)

type Picker struct {
}

func New() *Picker {
	return &Picker{}
}

func (p *Picker) Pick(ctx context.Context, nodes []naming.Instance) (node *naming.Instance, done func(balancer.DoneInfo), e error) {
	if len(nodes) == 0 {
		return nil, func(balancer.DoneInfo) {}, nil
	}
	cur := rand.Intn(len(nodes))
	return &nodes[cur], func(balancer.DoneInfo) {}, nil
}

func (p *Picker) Schema() string {
	return Name
}
