package apps

import (
	"github.com/yamakiller/velcro-go/cluster/gateway"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/gateway.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/gateway.service/rds"
	"github.com/yamakiller/velcro-go/utils/encryption/ecdh"
	_ "github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	_ "github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	_"github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	_"github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
)

type gatewayService struct {
	gwy     *gateway.Gateway
}

func (gs *gatewayService) Start() error {


	rds.WithAddr(envs.Instance().Get("configs").(*configs.Config).Redis.Addr)
	rds.WithPwd(envs.Instance().Get("configs").(*configs.Config).Redis.Pwd)
	rds.WithDialTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Dial)
	rds.WithReadTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Read)
	rds.WithWriteTimeout(envs.Instance().Get("configs").(*configs.Config).Redis.Timeout.Write)

	if err := rds.Connection(); err != nil {
		return err
	}

	gs.gwy = gateway.New(
		gateway.WithLAddr(envs.Instance().Get("configs").(*configs.Config).Server.LAddr),
		gateway.WithVAddr(envs.Instance().Get("configs").(*configs.Config).Server.VAddr),
		gateway.WithLAddrServant(envs.Instance().Get("configs").(*configs.Config).Server.LAddrServant),
		gateway.WithRoute(&envs.Instance().Get("configs").(*configs.Config).Router),
		gateway.WithNewEncryption(gs.newEncryption),
	)

	if err := gs.gwy.Start(); err != nil {
		return err
	}

	return nil
}

func (gs *gatewayService) Stop() error {

	if gs.gwy != nil {
		gs.gwy.Stop()
		gs.gwy = nil
	}

	return nil
}

func (gs *gatewayService) newEncryption() *gateway.Encryption {
	if !envs.Instance().Get("configs").(*configs.Config).Server.EncryptionEnabled {
		return nil
	}

	return &gateway.Encryption{Ecdh: &ecdh.Curve25519{A: 247, B: 127, C: 64}}
}