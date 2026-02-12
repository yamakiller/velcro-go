package accounts

import (
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/locals"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
	// "github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/steam"
)

var (
	signHandle sign.Sign
)

func init() {
	// signHandle = &steam.SteamSign{}
	signHandle = &locals.LocalSign{}
}
