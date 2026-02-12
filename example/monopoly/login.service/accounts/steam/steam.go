package steam

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/errs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
)

type SteamSign struct {
}

func (ls *SteamSign) Init() error {
	return nil
}

// In token username&password
func (ls *SteamSign) In(ctx context.Context, token string) (*sign.Account, error) {
	inarray := strings.Split(token, "&")
	if len(inarray) != 2 {
		return nil, errs.ErrSignAccountOrPass
	}

	// test_001-00n
	accounts := strings.Split(inarray[0], "_")
	if len(accounts) != 2 {
		return nil, errs.ErrSignAccountOrPass
	}

	if accounts[0] == "steam" {
		if len(accounts[1]) != 17 {
			return nil, errs.ErrSignAccountOrPass
		}
		result := &sign.Account{
			UID:         accounts[1],
			DisplayName: fmt.Sprintf("steam%s", accounts[1]),
			Rule:        3,
			Externs:     map[string]string{},
		}

		return result, nil
	} else if accounts[0] == "test" {
		sn, err := strconv.ParseInt(accounts[1], 10, 32)
		if err != nil {
			return nil, errs.ErrSignAccountOrPass
		}

		result := &sign.Account{
			UID:         accounts[1],
			DisplayName: fmt.Sprintf("米奇%d", sn),
			Rule:        3,
			Externs:     map[string]string{},
		}
		return result, nil
	}
	return nil, errs.ErrSignAccountOrPass
}

func (ls *SteamSign) Out(token string) error {
	return nil
}
