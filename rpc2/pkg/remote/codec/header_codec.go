package codec

import (
	"context"
	"fmt"
	"math"

	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote/codec/perrors"
	"github.com/yamakiller/velcro-go/vlog"
)

/**
 *  VelcroHeader Protocol
 *     0 1 2 3 4 5 6 8 9 a b c d e f   0 1 2 3 4 5 6 7 8 9 a b c d e f
 *  +-------------2Byte--------------|-------------2Byte-------------+
 *  +----------------------------------------------------------------+
 *  | 0|				        TOTAL LENGTH 						 |
 *  +----------------------------------------------------------------+
 *  | 0|				       		SN								 |
 *  +-----------------------------------------------------------------
 *  | 0|       LABEL                |            FLAGS				 |
 *  +-----------------------------------------------------------------
 *  | 0|    Meta Data Size (/32)    | 			 V
 *  +--------------------------------
 *
 *
 *
 *	+----------------------------------------------------------------+
 *	| PROTOCOL ID  |NUM METAS . | META 0 ID (uint8)|
 *	+----------------------------------------------------------------+
 *	|  META 0 DATA ...
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

/**
 * LABEL
 * ProtobufV1Label、ThriftV1Label、ThriftFramedV1Label
 * Velcro Tag 118 120
 *+----------------------------------------------------------------+
 *
 */
const (
	VelcroHeaderSize            = 14
	VelcroHeaderFlagsStart      = 6
	VelcroHeaderSnStart         = 8
	VelcroHeaderSnBits          = Size32
	VelcroHeaderMetaSizeStart   = 12
	VelcroHeaderMetaSizeBits    = 2
	VelcroHeaderMetaSizeLimit   = math.MaxUint16
	VelcroHeaderTotalLengthBits = Size32

	VelcroHeaderMetaProtocolIDIndex = 0
	VelcroHeaderMetaIDOfNumberIndex = 1
	VelcroHeaderMetaIDStart         = 2
)

type VelcroHeaderFlags uint16

const (
	VelcroHeaderFlagsKey              string            = "V-H-F"
	VelcroHeaderFlagSupportOutOfOrder VelcroHeaderFlags = 0x01
	VelcroHeaderFlagDuplexReverse     VelcroHeaderFlags = 0x08
	VelcroHeaderFlagSASL              VelcroHeaderFlags = 0x10
)

// ProtocolID is the wrapped protocol id used in THeader.
type ProtocolID uint8

// Supported ProtocolID values.
const (
	ProtocolIDThriftBinary   ProtocolID = 0x00
	ProtocolIDVelcroProtobuf ProtocolID = 0x02
	ProtocolIDDefault                   = ProtocolIDThriftBinary
)

type velcroHeader struct{}

func (v *velcroHeader) decode(ctx context.Context, message remote.Message, in remote.ByteBuffer) error {
	header, err := in.Next(VelcroHeaderSize)
	if err != nil {
		return perrors.NewProtocolError(err)
	}

	// TODO: 检测是否属于Velcro消息头

	// 获取总包长度 + 获取Flag位置
	totalLength := Bytes2Uint32NoCheck(header[:VelcroHeaderTotalLengthBits])
	flags := Bytes2Uint16NoCheck(header[VelcroHeaderFlagsStart:])
	// 将flags信息解码到message的Tags中.
	setFlags(flags, message)

	seqID := Bytes2Uint32NoCheck(header[VelcroHeaderSnStart : VelcroHeaderSnStart+VelcroHeaderSnBits])
	//检测序列号合法性
	if err = SetOrCheckSeqID(int32(seqID), message); err != nil {
		vlog.Warnf("the seqID in VelcroHeader check failed, error=%s", err.Error())
		// some framework doesn't write correct seqID in VelcroHeader, to ignore err only check it in payload
		// print log to push the downstream framework to refine it.
	}

	headerMetaSize := Bytes2Uint16NoCheck(header[VelcroHeaderMetaSizeStart : VelcroHeaderMetaSizeStart+VelcroHeaderMetaSizeBits])
	if uint32(headerMetaSize) > VelcroHeaderMetaSizeLimit || headerMetaSize < 2 {
		return perrors.NewProtocolErrorWithMsg(fmt.Sprintf("invalid header length[%d]", headerMetaSize))
	}

	var headerMeta []byte
	if headerMeta, err = in.Next(int(headerMetaSize)); err != nil {
		return perrors.NewProtocolError(err)
	}
	// 校验是否支持此协议
	if err = verifyProtocolID(headerMeta[VelcroHeaderMetaProtocolIDIndex], message); err != nil {
		return err
	}
	metaIDNum := int(headerMeta[VelcroHeaderMetaIDOfNumberIndex])
	if int(headerMetaSize)-VelcroHeaderMetaIDStart < metaIDNum {
		// 数据不足
		return perrors.NewProtocolErrorWithType(perrors.InvalidData, fmt.Sprintf("need read %d metaIDs, but not enough", metaIDNum))
	}
	metaIDs := make([]uint8, metaIDNum)
	for i := 0; i < metaIDNum; i++ {
		metaIDs[i] = headerMeta[VelcroHeaderMetaIDStart+i]
	}
	// 1.开始读取meta => key,value信息并存入message中
	// 2.填充头信息

	// Pay framed 字段 Size32 ?
	message.SetPayloadLen(int(totalLength - VelcroHeaderSize - uint32(headerMetaSize) + Size32))

	return err
}

func verifyProtocolID(pid uint8, message remote.Message) error {
	switch pid {
	case uint8(ProtocolIDThriftBinary):
	case uint8(ProtocolIDVelcroProtobuf):
	default:
		return fmt.Errorf("unsupported ProtocolID[%d]", pid)
	}
	return nil
}

func setFlags(flags uint16, message remote.Message) {
	if message.MessageType() == remote.Call {
		message.Tags()[VelcroHeaderFlagsKey] = VelcroHeaderFlags(flags)
	}
}

/*
// Header 关键字
const (
	// Header Magic 部分
	// 第0位和第16位必须为0以区分成帧和非成帧
	VelcroHeaderMagic uint32 = 0x10000000 // 0001 0000 0000 0000 ...0000[4字节]
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

	message.Tags()["HeaderFlags"]
	// 数据包总长度字段
	totalLengthField = headerMeta[0:4]
	headerInfoSizeField := headerMeta[12:14]

	var transformIDs []uint8 // 不支持压缩
}
*/
