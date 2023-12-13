package main

import (
	"fmt"

	"github.com/yamakiller/velcro-go/utils/snowflakealien"
)

func main() {

	for i := 0; i < 100; i++ {
		fmt.Printf("%d\n", snowflakealien.Generate())

	}
}
