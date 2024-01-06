package player

import (
	"fmt"

	"github.com/yamakiller/magicLibs/boxs"
)

type Player struct {
	boxs.Box
}

func (p *Player) onPingMsg(ctx *boxs.Context){
	fmt.Println("onPingMsg")
}

func (p *Player) onSignInResp(ctx *boxs.Context){
	fmt.Println("onSignInResp")
}
func (p *Player) onGetBattleSpaceListResp(ctx *boxs.Context){
	fmt.Println("onGetBattleSpaceListResp")
}


