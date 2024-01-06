package player

import "sync"

var (
	_instance *Manager
	_once   sync.Once
)

func Instance() *Manager {
	_once.Do(func() {
		_instance = &Manager{
			_player: make(map[string]*Player),
		}
	})
	return _instance
}

type Manager struct{
	_player map[string]*Player
	_m sync.Mutex
}
func (m *Manager) Register(addr string, p *Player) {
	m._m.Lock()
	defer m._m.Unlock()
	if _,ok := m._player[addr]; ok {
		return
	}
	m._player[addr] = p
}
func (m *Manager) GetPlayer(addr string) *Player {
	m._m.Lock()
	defer m._m.Unlock()
	if _,ok := m._player[addr]; !ok {
		return nil
	}
	return m._player[addr]
}
func (m *Manager) RemovePlayer(addr string){
	m._m.Lock()
	defer m._m.Unlock()
	delete(m._player, addr)
}