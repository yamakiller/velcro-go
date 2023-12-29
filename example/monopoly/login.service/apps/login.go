package apps

import (
	"reflect"

	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/logs"
)

type loginService struct {
	login *serve.Servant
}

func (ls *loginService) Start(logAgent logs.LogAgent) error {

	rds.WithAddr(envs.Instance().Get("configs").(*configs.Config).Redis.Addr)
	rds.WithPwd(envs.Instance().Get("configs").(*configs.Config).Redis.Pwd)
	rds.WithDialTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.DialTimeout)
	rds.WithReadTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.ReadTimeout)
	rds.WithWriteTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.WriteTimeout)

	if err := rds.Connection(); err != nil {
		return err
	}

	ls.login = serve.New(
		serve.WithLoggerAgent(logAgent),
		serve.WithProducerActor(newActor),
		serve.WithFromRouterRecvice(ls.onFromRouteRecvice),
		serve.WithName("loginService"),
		serve.WithLAddr(envs.Instance().Get("configs").(*configs.Config).Server.LAddr),
		serve.WithVAddr(envs.Instance().Get("configs").(*configs.Config).Server.VAddr),
		serve.WithKleepalive(int32(envs.Instance().Get("configs").(*configs.Config).Server.Kleepalive)),
		serve.WithRoute(&envs.Instance().Get("configs").(*configs.Config).Router),
	)

	return nil
}

func (ls *loginService) Stop() error {

	if ls.login != nil {
		ls.login.Stop()
		ls.login = nil
	}

	rds.Disconnect()

	return nil
}

func (ls *loginService) onFromRouteRecvice(msg interface{}) {
	ls.login.System.Info("from route recvice %s", reflect.TypeOf(msg))
}
