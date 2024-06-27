package apps

import (
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	_ "github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	_ "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	"github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
)

type loginService struct {
	login *serve.Servant
}

func (ls *loginService) Start() error {

	rds.WithAddr(envs.Instance().Get("configs").(*configs.Config).Redis.Addr)
	rds.WithPwd(envs.Instance().Get("configs").(*configs.Config).Redis.Pwd)
	// rds.WithDialTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Dial)
	// rds.WithReadTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Read)
	// rds.WithWriteTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Write)

	if err := rds.Connection(); err != nil {
		return err
	}

	ls.login = serve.New(
		serve.WithProducerActor(ls.newLoginActor),
		serve.WithName("LoginService"),
		serve.WithLAddr(envs.Instance().Get("configs").(*configs.Config).Server.LAddr),
		serve.WithVAddr(envs.Instance().Get("configs").(*configs.Config).Server.VAddr),
		serve.WithKleepalive(int32(envs.Instance().Get("configs").(*configs.Config).Server.Kleepalive)),
		serve.WithRoute(&envs.Instance().Get("configs").(*configs.Config).Router),
	)

	if err := ls.login.Start(); err != nil {
		rds.Disconnect()
		return err
	}

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

func (ls *loginService) newLoginActor(conn *serve.ServantClientConn) serve.ServantClientActor {
	actor := &LoginActor{ancestor: ls.login}

	conn.Register(&pubs.SignIn{}, actor.onSignIn)
	conn.Register(&pubs.SignOut{}, actor.onSignOut)
	conn.Register(&prvs.ClientClosed{},actor.onClientClosed)
	return actor
}
