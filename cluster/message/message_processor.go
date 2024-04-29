package message

import (
	"context"
	"fmt"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/protocol"
)

/*type IMessageProcessor interface {
	Do(ctx network.Context, msg []byte, timeout int64) error
}*/

type IMessageAgentStruct interface {
	UnMarshal(msg []byte) error
	Method(ctx network.Context, seqid int32, timeout int64) error
}

func NewRepeatMessageAgent() *RepeatMessageAgent {
	return &RepeatMessageAgent{
		agents: make(map[string]IMessageAgentStruct),
		iprot:  protocol.NewBinaryProtocol(),
	}
}

type RepeatMessageAgent struct {
	agents        map[string]IMessageAgentStruct
	default_agent IMessageAgentStruct
	iprot         protocol.IProtocol
}

func (d *RepeatMessageAgent) Register(key string, agent IMessageAgentStruct) {
	d.agents[key] = agent
}

func (d *RepeatMessageAgent) WithDefaultMethod(agent IMessageAgentStruct) {
	d.default_agent = agent
}

func (d *RepeatMessageAgent) Message(ctx network.Context, msg []byte, timeout int64) error {
	d.iprot.Release()
	d.iprot.Write(msg)
	name, _, seqid, err := d.iprot.ReadMessageBegin(context.Background())
	if err != nil {
		return err
	}
	if info, ok := d.agents[name]; ok {
		if err := info.UnMarshal(msg); err != nil {
			return err
		}
		return info.Method(ctx, seqid, timeout)
	}
	if d.default_agent != nil {
		if err := d.default_agent.UnMarshal(msg); err != nil {
			return err
		}
		return d.default_agent.Method(ctx, seqid, timeout)
	}
	return fmt.Errorf("unkown message %s", name)
}
