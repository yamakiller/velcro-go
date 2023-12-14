package test2

import (
	"fmt"
	"os"

	"github.com/yamakiller/velcro-go/extensions"
)

type Test2 struct {
}

func (t *Test2) Test2() {
	fmt.Fprintf(os.Stderr,"Test2 extensionId %v\n",extensions.NextExtensionID())
}