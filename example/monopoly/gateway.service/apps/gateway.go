package apps

import (
	"context"
	"encoding/binary"
	"net"
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/cluster/gateway"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/gateway.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/gateway.service/rds"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/utils/encryption"
	"github.com/yamakiller/velcro-go/utils/encryption/ecdh"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type gatewayService struct {
	gwy     *gateway.Gateway
	udp     *net.UDPConn
	udpWait sync.WaitGroup
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
		gateway.WithLAddr(envs.Instance().Get("configs").(*configs.Config).Server.LAddr),
		gateway.WithVAddr(envs.Instance().Get("configs").(*configs.Config).Server.VAddr),
		gateway.WithLAddrServant(envs.Instance().Get("configs").(*configs.Config).Server.LAddrServant),
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
/*
**************************UDP 消息数据结构***总长度不可小于30字节***********************************
* |-1字节 用户ID长度-|--------用户ID---------------|-2字节 大端 数据长度-|--------pubs.ReportNatClient 数据--数据大小 小于128字节-------------------
*****************************************************************
*/
func (gs *gatewayService) udpLoop() {
	defer gs.udpWait.Done()
	var temp [1500]byte
	for {
		n, addr, err := gs.udp.ReadFromUDP(temp[:])
		if err != nil {
			break
		}

		if n < 30 {
			// TODO: 可以怀疑有人攻击
			vlog.Errorf("udp n fail[error:%d]", n)
			continue
		}
		offset := 1
		id := string(temp[offset : temp[0]+1])
		offset += int(temp[0])
		dLen := int(binary.BigEndian.Uint16(temp[offset : offset+2]))
		offset += 2
		if (dLen) > n || dLen > 128 {
			vlog.Errorf("udp binary.BigEndian.Uint16 fail[error:%d]", dLen)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		clientID := rds.GetPlayerClientID(ctx, id)
		if clientID == nil {
			continue
		}
		client := gs.gwy.GetClient(clientID)
		if client == nil {
			// TODO: 可以怀疑有人攻击
			vlog.Errorf("udp client fail[error:%s]", id)
			continue
		}

		request := mpubs.ReportNatClient{}
		// 设定生存周期
		{
			defer gs.gwy.ReleaseClient(client)
			if client.Secret() != nil {
				decrypt, err := encryption.AesDecryptByGCM(temp[offset:offset+dLen], client.Secret())
				if err != nil {
					vlog.Errorf("udp decrypt fail[error:%s]", err.Error())
					continue
				}
				dLen = copy(temp[:len(decrypt)], decrypt)
			} else {
				dLen = copy(temp[:dLen], temp[offset:offset+dLen])
			}

			if err := proto.Unmarshal(temp[:dLen], &request); err != nil {
				vlog.Errorf("udp unmarshal protobuff fail[error:%s]", err.Error())
				continue
			}
		}

		postRequest := &mprvs.ReportNat{BattleSpaceID: request.BattleSpaceID,
			VerifiyCode: request.VerifiyCode,
			NatAddr:     addr.AddrPort().String(),
		}
		bodyAny, err := anypb.New(postRequest)
		if err != nil {
			vlog.Warnf("%s message encoding failed error %s",
				string(protoreflect.FullName(proto.MessageName(postRequest))), err.Error())
			continue
		}
		// 查找目标路由
		r := gs.gwy.FindRouter(postRequest)
		if r == nil {
			vlog.Warn("protocols.ReportNat message unfound router")
			continue
		}
		forwardBundle := &prvs.ForwardBundle{
			Sender: clientID,
			Body:   bodyAny,
		}
		// 推送到目标服务
		if _, err := r.Proxy.RequestMessage(forwardBundle, 2000); err != nil {
			vlog.Errorf("protocols.ReportNat post message fail %s", err.Error())
			continue
		}
	}
}