package test1

import (
	"fmt"
	"os"

	"github.com/yamakiller/velcro-go/extensions"
)

type Test1 struct {
}

func (t *Test1) Test1() {
	fmt.Fprintf(os.Stderr,"Test1 extensionId %v\n",extensions.NextExtensionID())
}