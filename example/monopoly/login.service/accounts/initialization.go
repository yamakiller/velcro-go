package accounts

import (
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/locals"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
)

var (
	signHandle sign.Sign
)

func init() {
	signHandle = &locals.LocalSign{}
}

