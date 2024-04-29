package server

/*type IMessageProcessor interface {
	Do(ctx network.Context, msg []byte, timeout int64) error
}

type defaultMessageProcessor struct {
	proto protocol.IProtocol
}

func (d *defaultMessageProcessor) Do(ctx network.Context, msg []byte, timeout int64) error {
	d.proto.Release()
	d.proto.Write(msg)
	name, _, seqid, err := d.proto.ReadMessageBegin(context.Background())
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
}*/
