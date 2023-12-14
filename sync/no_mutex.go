package sync

type NoMutex struct {
}

func (m *NoMutex) Lock() {

}

func (m *NoMutex) Unlock() {

}
