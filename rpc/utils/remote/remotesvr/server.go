package remotesvr

import (
	"context"
	"net"
	"sync"

	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/rpc/utils/endpoint"
	"github.com/yamakiller/velcro-go/rpc/utils/remote"
	"github.com/yamakiller/velcro-go/vlog"
)

// Server is the interface for remote server.
type Server interface {
	Start() chan error
	Stop() error
	Address() net.Addr
}

type server struct {
	opt      *remote.ServerOption
	listener net.Listener
	transSvr remote.TransServer

	inkHdlFunc endpoint.Endpoint
	sync.Mutex
}

// NewServer creates a remote server.
func NewServer(opt *remote.ServerOption, inkHdlFunc endpoint.Endpoint, transHdlr remote.ServerTransHandler) (Server, error) {
	transSvr := opt.TransServerFactory.NewTransServer(opt, transHdlr)
	s := &server{
		opt:        opt,
		inkHdlFunc: inkHdlFunc,
		transSvr:   transSvr,
	}
	return s, nil
}

// Start starts the server and return chan, the chan receive means server shutdown or err happen
func (s *server) Start() chan error {
	errCh := make(chan error, 1)
	ln, err := s.buildListener()
	if err != nil {
		errCh <- err
		return errCh
	}

	s.Lock()
	s.listener = ln
	s.Unlock()

	gofunc.GoFunc(context.Background(), func() { errCh <- s.transSvr.BootstrapServer(ln) })
	return errCh
}

func (s *server) buildListener() (ln net.Listener, err error) {
	if s.opt.Listener != nil {
		vlog.Infof("VELCOR: server listen at addr=%s", s.opt.Listener.Addr().String())
		return s.opt.Listener, nil
	}
	addr := s.opt.Address
	if ln, err = s.transSvr.CreateListener(addr); err != nil {
		vlog.Errorf("VELCOR: server listen failed, addr=%s error=%s", addr.String(), err)
	} else {
		vlog.Infof("VELCOR: server listen at addr=%s", ln.Addr().String())
	}
	return
}

// Stop stops the server gracefully.
func (s *server) Stop() (err error) {
	s.Lock()
	defer s.Unlock()
	if s.transSvr != nil {
		err = s.transSvr.Shutdown()
		s.listener = nil
	}
	return
}

func (s *server) Address() net.Addr {
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}