package apps

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/configs"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/internet"
	clientpool "github.com/yamakiller/velcro-go/rpc/client/connpool"
)

var appName string = "test-rpc-client"

type Program struct {
	_c *clientpool.ConnectPool
}

func (p *Program) Start(s service.Service) error {

	cfgFilePath, err := p.GetLocalConfigFilePath()
	if err != nil {
		return err
	}

	config := configs.Config{}
	envs.With(config.IEnv())
	if err := envs.Instance().Load("config", cfgFilePath, &config); err != nil {
		return err
	}
	p._c = clientpool.NewConnectPool(config.TargetAddr, clientpool.IdleConfig{
		Closed:internet.Closed,
	})
	t1 := time.NewTicker(time.Second * 1)
	for {
		select{
		case <-t1.C:
			p._c.RequestMessage(&protos.Auth{Msg: "123456"})
		default:

		}
	}

	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p._c != nil {
		p._c.Shudown()
		p._c = nil
	}
	return nil
}
func (p *Program) GetLocalConfigFilePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	exPath := filepath.Dir(ex)
	cfgFilePath := filepath.Join(exPath, "config.yml")

	return cfgFilePath, nil
}
