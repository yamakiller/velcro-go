package rdsstruct

import (
	"github.com/yamakiller/velcro-go/example/monopoly/protocols/rdsstruct"
	"google.golang.org/protobuf/proto"
)

type RdsPlayerData struct {
	rdsstruct.RdsPlayerData
}

func (m *RdsPlayerData) MarshalBinary() (data []byte, err error) {
	return proto.Marshal(m)
}

func (m *RdsPlayerData) UnmarshalBinary(data []byte) error {
	return proto.Unmarshal(data, m)
}
