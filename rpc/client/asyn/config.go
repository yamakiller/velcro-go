package asyn

import "google.golang.org/protobuf/proto"

type ConnConfig struct {
	Kleepalive int32

	Connected func()
	Receive   func(proto.Message)
	Closed    func()
}

func defaultConfig() *ConnConfig {
	return &ConnConfig{
		Kleepalive: 10 * 1000,
	}
}
