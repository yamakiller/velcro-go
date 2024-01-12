package apps

import (
	"fmt"
	"sync/atomic"

	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/tcpclient"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/rpc/client"
	"github.com/yamakiller/velcro-go/vlog"
)

var index = int32(1)

func Test() {
	for i := int32(0); i < envs.Instance().Get("configs").(*configs.Config).ClientNumber; i++ {
		go clientRun(atomic.AddInt32(&index, 1))
	}
}


func owner(cli client.IConnect,i int32) string{
	singin(cli,fmt.Sprintf("test_00%d&123456", i))
	// return createbattlespace(cli)
	return "1"
}
func user(cli client.IConnect,i int32,spaceid string){
	singin(cli,fmt.Sprintf("test_00%d&123456", i))
	// enterbattlespace(cli,spaceid)
	// readybattlespace(cli,uid,spaceid,true)
}


func NewConn() client.IConnect{
	cli := tcpclient.NewConn(client.WithClosed(tcpclient.Closed))
	cli.Dial( envs.Instance().Get("configs").(*configs.Config).TargetAddr,0)
	return cli
}

func clientRun(i int32) {
	cli1 := NewConn()
	owner(cli1,i)
	// cli2 := NewConn()
	// t1 := time.NewTicker(time.Millisecond * 500)
	// for {
	// 	select {
	// 	case <-t1.C:
	// 		if spaceid != ""{
	// 			// user(cli2,i + envs.Instance().Get("configs").(*configs.Config).ClientNumber,spaceid )
	// 			spaceid = ""
	// 		}
	// 	}
	// }
}

func singin(cp client.IConnect, token string) string {
	req := &mpubs.SignIn{
		Token: token,
	}
	res, err := cp.RequestMessage(req, 2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "singin failed  ", err.Error())
		return ""
	}
	if res.Result() != nil {
		vlog.Info("[PROGRAM]", token, "  singin : ", res.Result().(*mpubs.SignInResp).Uid)
		return res.Result().(*mpubs.SignInResp).Uid
	} else if res.Error() != nil {
		vlog.Info("[PROGRAM]", token, "  singin : ", res.Error().Error())
	}
	return ""
}

func createbattlespace(cp client.IConnect) string {
	req := &mpubs.CreateBattleSpace{
		MapURI:   "123456",
		MaxCount: 6,
	}
	res, err := cp.RequestMessage(req, 2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "createbattlespace failed  ", err.Error())
		return ""
	}
	if res.Result() != nil {
		fmt.Println("createbattlespace : ", res.Result().(*mpubs.CreateBattleSpaceResp))
		return res.Result().(*mpubs.CreateBattleSpaceResp).SpaceId
	} else if res.Error() != nil {
		vlog.Info("[PROGRAM]", " readybattlespace error: ", res.Error().Error())
	}
	return ""
}

func enterbattlespace(cp client.IConnect, spaceid string) string {
	req := &mpubs.EnterBattleSpace{
		SpaceId: spaceid,
	}
	res, err := cp.RequestMessage(req, 8000)
	if err != nil {
		vlog.Info("[PROGRAM]", "enterbattlespace failed  ", err.Error())
		return ""
	}
	if res.Result() != nil {
		fmt.Println("enterbattlespace : ", res.Result().(*mpubs.EnterBattleSpaceResp))
		return res.Result().(*mpubs.EnterBattleSpaceResp).Space.SpaceId
	} else if res.Error() != nil {
		vlog.Info("[PROGRAM]", " enterbattlespace error: ", res.Error().Error())
	}
	return ""
}

func readybattlespace(cp client.IConnect, uid, spaceid string, ready bool) bool {
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
	if res.Result() != nil {
		fmt.Println("readybattlespace : ", res.Result().(*mpubs.ReadyBattleSpaceResp))
		return res.Result().(*mpubs.ReadyBattleSpaceResp).Ready
	} else if res.Error() != nil {
		vlog.Info("[PROGRAM]", " readybattlespace error: ", res.Error().Error())
	}
	return false
}

func exitbattlespace(cp client.IConnect, spaceid string, uid string) {

	req := &mprvs.RequestExitBattleSpace{
		BattleSpaceID: spaceid,
		UID:           uid,
	}
	res, err := cp.RequestMessage(req, 8000)
	if err != nil {
		vlog.Info("[PROGRAM]", "exitbattlespace failed  ", err.Error())
		return
	}
	if res.Result() != nil {
		fmt.Println("exitbattlespace : ", res.Result().(*mprvs.RequestExitBattleSpace))
	} else if res.Error() != nil {
		vlog.Info("[PROGRAM]", " exitbattlespace error: ", res.Error().Error())
	}
}

func getlist(cp client.IConnect) {
	req := &mpubs.GetBattleSpaceList{
		Start: 0,
		Size:  10,
	}
	res, err := cp.RequestMessage(req, 2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "getlist failed  ", err.Error())
		return
	}
	if res.Result() != nil {
		fmt.Println("getlist : ", res.Result().(*mpubs.GetBattleSpaceListResp).Count)
	} else if res.Error() != nil {
		vlog.Info("[PROGRAM]", " getlist error: ", res.Error().Error())
	}
}
