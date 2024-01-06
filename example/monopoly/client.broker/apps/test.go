package apps

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/configs"
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
	
	t1 := time.NewTicker(time.Millisecond*300)
	for {
		select{
		case <-t1.C:
			singin(cp ,fmt.Sprintf("test_00%d&123456",i))
		}
	}
}

func singin(cp *clientpool.ConnectPool,token string){
	req := &mpubs.SignIn{
		Token: token,
	}
	res,err:= cp.RequestMessage(req,2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "singin failed  ",err.Error())
		return
	}
	fmt.Println("singin : ",res)
	getlist(cp)
}

func getlist(cp *clientpool.ConnectPool){
	req := &mpubs.GetBattleSpaceList{
		Start: 0,
		Size: 10,
	}
	res,err:= cp.RequestMessage(req,2000)
	if err != nil {
		vlog.Info("[PROGRAM]", "getlist failed  ",err.Error())
	}
	fmt.Println("getlist : ",res)	
}