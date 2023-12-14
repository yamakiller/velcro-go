package main

import (
	"github.com/yamakiller/velcro-go/example/metrics/test1"
	"github.com/yamakiller/velcro-go/example/metrics/test2"
)

func main() {
	t1 := test1.Test1{}
	

	t2 := test2.Test2{}
	
	t2.Test2()
	t1.Test1()
}