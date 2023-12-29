package apps

import (
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/configs"
	"github.com/yamakiller/velcro-go/logs"
)

/*func new(agnet logs.LogAgent) *LoginService {
	return &LoginService{
		Servant: serve.New(
			serve.WithLogger(agnet),
			serve.WithMetricsProvider(localPrometheusProvider(8089)),
			serve.WithProducerActor(newActor),
			serve.WithKleepalive(2000),
		),
	}
}*/

type LoginService struct {
	login *serve.Servant
}

func (ls *LoginService) Start(logAgent logs.LogAgent) error {
	//serve.WithRouterURI(files.NewLocalPathFull("routes.yaml")),
	ls.login = serve.New(
		serve.WithLogger(logAgent),
		serve.WithLAddr(envs.Instance().Get("configs").(*configs.Config).Server.LAddr),
		serve.WithVAddr(envs.Instance().Get("configs").(*configs.Config).Server.VAddr),
	)

	return nil
}
