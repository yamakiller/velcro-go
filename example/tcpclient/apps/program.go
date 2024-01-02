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
	"github.com/yamakiller/velcro-go/example/tcpclient/configs"
	"github.com/yamakiller/velcro-go/vlog"
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

	cfgFilePath, err := p.GetLocalConfigFilePath()
	if err != nil {
		vlog.Errorf("load config fail:[error:%s]", err.Error())

		return err
	}

	config := configs.Config{}
	envs.With(config.IEnv())
	if err := envs.Instance().Load("config", cfgFilePath, &config); err != nil {
		vlog.Errorf("Load %s config file fail-%s", cfgFilePath, err.Error())
		return err
	}

	p.writeBytes = []byte(config.PostData)
	p.stopper = make(chan struct{})
	p.stopped.Add(1)
	go func() {
		defer p.stopped.Done()
		for i := 0; i < config.ClientNumber; i++ {
			p.spawnClient(&config, p.writeBytes)
		}

		for {
			if p.isStoped() {
				break
			}

			var print string = ""
			print += fmt.Sprintf("\r 连接成功次数:%d\n", atomic.LoadInt32(&p.connSuccess))
			print += fmt.Sprintf("\r 连接失败次数:%d\n", atomic.LoadInt32(&p.connFailed))
			print += fmt.Sprintf("\r 发送数据成功次数:%d\n", atomic.LoadInt32(&p.sendSuccess))
			print += fmt.Sprintf("\r 发送数据失败次数:%d\n", atomic.LoadInt32(&p.sendFailed))
			print += fmt.Sprintf("\r 接收数据成功次数:%d\n", atomic.LoadInt32(&p.recviceSuccess))
			print += fmt.Sprintf("\r 接收数据失败次数:%d\n", atomic.LoadInt32(&p.recviceFailed))
			print += fmt.Sprintf("\r 发送字节数:%d 字节\n", atomic.LoadInt64(&p.sendBytes))
			print += fmt.Sprintf("\r 接收字节数:%d 字节\n", atomic.LoadInt64(&p.recvBytes))
			print += fmt.Sprintf("\r 总成功数:%d\n", atomic.LoadInt64(&p.success))
			print += fmt.Sprintf("\r 总失败数:%d\n", atomic.LoadInt64(&p.failed))

			fmt.Printf("%s", print)

			time.Sleep(time.Duration(config.ScreenRefreshFrequency) * time.Millisecond)

		}
	}()

	return nil
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

		for {

			if p.isStoped() {
				break
			} else {
				time.Sleep(time.Duration(config.IntervalSecond) * time.Millisecond)
			}

			conn, err := net.DialTimeout("tcp", config.TargetAddr, time.Duration(config.ClientConnectionTimeout)*time.Millisecond)
			if err != nil {

				atomic.AddInt32(&p.connFailed, 1)
				atomic.AddInt64(&p.failed, 1)
				continue
			}

			atomic.AddInt32(&p.connSuccess, 1)

			if _, err := conn.Write(out); err != nil {
				conn.Close()
				conn = nil
				atomic.AddInt32(&p.sendFailed, 1)
				atomic.AddInt64(&p.failed, 1)
				continue
			}

			atomic.AddInt32(&p.sendSuccess, 1)

			n, err := conn.Read(temp[:])
			if err != nil {
				conn.Close()
				conn = nil

				atomic.AddInt32(&p.recviceFailed, 1)
				atomic.AddInt64(&p.failed, 1)
				continue
			}

			conn.Close()
			conn = nil

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
