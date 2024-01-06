package apps

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/configs"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/internet"
	"github.com/yamakiller/velcro-go/rpc/client/clientpool"
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
		Closed: internet.Closed,
	})

	index := int32(0)
	for i := 0; i < 5; i++ {
		go func() {
			id := atomic.AddInt32(&index, 1)
			t1 := time.NewTicker(time.Millisecond * 300)
			for {
				select {
				case <-t1.C:
					p._c.RequestMessage(&protos.Auth{Msg: fmt.Sprintf("test-00%d", id)}, 2000)
				default:
				}
			}
		}()
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
