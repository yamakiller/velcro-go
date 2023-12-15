package classes

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
)

type GatewayLinkerGroup interface {
	Register(cid *network.ClientID, l GatewayLinker)
	UnRegister(cid *network.ClientID)
	GetLinker(cid *network.ClientID)
	ReleaseLinkerOwner(linker GatewayLinker)
	Clear()
}

type DefaultLinkerGroup struct {
	links map[network.CIDKEY]GatewayLinker
	mu    sync.Mutex
}

func (lg *DefaultLinkerGroup) Register(cid *network.ClientID, l GatewayLinker) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	_, ok := lg.links[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	l.referenceIncrement()
	lg.links[network.Key(cid)] = l
}

func (lg *DefaultLinkerGroup) UnRegister(cid *network.ClientID) {

	lg.mu.Lock()
	l, ok := lg.links[network.Key(cid)]
	if !ok {
		lg.mu.Unlock()
		return
	}

	ref := l.referenceDecrement()
	delete(lg.links, network.Key(cid))

	lg.mu.Unlock()

	if ref == 0 {
		lg.free(l)
	}
}

func (lg *DefaultLinkerGroup) GetLinker(cid *network.ClientID) GatewayLinker {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	l, ok := lg.links[network.Key(cid)]
	if !ok {
		return nil
	}

	l.referenceIncrement()
	return l
}

func (lg *DefaultLinkerGroup) ReleaseLinker(linker GatewayLinker) {
	lg.mu.Lock()
	ref := linker.referenceDecrement()
	lg.mu.Unlock()

	if ref == 0 {
		lg.free(linker)
	}
}

func (lg *DefaultLinkerGroup) Clear() {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	for _, l := range lg.links {
		l.ClientID().UserClose()
	}
}

func (lg *DefaultLinkerGroup) free(l GatewayLinker) {
	l.Destory()
}
