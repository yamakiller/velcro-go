package accounts

import (
	"context"

	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
)

func Init() error {
	return signHandle.Init()
}

func SignIn(ctx context.Context, token string) (*sign.Account, error) {
	return signHandle.In(ctx, token)
}

func SignOut(token string) error {
	return signHandle.Out(token)
}
