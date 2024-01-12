package apps

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/configs"
	"github.com/yamakiller/velcro-go/utils/files"
	"github.com/yamakiller/velcro-go/vlog"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var appName string = "test-client"

type Program struct {
	stopper        chan struct{}
	stopped        sync.WaitGroup
	connSuccess    int32
	connFailed     int32
	sendSuccess    int32
	sendFailed     int32
	recviceSuccess int32
	recviceFailed  int32
	sendBytes      int64
	recvBytes      int64
	success        int64
	failed         int64
	writeBytes     []byte
}

func (p *Program) Start(s service.Service) error {

	vlog.Info("[PROGRAM]", "client.test Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("configs", files.NewLocalPathFull("config.yaml"), &configs.Config{}); err != nil {
		vlog.Fatal("[PROGRAM]", "Failed to load environment variables", err)
		return err
	}


	// if err := p.service.Start(); err != nil {
	// 	vlog.Info("[PROGRAM]", "client.test Failed to start network service", err)
	// 	return err
	// }
	vlog.Info("[PROGRAM]", "client.test Start network service completed")
	Test()
	return nil
}

func getMessageTypeFromTypeURL(typeURL string) (protoreflect.Message, error) {
	// 解析 typeURL

	// 获取类型
	messageType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(typeURL))
	if err != nil {
		return nil, fmt.Errorf("Type %s not found", typeURL)
	}
	return messageType.New(), nil
}

func (p *Program) Stop(s service.Service) error {
	close(p.stopper)
	p.stopped.Wait()
	return nil
}

func (p *Program) isStoped() bool {
	select {
	case <-p.stopper:
		return true
	default:
		return false
	}
}

func (p *Program) spawnClient(config *configs.Config, out []byte) {

	p.stopped.Add(1)
	go func() {
		defer p.stopped.Done()
		var (
			temp [256]byte
		)
		conn, err := net.DialTimeout("tcp", config.TargetAddr, time.Duration(config.ClientConnectionTimeout)*time.Millisecond)
		if err != nil {

			atomic.AddInt32(&p.connFailed, 1)
			atomic.AddInt64(&p.failed, 1)
			return
		}
		for {

			if p.isStoped() {
				break
			} else {
				time.Sleep(time.Duration(config.IntervalSecond) * time.Millisecond)
			}

			atomic.AddInt32(&p.connSuccess, 1)

			if _, err := conn.Write(out); err != nil {
				conn.Close()
				conn = nil
				atomic.AddInt32(&p.sendFailed, 1)
				atomic.AddInt64(&p.failed, 1)
				return
			}

			atomic.AddInt32(&p.sendSuccess, 1)

			n, err := conn.Read(temp[:])
			if err != nil {
				conn.Close()
				conn = nil

				atomic.AddInt32(&p.recviceFailed, 1)
				atomic.AddInt64(&p.failed, 1)
				return
			}

			// conn.Close()
			// conn = nil

			atomic.AddInt64(&p.recvBytes, int64(n))
			atomic.AddInt32(&p.recviceSuccess, 1)
			atomic.AddInt64(&p.success, 1)
		}

	}()
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

