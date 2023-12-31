package apps

import (
	"github.com/yamakiller/velcro-go/cluster/serve"
	"google.golang.org/protobuf/proto"
)

type BattleActor struct {
	ancestor *serve.Servant
}

func (actor *BattleActor) onCreateBattleSpace(ctx *serve.ServantClientContext) (proto.Message, error) {
	return nil, nil
}

func (actor *BattleActor) onGetBattleSpaceList(ctx *serve.ServantClientContext) (proto.Message, error) {
	return nil, nil
}

func (actor *BattleActor) onEnterBattleSpace(ctx *serve.ServantClientContext) (proto.Message, error) {
	return nil, nil
}

func (actor *BattleActor) onReportNat(ctx *serve.ServantClientContext) (proto.Message, error) {
	return nil, nil
}

func (actor *BattleActor) onRequestExitBattleSpace(ctx *serve.ServantClientContext) (proto.Message, error) {
	return nil, nil
}

func (actor *BattleActor) Closed(ctx *serve.ServantClientContext) {

}
