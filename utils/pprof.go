package utils

import (
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func Pprof(addr string) {
	go func() {
		runtime.SetBlockProfileRate(1) // 开启对阻塞操作的跟踪，block

		runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪，mutex

		http.ListenAndServe(addr, nil)
		
	}()
}
