package apps

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/tcpclient"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	"github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
)

var index = int32(0)

func Test() {
	for i := int32(0); i < envs.Instance().Get("configs").(*configs.Config).ClientNumber; i++ {
		go clientRun(atomic.AddInt32(&index, 1))
	}
}

func owner(cli *tcpclient.Conn, i int32) (string, string) {
	uid := singin(cli, fmt.Sprintf("test_00%d&123456", i))
	spaceid := createbattlespace(cli)
	return uid, spaceid
}
func user(cli *tcpclient.Conn, i int32, spaceid string) {
	uid := singin(cli, fmt.Sprintf("test_00%d&123456", i))
	enterbattlespace(cli, spaceid)
	readybattlespace(cli, uid, spaceid, true)
}

func NewConn() *tcpclient.Conn {
	cli := tcpclient.NewConn()
	cli.Dial(envs.Instance().Get("configs").(*configs.Config).TargetAddr, 2*time.Second)
	return cli
}
func NewUDPConn() *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp", envs.Instance().Get("configs").(*configs.Config).TargetAddr)
	if err != nil {
		fmt.Println("Err resolve UDP address: ", err)
		return nil
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Dial UDP error: ", err)
		return nil
	}
	return conn
}
func sendUDP(conn *net.UDPConn, uid, spaceid, code string) {
	req := &pubs.ReportNatClient{
		BattleSpaceID: spaceid,
		VerifiyCode:   code,
	}
	b, _ := proto.Marshal(req)
	buff := make([]byte, 1500)
	offset := 0
	buff[offset] = byte(len(uid))
	offset += 1
	copy(buff[offset:], uid)
	offset += len(uid)
	binary.BigEndian.PutUint16(buff[offset:offset+2], uint16(len(b)))
	offset += 2
	copy(buff[offset:], b)
	offset += len(b)
	conn.Write(buff[:offset])
}

func clientRun(i int32) {
	cli1 := NewConn()
	_, spaceid := owner(cli1, i)
	if spaceid != "" {
		// udpConn := NewUDPConn()
		// sendUDP(udpConn,uid,spaceid,"123456")
		getlist(cli1)
	}

	// cli2 := NewConn()
	// t1 := time.NewTicker(time.Millisecond * 500)
	// for {
	// 	select {
	// 	case <-t1.C:
	// 		if spaceid != ""{
	// 			user(cli2,i + envs.Instance().Get("configs").(*configs.Config).ClientNumber,spaceid )
	// 			spaceid = ""
	// 		}
	// 	}
	// }
}

func singin(cp *tcpclient.Conn, token string) string {
	req := &mpubs.SignIn{
		Token: token,
	}
	res, err := cp.RequestMessage(req, 5000)
	if err != nil {
		vlog.Info("[PROGRAM]", "singin failed  ", err.Error())
		return ""
	}
	if res != nil {
		vlog.Info("[PROGRAM]", token, "  singin : ", res.(*mpubs.SignInResp).Uid)
		return res.(*mpubs.SignInResp).Uid
	}
	return ""
}

func createbattlespace(cp *tcpclient.Conn) string {
	req := &mpubs.CreateBattleSpace{
		MapURI:   "123456",
		MaxCount: 6,
	}
	res, err := cp.RequestMessage(req, 2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "createbattlespace failed  ", err.Error())
		return ""
	}
	if res != nil {
		fmt.Println("createbattlespace : ", res.(*mpubs.CreateBattleSpaceResp))
		return res.(*mpubs.CreateBattleSpaceResp).SpaceId
	}
	return ""
}

func enterbattlespace(cp *tcpclient.Conn, spaceid string) string {
	req := &mpubs.EnterBattleSpace{
		SpaceId: spaceid,
	}
	res, err := cp.RequestMessage(req, 8000)
	if err != nil {
		vlog.Info("[PROGRAM]", "enterbattlespace failed  ", err.Error())
		return ""
	}
	if res != nil {
		fmt.Println("enterbattlespace : ", res.(*mpubs.EnterBattleSpaceResp))
		return res.(*mpubs.EnterBattleSpaceResp).Space.SpaceId
	}
	return ""
}

func readybattlespace(cp *tcpclient.Conn, uid, spaceid string, ready bool) bool {
	req := &mpubs.ReadyBattleSpace{
		SpaceId: spaceid,
		Uid:     uid,
		Ready:   ready,
	}
	res, err := cp.RequestMessage(req, 8000)
	if err != nil {
		vlog.Info("[PROGRAM]", "readybattlespace failed  ", err.Error())
		return false
	}
	if res != nil {
		fmt.Println("readybattlespace : ", res.(*mpubs.ReadyBattleSpaceResp))
		return res.(*mpubs.ReadyBattleSpaceResp).Ready
	}
	return false
}

func exitbattlespace(cp *tcpclient.Conn, spaceid string, uid string) {

	req := &mprvs.RequestExitBattleSpace{
		BattleSpaceID: spaceid,
		UID:           uid,
	}
	res, err := cp.RequestMessage(req, 8000)
	if err != nil {
		vlog.Info("[PROGRAM]", "exitbattlespace failed  ", err.Error())
		return
	}
	if res != nil {
		fmt.Println("exitbattlespace : ", res.(*mprvs.RequestExitBattleSpace))
	}
}

func getlist(cp *tcpclient.Conn) {
	req := &mpubs.GetBattleSpaceList{
		Start: 0,
		Size:  10,
	}
	res, err := cp.RequestMessage(req, 2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "getlist failed  ", err.Error())
		return
	}
	if res != nil {
		fmt.Println("getlist : ", res.(*mpubs.GetBattleSpaceListResp).Count)
	}
}
