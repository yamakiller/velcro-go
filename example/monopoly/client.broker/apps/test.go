package apps

import (
	"fmt"
	"sync/atomic"
	"time"

	// "github.com/yamakiller/velcro-go/envs"
	// "github.com/yamakiller/velcro-go/example/monopoly/client.broker/configs"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/rpc/client/clientpool"
	"github.com/yamakiller/velcro-go/vlog"
)

var index = int32(1)

func Test(cp *clientpool.ConnectPool) {
	// for i:=0; i < envs.Instance().Get("configs").(*configs.Config).ClientNumber; i++ {
	go clientRun(cp, atomic.AddInt32(&index, 1))
	// }
}

func clientRun(cp *clientpool.ConnectPool, i int32) {

	uid := singin(cp, fmt.Sprintf("test_00%d&123456", i))
	if uid != "" {
		spaceid := createbattlespace(cp)
		// spaceid := "ARq985AvzGojkcVfcAGwiN"
		if spaceid != "" {
			// enterbattlespace(cp,spaceid)
			// time.Sleep(2 * time.Second)
			getlist(cp)
			time.Sleep(3 * time.Second)
			readybattlespace(cp, uid, spaceid, true)
			time.Sleep(6 * time.Second)
			readybattlespace(cp, uid, spaceid, false)
			// time.Sleep(3*time.Second)
			// exitbattlespace(cp,spaceid,uid)
		}
	}
	t1 := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-t1.C:
			getlist(cp)
		}
	}
}

func singin(cp *clientpool.ConnectPool, token string) string {
	req := &mpubs.SignIn{
		Token: token,
	}
	res, err := cp.RequestMessage(req, 3000)
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

func createbattlespace(cp *clientpool.ConnectPool) string {
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

func enterbattlespace(cp *clientpool.ConnectPool, spaceid string) string {
	req := &mpubs.EnterBattleSpace{
		SpaceId: spaceid,
	}
	res, err := cp.RequestMessage(req, 2000)
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
func readybattlespace(cp *clientpool.ConnectPool, uid, spaceid string, ready bool) bool {
	req := &mpubs.ReadyBattleSpace{
		SpaceId: spaceid,
		Uid:     uid,
		Ready:   ready,
	}
	res, err := cp.RequestMessage(req, 2000)
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

func exitbattlespace(cp *clientpool.ConnectPool, spaceid string, uid string) {

	req := &mprvs.RequestExitBattleSpace{
		BattleSpaceID: spaceid,
		UID:           uid,
	}
	res, err := cp.RequestMessage(req, 2000)
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

func getlist(cp *clientpool.ConnectPool) {
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
