package network

type PropsOption func(props *Props)

func WithProducer(p ProducerWithClientSystem) PropsOption {
	return func(props *Props) {
		props._producer = p
	}
}

func PropsFromProducerWithClientSystem(producer ProducerWithClientSystem, opts ...PropsOption) *Props {
	p := &Props{
		_producer: producer,
	}
	p.Configure(opts...)

	return p
}
