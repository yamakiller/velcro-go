package codec

import (
	"context"
	"fmt"

	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote/codec/perrors"
)

/**
 *  VelcroHeader Protocol
 *  +-------------2Byte--------------|-------------2Byte-------------+
 *	+----------------------------------------------------------------+
 *	| 0|                          LENGTH                             |
 *	+----------------------------------------------------------------+
 *	| 0|       HEADER MAGIC          |            FLAGS              |
 *	+----------------------------------------------------------------+
 *	|                         SEQUENCE NUMBER                        |
 *	+----------------------------------------------------------------+
 *	| 0|     Header Size(/32)        | ...
 *	+---------------------------------
 *
 *	头大小变量存放于第12位2个字节,并从14位开始:
 *
 *	+----------------------------------------------------------------+
 *	| PROTOCOL ID  |NUM TRANSFORMS . | TRANSFORM 0 ID (uint8)|
 *	+----------------------------------------------------------------+
 *	|  TRANSFORM 0 DATA ...
 *	+----------------------------------------------------------------+
 *	|         ...                              ...                   |
 *	+----------------------------------------------------------------+
 *	|        INFO 0 ID (uint8)      |       INFO 0  DATA ...
 *	+----------------------------------------------------------------+
 *	|         ...                              ...                   |
 *	+----------------------------------------------------------------+
 *	|                                                                |
 *	|                              PAYLOAD                           |
 *	|                                                                |
 *	+----------------------------------------------------------------+
 */

// Header 关键字
const (
	// Header Magic 部分
	// 第0位和第16位必须为0以区分成帧和非成帧
	VelcroHeaderMagic uint32 = 0x10000000 // 1000 0000 0000 0000 ...0000[4字节]
	MeshHeaderMagic   uint32 = 0xFFAF0000 // 1111 1111 1010 1111 ...0000[4字节]
	MeshHeaderLenMask uint32 = 0x0000FFFF // 低4位掩码

	// HeaderMask uint32 = 0xFFFF0000
	FlagsMask     uint32 = 0x0000FFFF
	MethodMask    uint32 = 0x41000000 // 0100 0001 0000 0000 ...0000[4字节] method first byte [A-Za-z_]
	MaxFrameSize  uint32 = 0x3FFFFFFF // 0011 1111 1111 1111 ...1111[4字节]
	MaxHeaderSize uint32 = 65536
)

type HeaderFlags uint16

const (
	HeaderFlagsKey              string      = "HeaderFlags"
	HeaderFlagSupportOutOfOrder HeaderFlags = 0b00001
	HeaderFlagDuplexReverse     HeaderFlags = 0b10000
	HeaderFlagSASL              HeaderFlags = 0b10000
)

const (
	VelcroHeaderMetaSize = 14
)

// ProtocolID is the wrapped protocol id used in THeader.
type ProtocolID uint8

// Supported ProtocolID values.
const (
	ProtocolIDThriftBinary    ProtocolID = 0b000
	ProtocolIDThriftCompact   ProtocolID = 0b010 // Velcro not support
	ProtocolIDThriftCompactV2 ProtocolID = 0b011 // Velcro not support
	ProtocolIDVelcroProtobuf  ProtocolID = 0b100
	ProtocolIDDefault                    = ProtocolIDThriftBinary
)

type InfoIDType uint8 // uint8

const (
	InfoIDPadding     InfoIDType = 0b00000
	InfoIDKeyValue    InfoIDType = 0b00001
	InfoIDIntKeyValue InfoIDType = 0b10000
	InfoIDACLToken    InfoIDType = 0b10001
)

type VelcroHeader struct{}

func (v VelcroHeader) encode(ctx context.Context, message remote.Message, out remote.ByteBuffer) (totalLengthField []byte, err error) {
	// 1.header meta
	var headerMeta []byte
	headerMeta, err = out.Malloc(VelcroHeaderMetaSize)
	if err != nil {
		return nil, perrors.NewProtocolErrorWithMsg(fmt.Sprintf("VelcroHeader malloc header meta failed, %s", err.Error()))
	}

	// 数据包总长度字段
	totalLengthField = headerMeta[0:4]
	headerInfoSizeField := headerMeta[12:14]

	var transformIDs []uint8 // 不支持压缩
}
