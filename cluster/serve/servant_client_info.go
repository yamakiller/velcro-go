package serve

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/yamakiller/velcro-go/network"
)

var (
	clientInfoPool       sync.Pool
	enablePoolClientInfo int32 = 1
)

func init() {
	// allow disabling by env without modifying the code and recompiling
	if os.Getenv("VELCRO_DISABLE_SERVANT_CLIENT_INFO_POOL") != "" {
		EnablePoolClientInfo(false)
	}
}

func EnablePoolClientInfo(enable bool) {
	if enable {
		atomic.StoreInt32(&enablePoolClientInfo, 1)
	} else {
		atomic.StoreInt32(&enablePoolClientInfo, 0)
	}
}

// PoolEnabled returns true if pool is enabled.
func PoolEnabledClientInfo() bool {
	return atomic.LoadInt32(&enablePoolClientInfo) == 1
}

type ServantClientInfo struct {
	context network.Context
	message interface{}
}

func (s *ServantClientInfo) SeqId() int32 {
	cid, ok := UnwrapEnvelopeSequenceId(s.message)
	if !ok {
		return -1
	}
	return cid
}

func (s *ServantClientInfo) Sender() *network.ClientID {
	cid, ok := UnwrapEnvelopeSender(s.message)
	if !ok {
		return nil
	}
	return cid
}

func (s *ServantClientInfo) Message() interface{} {
	return UnwrapEnvelopeMessage(s.message)
}

func (s *ServantClientInfo) zero() {
	s.context = nil
	if s.message == nil {
		return
	}

	if v, ok := s.message.(*MessageEnvelope); ok {
		v.Recycle()
		s.message = nil
	} else {
		s.message = nil
	}
}

func (s *ServantClientInfo) Recycle() {
	if !PoolEnabledClientInfo() {
		return
	}

	s.zero()
	clientInfoPool.Put(s)
}

func init() {
	clientInfoPool.New = newClientInfo
}

func NewClientInfo(ctx network.Context, message interface{}) *ServantClientInfo {
	s := clientInfoPool.Get().(*ServantClientInfo)
	s.context = ctx
	s.message = message
	return s
}

func newClientInfo() interface{} {
	return &ServantClientInfo{}
}
