package serve

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/yamakiller/velcro-go/network"
)

var (
	messageEnvelopePool       sync.Pool
	enablePoolMessageEnvelope int32 = 1
)

func init() {
	// allow disabling by env without modifying the code and recompiling
	if os.Getenv("VELCRO_DISABLE_MESSAGE_ENVELOPE_POOL") != "" {
		EnablePoolMessageEnvelope(false)
	}
}

func EnablePoolMessageEnvelope(enable bool) {
	if enable {
		atomic.StoreInt32(&enablePoolMessageEnvelope, 1)
	} else {
		atomic.StoreInt32(&enablePoolMessageEnvelope, 0)
	}
}

func PoolEnabledMessageEnvelope() bool {
	return atomic.LoadInt32(&enablePoolMessageEnvelope) == 1
}

type MessageEnvelope struct {
	sequenceId int32
	sender     network.ClientID
	message    interface{}
}

func (me *MessageEnvelope) zero() {
	me.sequenceId = -1
	me.sender.Address = ""
	me.sender.Id = ""
	me.message = nil
}

func (me *MessageEnvelope) Recycle() {
	if !PoolEnabledMessageEnvelope() {
		return
	}

	me.zero()
	messageEnvelopePool.Put(me)
}

func init() {
	messageEnvelopePool.New = newMessageEnvelope
}

func NewMessageEnvelopePool(seqId int32, clientId *network.ClientID, message interface{}) *MessageEnvelope {
	m := messageEnvelopePool.Get().(*MessageEnvelope)
	m.sequenceId = seqId
	if clientId != nil {
		m.sender.Address = clientId.Address
		m.sender.Id = clientId.Id
	}
	m.message = message

	return m
}

func newMessageEnvelope() interface{} {
	return &MessageEnvelope{}
}

func UnwrapEnvelopeSequenceId(message interface{}) (int32, bool) {
	if env, ok := message.(*MessageEnvelope); ok {
		return env.sequenceId, true
	}
	return -1, false
}

func UnwrapEnvelopeSender(message interface{}) (*network.ClientID, bool) {
	if env, ok := message.(*MessageEnvelope); ok {
		if env.sender.Id == "" {
			return nil, false
		}
		return &env.sender, true
	}
	return nil, false
}

func UnwrapEnvelopeMessage(message interface{}) interface{} {
	if env, ok := message.(*MessageEnvelope); ok {
		return env.message
	}
	return message
}
