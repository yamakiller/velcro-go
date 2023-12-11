package network

import "net"

type ProducerWithClientSystem func(system *NetworkSystem, conn net.Conn) error

type Props struct {
	_spawnHandler SpawnHandler
	_producer     ProducerWithClientSystem
}

func (props *Props) Configure(opts ...PropsOption) *Props {
	for _, opt := range opts {
		opt(props)
	}

	return props
}
