package gateway

import (
	"github.com/yamakiller/velcro-go/utils/encryption/ecdh"
)

type Encryption struct {
	Ecdh ecdh.ECDH
}
