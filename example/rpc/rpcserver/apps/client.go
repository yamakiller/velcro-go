package apps

import (
	"fmt"
	"os"

	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/rpc/server"
	"google.golang.org/protobuf/proto"
)

type TestClient struct {
	server.RpcClient
}

func (tc *TestClient) OnAuth(ctx *server.RpcClientContext) (proto.Message, error) {
	requst := ctx.Message().(*protos.Auth)
	fmt.Fprintf(os.Stderr, "123456")
	return requst, nil
}
