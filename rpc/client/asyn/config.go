package asyn

type ConnConfig struct {
	Kleepalive int32

	Connected func()
	Closed    func()
}

func defaultConfig() *ConnConfig {
	return &ConnConfig{
		Kleepalive: 10 * 1000,
	}
}
