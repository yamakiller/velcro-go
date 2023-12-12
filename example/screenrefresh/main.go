package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

func main() {
	t := 0

	var (
		connSuccess    int32 = 1000
		connFailed     int32 = 50
		sendSuccess    int32 = 5
		sendFailed     int32 = 100023
		recviceSuccess int32 = 99999
		recviceFailed  int32 = 393
		sendBytes      int64 = 3434343434342323
		recvBytes      int64 = 99848334112
		success        int64 = 989899
		failed         int64 = 9999
	)
	for {
		var print string = ""
		print += fmt.Sprintf("\r 连接成功次数:%d\n", atomic.LoadInt32(&connSuccess))
		print += fmt.Sprintf("\r 连接失败次数:%d\n", atomic.LoadInt32(&connFailed))
		print += fmt.Sprintf("\r 发送数据成功次数:%d\n", atomic.LoadInt32(&sendSuccess))
		print += fmt.Sprintf("\r 发送数据失败次数:%d\n", atomic.LoadInt32(&sendFailed))
		print += fmt.Sprintf("\r 接收数据成功次数:%d\n", atomic.LoadInt32(&recviceSuccess))
		print += fmt.Sprintf("\r 接收数据失败次数:%d\n", atomic.LoadInt32(&recviceFailed))
		print += fmt.Sprintf("\r 发送字节数:%d 字节\n", atomic.LoadInt64(&sendBytes))
		print += fmt.Sprintf("\r 接收字节数:%d 字节\n", atomic.LoadInt64(&recvBytes))
		print += fmt.Sprintf("\r 总成功数:%d\n", atomic.LoadInt64(&success))
		print += fmt.Sprintf("\r 总失败数:%d\n", atomic.LoadInt64(&failed))

		fmt.Printf("%s", print)

		time.Sleep(time.Duration(2000) * time.Millisecond)

		/*for ic := 0; ic < len(print); ic++ {
			fmt.Printf("\b")
		}*/

		t++
		if t >= 10 {
			break
		}
	}
}
