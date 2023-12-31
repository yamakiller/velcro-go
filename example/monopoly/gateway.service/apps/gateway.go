package apps

import (
	"encoding/binary"
	"net"
	"sync"

	"github.com/yamakiller/velcro-go/cluster/gateway"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/gateway.service/configs"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/utils/encryption"
	"github.com/yamakiller/velcro-go/utils/encryption/ecdh"
	"google.golang.org/protobuf/proto"
)

type gatewayService struct {
	gwy     *gateway.Gateway
	udp     *net.UDPConn
	udpWait sync.WaitGroup
}

func (gs *gatewayService) Start(logAgent logs.LogAgent) error {

	udpAddr, err := net.ResolveUDPAddr("udp", envs.Instance().Get("configs").(*configs.Config).Server.LAddr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	gs.udp = udpConn

	gs.gwy = gateway.New(
		gateway.WithLoggerAgent(logAgent),
		gateway.WithLAddr(envs.Instance().Get("configs").(*configs.Config).Server.LAddr),
		gateway.WithVAddr(envs.Instance().Get("configs").(*configs.Config).Server.VAddr),
		gateway.WithRoute(&envs.Instance().Get("configs").(*configs.Config).Router),
		gateway.WithNewEncryption(gs.newEncryption),
	)

	if err := gs.gwy.Start(); err != nil {
		gs.udp.Close()
		gs.udp = nil
		return err
	}

	gs.udpWait.Add(1)
	go gs.udpLoop()

	return nil
}

func (gs *gatewayService) Stop() error {
	if gs.udp != nil {
		gs.udp.Close()
		gs.udpWait.Wait()
		gs.udp = nil
	}

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

func (gs *gatewayService) udpLoop() {
	defer gs.udpWait.Done()
	var temp [1500]byte
	for {
		n, addr, err := gs.udp.ReadFromUDP(temp[:])
		if err != nil {
			break
		}

		if n < 38 {
			// TODO: 可以怀疑有人攻击
			continue
		}

		id := string(temp[:36])
		dLen := int(binary.BigEndian.Uint16(temp[36:38]))
		if (dLen+38) > n || dLen > 128 {
			continue
		}

		clientID := gs.gwy.NewClientID(id)
		client := gs.gwy.GetClient(clientID)
		if client == nil {
			// TODO: 可以怀疑有人攻击
			continue
		}

		msg := mpubs.ReportNatClient{}
		// 设定生存周期
		{
			defer gs.gwy.ReleaseClient(client)
			if client.Secret() != nil {
				decrypt, err := encryption.AesDecryptByGCM(temp[38:38+dLen], client.Secret())
				if err != nil {
					gs.gwy.System.Error("udp decrypt fail[error:%s]", err.Error())
					continue
				}
				dLen = copy(temp[:len(decrypt)], decrypt)
			}

			if err := proto.Unmarshal(temp[:dLen], &msg); err != nil {
				gs.gwy.System.Error("udp unmarshal protobuff fail[error:%s]", err.Error())
				continue
			}
		}

		postMsg := &mprvs.ReportNat{RoomID: msg.RoomID,
			VerifiyCode: msg.VerifiyCode,
			NatAddr:     addr.AddrPort().String()}

		// 查找目标路由
		r := gs.gwy.FindRouter(postMsg)
		if r == nil {
			gs.gwy.System.Warning("protocols.ReportNat message unfound router")
			continue
		}

		// 推送到目标服务
		if _, err := r.Proxy.RequestMessage(postMsg, 2000); err != nil {
			gs.gwy.System.Error("protocols.ReportNat post message fail %s", err.Error())
			continue
		}
	}
}
