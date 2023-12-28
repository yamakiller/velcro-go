package proxy

import "google.golang.org/protobuf/proto"

type RpcProxyStrategy interface {
	RequestMessage(string, proto.Message, int64) (proto.Message, error)
}
