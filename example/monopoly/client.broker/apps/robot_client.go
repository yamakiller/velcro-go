package apps

import (
	// "fmt"
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/tcpclient"
	"github.com/yamakiller/velcro-go/rpc/protocol"

	// mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/vlog"
)

var index = int32(0)

func Test() {
	for i := int32(0); i < envs.Instance().Get("configs").(*configs.Config).ClientNumber; i++ {
		go clientRun(atomic.AddInt32(&index, 1))
	}
}


func owner(cli *tcpclient.Conn,i int32) string{
	singin(cli,fmt.Sprintf("test_00%d&123456", i))
	return ""
	// return createbattlespace(cli)
}
// func user(cli *tcpclient.Conn,i int32,spaceid string){
// 	uid:=singin(cli,fmt.Sprintf("test_00%d&123456", i))
// 	enterbattlespace(cli,spaceid)
// 	readybattlespace(cli,uid,spaceid,true)
// }


func NewConn() *tcpclient.Conn{
	cli := tcpclient.NewConn()
	cli.Dial( envs.Instance().Get("configs").(*configs.Config).TargetAddr,2*time.Second)
	return cli
}

func clientRun(i int32) {
	cli1 := NewConn()
	
	// cli2 := NewConn()
	t1 := time.NewTicker(time.Millisecond * 3000)
	for {
		select {
		case <-t1.C:
			spaceid :=owner(cli1,i)
			if spaceid != ""{
			// 	user(cli2,i + envs.Instance().Get("configs").(*configs.Config).ClientNumber,spaceid )
			// 	spaceid = ""
			}
		}
	}
	for {

	}
}
type Response struct{
	cp *tcpclient.Conn
	seqId int32
}

func  (r *Response)Call(ctx context.Context, method string, args, result thrift.TStruct) (thrift.ResponseMeta, error){
	r.seqId++
	seqId := r.seqId
	oprot:= protocol.NewBinaryProtocol()
	defer oprot.Close()
	var err error
	r.cp.Register(method, result)
	if err := oprot.WriteMessageBegin(ctx, method,thrift.CALL, seqId); err != nil {
		return thrift.ResponseMeta{},nil
	}
	if err := args.Write(ctx, oprot); err != nil {
		return thrift.ResponseMeta{},nil
	}
	if err := oprot.WriteMessageEnd(ctx); err != nil {
		return thrift.ResponseMeta{},nil
	}
	result, err = r.cp.RequestMessage(oprot.GetBytes(), 5000)
	if err != nil {
		vlog.Info("[PROGRAM]", "singin failed  ", err.Error())
		return thrift.ResponseMeta{},nil
	}

	return thrift.ResponseMeta{},nil
}
func singin(cp *tcpclient.Conn, token string) string {
	c:= mpubs.NewLoginServiceClient(&Response{cp: cp})
	res ,_ := c.OnSignIn(context.Background(),&mpubs.SignIn{Token: token})
	if res == nil{
		return ""
	}
	fmt.Fprintf(os.Stderr,"uid %s\n",res.UID)
	return res.UID
}

// func createbattlespace(cp *tcpclient.Conn) string {
// 	req := &mpubs.CreateBattleSpace{
// 		MapURI:   "123456",
// 		MaxCount: 6,
// 	}
// 	res, err := cp.RequestMessage(req, 2000)
// 	if err != nil {
// 		vlog.Info("[PROGRAM]", "createbattlespace failed  ", err.Error())
// 		return ""
// 	}
// 	if res != nil {
// 		fmt.Println("createbattlespace : ", res.(*mpubs.CreateBattleSpaceResp))
// 		return res.(*mpubs.CreateBattleSpaceResp).SpaceId
// 	}
// 	return ""
// }

// func enterbattlespace(cp *tcpclient.Conn, spaceid string) string {
// 	req := &mpubs.EnterBattleSpace{
// 		SpaceId: spaceid,
// 	}
// 	res, err := cp.RequestMessage(req, 8000)
// 	if err != nil {
// 		vlog.Info("[PROGRAM]", "enterbattlespace failed  ", err.Error())
// 		return ""
// 	}
// 	if res!= nil {
// 		fmt.Println("enterbattlespace : ", res.(*mpubs.EnterBattleSpaceResp))
// 		return res.(*mpubs.EnterBattleSpaceResp).Space.SpaceId
// 	}
// 	return ""
// }

// func readybattlespace(cp *tcpclient.Conn, uid, spaceid string, ready bool) bool {
// 	req := &mpubs.ReadyBattleSpace{
// 		SpaceId: spaceid,
// 		Uid:     uid,
// 		Ready:   ready,
// 	}
// 	res, err := cp.RequestMessage(req, 8000)
// 	if err != nil {
// 		vlog.Info("[PROGRAM]", "readybattlespace failed  ", err.Error())
// 		return false
// 	}
// 	if res != nil {
// 		fmt.Println("readybattlespace : ", res.(*mpubs.ReadyBattleSpaceResp))
// 		return res.(*mpubs.ReadyBattleSpaceResp).Ready
// 	}
// 	return false
// }

// func exitbattlespace(cp *tcpclient.Conn, spaceid string, uid string) {

// 	req := &mprvs.RequestExitBattleSpace{
// 		BattleSpaceID: spaceid,
// 		UID:           uid,
// 	}
// 	res, err := cp.RequestMessage(req, 8000)
// 	if err != nil {
// 		vlog.Info("[PROGRAM]", "exitbattlespace failed  ", err.Error())
// 		return
// 	}
// 	if res != nil {
// 		fmt.Println("exitbattlespace : ", res.(*mprvs.RequestExitBattleSpace))
// 	}
// }

// func getlist(cp *tcpclient.Conn) {
// 	req := &mpubs.GetBattleSpaceList{
// 		Start: 0,
// 		Size:  10,
// 	}
// 	res, err := cp.RequestMessage(req, 2000)
// 	if err != nil {
// 		vlog.Info("[PROGRAM]", "getlist failed  ", err.Error())
// 		return
// 	}
// 	if res != nil {
// 		fmt.Println("getlist : ", res.(*mpubs.GetBattleSpaceListResp).Count)
// 	}
// }
