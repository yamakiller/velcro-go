package classes

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
)

type ClientGroup interface {
	Register(cid *network.ClientID, l Client)
	UnRegister(cid *network.ClientID)
	GetClient(cid *network.ClientID)
	ReleaseLinkerOwner(client Client)
	Clear()
}

type DefaultClientGroup struct {
	clients map[network.CIDKEY]Client
	mu      sync.Mutex
}

func (lg *DefaultClientGroup) Register(cid *network.ClientID, l Client) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	_, ok := lg.clients[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	l.referenceIncrement()
	lg.clients[network.Key(cid)] = l
}

func (lg *DefaultClientGroup) UnRegister(cid *network.ClientID) {

	lg.mu.Lock()
	l, ok := lg.clients[network.Key(cid)]
	if !ok {
		lg.mu.Unlock()
		return
	}

	ref := l.referenceDecrement()
	delete(lg.clients, network.Key(cid))

	lg.mu.Unlock()

	if ref == 0 {
		lg.free(l)
	}
}

func (lg *DefaultClientGroup) GetClient(cid *network.ClientID) Client {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	l, ok := lg.clients[network.Key(cid)]
	if !ok {
		return nil
	}

	l.referenceIncrement()
	return l
}

func (lg *DefaultClientGroup) ReleaseLinker(linker Client) {
	lg.mu.Lock()
	ref := linker.referenceDecrement()
	lg.mu.Unlock()

	if ref == 0 {
		lg.free(linker)
	}
}

func (lg *DefaultClientGroup) Clear() {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	for _, l := range lg.clients {
		l.ClientID().UserClose()
	}
}

func (lg *DefaultClientGroup) free(l Client) {
	l.Destory()
}
