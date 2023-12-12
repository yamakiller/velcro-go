package application

import (
	"os"

	"github.com/kardianos/service"
	"github.com/pkg/errors"
)

type Guardian struct {
	Name        string
	Display     string
	Description string
}

func (g *Guardian) Startup(app service.Interface) (string, error) {
	scfg := &service.Config{
		Name:        g.Name,
		DisplayName: g.Display,
		Description: g.Description,
	}

	s, err := service.New(app, scfg)
	if err != nil {
		return "", err
	}

	if len(os.Args) > 1 {
		serviceAction := os.Args[1]
		switch serviceAction {
		case "install":
			err := s.Install()
			if err != nil {
				return "", errors.Errorf("service install fail[error:%s]", err.Error())
			}
			return "service install success", nil
		case "uninstall":
			err := s.Uninstall()
			if err != nil {
				return "", errors.Errorf("service uninstall fail[error:%s]", err.Error())
			}
			return "service uninstall success", nil
		case "start":
			err := s.Start()
			if err != nil {
				return "", errors.Errorf("service start fail[error:%s]", err.Error())
			}
			return "service start success", nil
		case "stop":
			err := s.Stop()
			if err != nil {
				return "", errors.Errorf("service stop fail[error:%s]", err.Error())
			}
			return "service stop success", nil
		}
	}

	return "", s.Run()
}
