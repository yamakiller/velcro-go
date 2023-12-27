package locals

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/errs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
)

type LocalSign struct {
}

func (ls *LocalSign) Init() error {
	return nil
}

// In token username&password
func (ls *LocalSign) In(ctx context.Context, token string) (*sign.Account, error) {
	inarray := strings.Split(token, "&")
	if len(inarray) != 2 {
		return nil, errs.ErrSignAccountOrPass
	}

	// test_001-00n
	accounts := strings.Split(inarray[0], "_")
	if len(accounts) != 2 {
		return nil, errs.ErrSignAccountOrPass
	}

	sn, err := strconv.ParseInt(accounts[1], 10, 32)
	if err != nil {
		return nil, errs.ErrSignAccountOrPass
	}

	result := &sign.Account{
		UID:     fmt.Sprintf("t%d", sn),
		Externs: map[string]string{},
	}

	return result, nil
}

func (ls *LocalSign) Out() error {
	return nil
}
