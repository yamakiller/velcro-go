package balancer

import "errors"

var (
	NoHostError                = errors.New("no host")
	AlgorithmNotSupportedError = errors.New("algorithm not supported")
)

type Balancer interface {
	Add(string)
	Remove(string)
	Balance(string) (string, error)
	Inc(string)
	Done(string)
}

// Factory 是生成Balancer的工厂, 这里采用了工厂设计模式.
type Factory func([]string) Balancer

var factories = make(map[string]Factory)

// Build generates the corresponding Balancer according to the algorithm
func Build(algorithm string, hosts []string) (Balancer, error) {
	factory, ok := factories[algorithm]
	if !ok {
		return nil, AlgorithmNotSupportedError
	}
	return factory(hosts), nil
}
