package rdsstruct

import (
	"github.com/yamakiller/velcro-go/example/monopoly/protocols/rdsstruct"
	"google.golang.org/protobuf/proto"
)

type RdsBattleSpaceData struct {
	rdsstruct.RdsBattleSpaceData
}

func (m *RdsBattleSpaceData) MarshalBinary() (data []byte, err error) {
	return proto.Marshal(m)
}

func (m *RdsBattleSpaceData) UnmarshalBinary(data []byte) error {
	return proto.Unmarshal(data, m)
}
