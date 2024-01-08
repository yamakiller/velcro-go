package apps

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/configs"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/rpc/client/clientpool"
	"github.com/yamakiller/velcro-go/vlog"
)

var index = int32(0)
func Test(cp *clientpool.ConnectPool) {
	for i:=0; i < envs.Instance().Get("configs").(*configs.Config).ClientNumber; i++ {
		go clientRun(cp,atomic.AddInt32(&index,1))
	}
}

func clientRun(cp *clientpool.ConnectPool,i int32) {
	
	uid:= singin(cp ,fmt.Sprintf("test_00%d&123456",i))
	if uid != ""{
		spaceid := createbattlespace(cp)
		if spaceid != ""{
			time.Sleep(5*time.Second)
			exitbattlespace(cp,spaceid,uid)
		}
	}
	// t1 := time.NewTicker(time.Millisecond*500)
	// for {
	// 	select{
	// 	case <-t1.C:
	// 		getlist(cp)
	// 	}
	// }
}

func singin(cp *clientpool.ConnectPool,token string)string{
	req := &mpubs.SignIn{
		Token: token,
	}
	res,err:= cp.RequestMessage(req,3000)
	if err != nil {
		vlog.Info("[PROGRAM]", "singin failed  ",err.Error())
		return ""
	}
	if res.Result() != nil{
		vlog.Info("[PROGRAM]", token,"  singin : ",res.Result().(*mpubs.SignInResp).Uid)
		return res.Result().(*mpubs.SignInResp).Uid
	}else if res.Error() != nil{
		vlog.Info("[PROGRAM]", token,"  singin : ", res.Error().Error())
	}
	return ""
}

func createbattlespace(cp *clientpool.ConnectPool)string{
	req := &mpubs.CreateBattleSpace{
		MapURI: "123456",
		MaxCount: 6,
	}
	res,err:= cp.RequestMessage(req,2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "createbattlespace failed  ",err.Error())
		return ""
	}
	if res.Result() !=nil{
		fmt.Println("createbattlespace : ",res.Result().(*mpubs.CreateBattleSpaceResp))	
		return res.Result().(*mpubs.CreateBattleSpaceResp).SpaceId
	}else if res.Error() != nil{
		vlog.Info("[PROGRAM]"," createbattlespace error: ", res.Error().Error())
	}
	return ""
}

func exitbattlespace(cp *clientpool.ConnectPool,spaceid string,uid string){
	
	req := &mprvs.RequestExitBattleSpace{
		BattleSpaceID: spaceid,
		UID: uid,
	}
	res,err:= cp.RequestMessage(req,2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "exitbattlespace failed  ",err.Error())
		return
	}
	if res.Result() !=nil{
		fmt.Println("exitbattlespace : ",res.Result().(*mprvs.RequestExitBattleSpace))	
	}else if res.Error() != nil{
		vlog.Info("[PROGRAM]"," exitbattlespace error: ", res.Error().Error())
	}
}

func getlist(cp *clientpool.ConnectPool){
	req := &mpubs.GetBattleSpaceList{
		Start: 0,
		Size: 10,
	}
	res,err:= cp.RequestMessage(req,2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "getlist failed  ",err.Error())
		return
	}
	if res.Result() !=nil{
		fmt.Println("getlist : ",res.Result().(*mpubs.GetBattleSpaceListResp).Count)	
	}else if res.Error() != nil{
		vlog.Info("[PROGRAM]"," getlist error: ", res.Error().Error())
	}
}