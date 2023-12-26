package client

import (
	"github.com/yamakiller/velcro-go/cluster/service"
)

func New()*LoginService{
	s := &LoginService{}
	s.Service = service.New(service.WithPool(NewLoginServiceClientPool(s)))
	return s
}

type LoginService struct {
	*service.Service
}