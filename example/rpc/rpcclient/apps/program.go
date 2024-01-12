package apps

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/envs"

	// "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/configs"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/internet"
	"github.com/yamakiller/velcro-go/rpc/client"
	"github.com/yamakiller/velcro-go/rpc/client/clientpool"
	// "github.com/yamakiller/velcro-go/rpc/messages"
	// "github.com/yamakiller/velcro-go/vlog"
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
	cli := client.NewConn()
	if err := cli.Dial(config.TargetAddr,2*time.Second);err != nil{
		fmt.Fprintln(os.Stderr,err.Error())
		return err
	}
	// index := int32(0)
	// for i := 0; i < 1; i++ {
	// 	go func() {
	// 		// id := atomic.AddInt32(&index, 1)
	// 		t1 := time.NewTicker(time.Millisecond * 300)
	// 		for {
	// 			select {
	// 			case <-t1.C:
	// 				// req := &pubs.SignIn{
	// 				// 	Token: "test_001&123456",
	// 				// }
	// 				cli.(*client.Conn).OnPing(&messages.RpcPingMessage{VerifyKey: 10})
	// 				// res,_ :=	p._c.RequestMessage(req, 2000)
	// 				// if res != nil && res.Result() != nil{
	// 				// 	vlog.Info("RequestMessage : ",res.Result())
	// 				// }
	// 			default:
	// 			}
	// 		}
	// 	}()
	// }

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
