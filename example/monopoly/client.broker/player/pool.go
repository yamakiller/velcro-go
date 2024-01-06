package player

import (
	"reflect"

	"github.com/yamakiller/magicLibs/actors"
	"github.com/yamakiller/magicLibs/boxs"
	spubs "github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
)
var (
	_core *actors.Core
)

func StartUp(core *actors.Core){
	_core = core
}

func NewPlayer(id string) *Player {
	c := &Player{
		Box: *boxs.SpawnBox(nil),
	}
	_core.New(func(pid *actors.PID) actors.Actor {
		c.Box.WithPID(pid)
		c.Box.Register(reflect.TypeOf(&spubs.PingMsg{}), c.onPingMsg)
		c.Box.Register(reflect.TypeOf(&mpubs.SignInResp{}), c.onSignInResp)
		c.Box.Register(reflect.TypeOf(&mpubs.GetBattleSpaceListResp{}), c.onGetBattleSpaceListResp)

		return &c.Box
	},0)
	return c
}
