package apps

import (
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
)

type battleService struct {
	battle *serve.Servant
}

func (bs *battleService) Start() error {

	rds.WithAddr(envs.Instance().Get("configs").(*configs.Config).Redis.Addr)
	rds.WithPwd(envs.Instance().Get("configs").(*configs.Config).Redis.Pwd)
	rds.WithDialTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Dial)
	rds.WithReadTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Read)
	rds.WithWriteTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Write)

	if err := rds.Connection(); err != nil {
		return err
	}

	bs.battle = serve.New(
		serve.WithProducerActor(bs.newBattleActor),
		serve.WithName("BattleService"),
		serve.WithLAddr(envs.Instance().Get("configs").(*configs.Config).Server.LAddr),
		serve.WithVAddr(envs.Instance().Get("configs").(*configs.Config).Server.VAddr),
		serve.WithKleepalive(int32(envs.Instance().Get("configs").(*configs.Config).Server.Kleepalive)),
		serve.WithRoute(&envs.Instance().Get("configs").(*configs.Config).Router),
	)

	if err := bs.battle.Start(); err != nil {
		rds.Disconnect()
		return err
	}

	return nil
}

func (bs *battleService) Stop() error {

	if bs.battle != nil {
		bs.battle.Stop()
		bs.battle = nil
	}

	rds.Disconnect()

	return nil
}

func (bs *battleService) newBattleActor(conn *serve.ServantClientConn) serve.ServantClientActor {
	actor := &BattleActor{ancestor: bs.battle}

	conn.Register(&mpubs.CreateBattleSpace{}, actor.onCreateBattleSpace)
	conn.Register(&mpubs.GetBattleSpaceList{}, actor.onGetBattleSpaceList)
	conn.Register(&mpubs.EnterBattleSpace{}, actor.onEnterBattleSpace)
	conn.Register(&mpubs.ReadyBattleSpace{}, actor.onReadyBattleSpace)
	conn.Register(&mpubs.RequsetStartBattleSpace{}, actor.onRequestStartBattleSpace)

	conn.Register(&mprvs.ReportNat{}, actor.onReportNat)
	conn.Register(&mprvs.RequestExitBattleSpace{}, actor.onRequestExitBattleSpace)
	return actor
}
