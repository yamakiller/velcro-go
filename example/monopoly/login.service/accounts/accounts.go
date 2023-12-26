package accounts

import (
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
)


func Init() error {
	return signHandle.Init()
}


func SignIn(token string) (*sign.Account, error) {
	return signHandle.In(token)
}

func SignOut() error {
	return signHandle.Out()
} 