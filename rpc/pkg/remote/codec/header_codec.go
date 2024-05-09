package codec

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/yamakiller/velcro-go/rpc/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc/pkg/remote/codec/perrors"
	"github.com/yamakiller/velcro-go/rpc/pkg/remote/tansmeta"
	"github.com/yamakiller/velcro-go/rpc/pkg/serviceinfo"
	"github.com/yamakiller/velcro-go/vlog"
)

/**
 *  VelcroHeader Protocol
 *     0 1 2 3 4 5 6 8 9 a b c d e f   0 1 2 3 4 5 6 7 8 9 a b c d e f
 *  +-------------2Byte--------------|-------------2Byte-------------+
 *  +----------------------------------------------------------------+
 *  | 0|				        DATA LENGTH 						 |
 *  +----------------------------------------------------------------+
 *  | 0|				       	SERIAL NUMBER						 |
 *  +-----------------------------------------------------------------
 *  | 0|        V01                                  |  	MAGIC	 |
 *  +-----------------------------------------------------------------
 *  | 0|        FLAGS                |       META SIZE (/32)         |
 *  +-----------------------------------------------------------------
 *
 *  说明:
 *  	DATA LENGTH:   除开固定头大小的实际数据大小
 *  	SERIAL NUMBER: 数据包唯一序列号
 *  	V01:           为VELCRO协议版本号-V代表:Velcro,A-C代表版本类型,0-F为协议版本编号
 *  	MAGIC:		   标记ProtobufV1Magic、ThriftV1Magic、ThriftFramedV1Magic
 *  	FLAGS: 		   标签记录协议标记
 *  	META SIZE:     Meta数据块大小, 最大限制65535
 *
 *  META DATA Protocol
 *    0 1 2 3 4 5 6 8 9 a b c d e f    0 1 2 3 4 5 6 7 8 9 a b c d e f
 *  +-------------2Byte--------------|-------------2Byte-------------+
 *  +----------------------------------------------------------------+
 *  | PROTOCOL ID  |  META OF NUMBER | META ID (uint8) ... |
 *  +----------------------------------------------------------------+
 *  | META DATA ...
 *  +-----------------------------------------------------------------
 *  |            ...                              ...                |
 *  +-----------------------------------------------------------------
 *  |          INFO ID (uint8)      |      INFO DATA ...
 *  +----------------------------------------------------------------+
 *  |            ....                            ....                |
 *  +----------------------------------------------------------------+
 *	|                                                                |
 *	|                              PAYLOAD                           |
 *	|                                                                |
 *	+----------------------------------------------------------------+
 */

const (
	VelcroVersion = "VA1"
)

const (
	VelcroFixedHeaderSize           = 16
	VelcroHeaderDataSizeStart       = 0
	VelcroHeaderDataSizeBits        = Size32
	VelcroHeaderSequenceNumberStart = VelcroHeaderDataSizeStart + VelcroHeaderDataSizeBits
	VelcroHeaderSequenceNumberBits  = Size32
	VelcroHeaderVersionStart        = VelcroHeaderSequenceNumberStart + VelcroHeaderSequenceNumberBits
	VelcroHeaderVersionBits         = 3

	VelcroHeaderMagicStart = VelcroHeaderVersionStart + VelcroHeaderVersionBits
	VelcroHeaderMagicBits  = 1

	VelcroHeaderFlagStart = VelcroHeaderMagicStart + VelcroHeaderMagicBits
	VelcroHeaderFlagBits  = Size16

	VelcroHeaderMetaSizeStart = VelcroHeaderFlagStart + VelcroHeaderFlagBits
	VelcroHeaderMetaSizeBits  = Size16
	VelcroHeaderMetaSizeLimit = math.MaxUint16

	VelcroHeaderMetaProtocolIDOffset = 0
	VelcroHeaderMetaIDNumberOffset   = 1
	VelcroHeaderMetaIDOffset         = 2
)

/**
 * LABEL
 * ProtobufV1Label、ThriftV1Label、ThriftFramedV1Label
 * Velcro Tag 118 120
 *+----------------------------------------------------------------+
 *
 */

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

type InfoIDType uint8 // uint8

const (
	InfoIDPadding     InfoIDType = 0x00
	InfoIDKeyValue    InfoIDType = 0x01
	InfoIDIntKeyValue InfoIDType = 0x02
	InfoIDACLToken    InfoIDType = 0x03
)

type velcroHeader struct{}

func (v *velcroHeader) decode(ctx context.Context, message remote.Message, in remote.ByteBuffer) error {
	header, err := in.Next(VelcroFixedHeaderSize)
	if err != nil {
		return perrors.NewProtocolError(err)
	}

	version := header[VelcroHeaderVersionStart : VelcroHeaderVersionStart+VelcroHeaderVersionBits]

	if string(version) != VelcroVersion {
		vlog.Warnf("velcro protocol version error, version:%s", string(version))
		return perrors.NewProtocolError(fmt.Errorf("velcro protocol version error"))
	}

	// 获取数据包除固定头以外的大小
	dataTotoalLength := Bytes2Uint32NoCheck(header[VelcroHeaderDataSizeStart : VelcroHeaderDataSizeStart+VelcroHeaderDataSizeBits])
	flags := Bytes2Uint16NoCheck(header[VelcroHeaderFlagStart : VelcroHeaderFlagStart+VelcroHeaderFlagBits])

	// 将flags信息解码到message的Tags中.
	setFlags(flags, message)

	seqID := Bytes2Uint32NoCheck(header[VelcroHeaderSequenceNumberStart : VelcroHeaderSequenceNumberStart+VelcroHeaderSequenceNumberBits])
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
	if err = verifyProtocolID(headerMeta[VelcroHeaderMetaProtocolIDOffset], message); err != nil {
		return err
	}
	metaIDNum := int(headerMeta[VelcroHeaderMetaIDNumberOffset])
	if int(headerMetaSize)-VelcroHeaderMetaIDOffset < metaIDNum {
		// 数据不足
		return perrors.NewProtocolErrorWithType(perrors.InvalidData, fmt.Sprintf("need read %d metaIDs, but not enough", metaIDNum))
	}
	metaIDs := make([]uint8, metaIDNum)
	for i := 0; i < metaIDNum; i++ {
		metaIDs[i] = headerMeta[VelcroHeaderMetaIDOffset+i]
	}
	// 1.开始读取meta => key,value信息并存入message中
	if err := readInfo(VelcroHeaderMetaIDOffset+metaIDNum, headerMeta, message); err != nil {
		return perrors.NewProtocolError(err)
	}
	// 2.填充头信息

	// + Size32 Pay framed 字段 大小
	message.SetPayloadLen(int(dataTotoalLength - uint32(headerMetaSize) + Size32))

	return err
}

func (v *velcroHeader) encode(ctx context.Context, message remote.Message, out remote.ByteBuffer) (dataTotalLengthField []byte, err error) {
	// 1. header
	var header []byte
	header, err = out.Malloc(VelcroFixedHeaderSize)
	if err != nil {
		return nil, perrors.NewProtocolErrorWithMsg(fmt.Sprintf("VelcroHader malloc header failed, %s", err.Error()))
	}

	dataTotalLengthField = header[VelcroHeaderDataSizeStart:VelcroHeaderDataSizeBits]
	headerMetaSizeField := header[VelcroHeaderMetaSizeStart : VelcroHeaderMetaSizeStart+VelcroHeaderMetaSizeBits]
	binary.BigEndian.PutUint32(header[VelcroHeaderSequenceNumberStart:VelcroHeaderSequenceNumberStart+VelcroHeaderSequenceNumberBits], uint32(message.RPCInfo().Invocation().SeqID()))
	copy(header[VelcroHeaderVersionStart:VelcroHeaderVersionStart+VelcroHeaderVersionBits], []byte(VelcroVersion))

	binary.BigEndian.PutUint16(header[VelcroHeaderFlagStart:VelcroHeaderFlagStart+VelcroHeaderFlagBits], uint16(getFlags(message)))

	// 2. header meta, malloc and write
	if err = WriteByte(byte(getProtocolID(message.ProtocolInfo())), out); err != nil {
		return nil, perrors.NewProtocolErrorWithMsg(fmt.Sprintf("VelcroHeader write protocol id failed, %s", err.Error()))
	}

	var metaIDs []uint8
	if err = WriteByte(byte(len(metaIDs)), out); err != nil {
		return nil, perrors.NewProtocolErrorWithMsg(fmt.Sprintf("VelcroHeader write metaIDs length failed, %s", err.Error()))
	}

	for mid := range metaIDs {
		if err = WriteByte(byte(mid), out); err != nil {
			return nil, perrors.NewProtocolErrorWithMsg(fmt.Sprintf("VelcroHeader write transformIDs failed, %s", err.Error()))
		}
	}

	headerMetaSize := 1 + 1 + len(metaIDs)
	headerMetaSize, err = writeInfo(headerMetaSize, message, out)
	if err != nil {
		return nil, perrors.NewProtocolErrorWithMsg((fmt.Sprintf("VelcroHeader write key/value info failed, %s", err.Error())))
	}

	if uint32(headerMetaSize) > VelcroHeaderMetaSizeLimit {
		return nil, perrors.NewProtocolErrorWithMsg(fmt.Sprintf("invalid header length[%d]", headerMetaSize))
	}

	binary.BigEndian.PutUint16(headerMetaSizeField, uint16(headerMetaSize/4))

	return dataTotalLengthField, err
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

func getFlags(message remote.Message) VelcroHeaderFlags {
	var headerFlags VelcroHeaderFlags
	if message.Tags() != nil && message.Tags()[VelcroHeaderFlagsKey] != nil {
		if hfs, ok := message.Tags()[VelcroHeaderFlagsKey].(VelcroHeaderFlags); ok {
			headerFlags = hfs
		} else {
			vlog.Warnf("VELCRO: the type of headerFlags is invalid, %T", message.Tags()[VelcroHeaderFlagsKey])
		}
	}

	return headerFlags
}

func readInfo(offset int, buf []byte, message remote.Message) error {
	intInfo := message.TransInfo().TransIntInfo()
	strInfo := message.TransInfo().TransStrInfo()
	for {
		infoID, err := Bytes2Uint8(buf, offset)
		offset++
		if err != nil {
			// this is the last field, read until there is no more padding
			if err == io.EOF {
				break
			}
			return err
		}

		switch InfoIDType(infoID) {
		case InfoIDPadding:
			continue
		case InfoIDIntKeyValue:
			err := readIntKeyValueInfo(&offset, buf, intInfo)
			if err != nil {
				return err
			}
			break
		case InfoIDKeyValue:
			err := readStrKeyValueInfo(&offset, buf, strInfo)
			if err != nil {
				return err
			}
			break
		case InfoIDACLToken:
			err := readACLToken(&offset, buf, strInfo)
			if err != nil {
				return err
			}
			break
		default:
			return fmt.Errorf("invalid infoIDType[%#x]", infoID)
		}
	}
	return nil
}

func readIntKeyValueInfo(offset *int, buf []byte, info map[uint16]string) error {
	sz, err := Bytes2Uint16(buf, *offset)
	*offset += 2
	if err != nil {
		return fmt.Errorf("error reading int key/value info size: %s", err.Error())
	}
	if sz <= 0 {
		return nil
	}
	for i := uint16(0); i < sz; i++ {
		key, err := Bytes2Uint16(buf, *offset)
		*offset += 2
		if err != nil {
			return fmt.Errorf("error reading int key/value info: %s", err.Error())
		}
		val, n, err := ReadString2BLen(buf, *offset)
		*offset += n
		if err != nil {
			return fmt.Errorf("error reading int key/value info: %s", err.Error())
		}
		info[key] = val
	}
	return nil
}

func readStrKeyValueInfo(offset *int, buf []byte, info map[string]string) error {
	sz, err := Bytes2Uint16(buf, *offset)
	*offset += 2
	if err != nil {
		return fmt.Errorf("error reading str key/value info size: %s", err.Error())
	}
	if sz <= 0 {
		return nil
	}
	for i := uint16(0); i < sz; i++ {
		key, n, err := ReadString2BLen(buf, *offset)
		*offset += n
		if err != nil {
			return fmt.Errorf("error reading str key/value info: %s", err.Error())
		}
		val, n, err := ReadString2BLen(buf, *offset)
		*offset += n
		if err != nil {
			return fmt.Errorf("error reading str key/value info: %s", err.Error())
		}
		info[key] = val
	}
	return nil
}

func readACLToken(offset *int, buf []byte, info map[string]string) error {
	val, n, err := ReadString2BLen(buf, *offset)
	*offset += n
	if err != nil {
		return fmt.Errorf("error reading acl token: %s", err.Error())
	}

	info[tansmeta.GDPRToken] = val
	return nil
}

// protoID
func getProtocolID(pi remote.ProtocolInfo) ProtocolID {
	switch pi.CodecType {
	case serviceinfo.Protobuf:
		return ProtocolIDDefault
	}
	return ProtocolIDDefault
}

func writeInfo(writtenSize int, message remote.Message, out remote.ByteBuffer) (int, error) {
	writeSize := writtenSize
	tm := message.TransInfo()
	// string key/value info
	strKVMap := tm.TransStrInfo()
	strKVSize := len(strKVMap)
	// write gdpr token into InfoIDACLToken
	if gdprToken, ok := strKVMap[tansmeta.GDPRToken]; ok {
		strKVSize--
		// INFO ID TYPE(u8)
		if err := WriteByte(byte(InfoIDACLToken), out); err != nil {
			return writeSize, err
		}
		writeSize += 1

		wLen, err := WriteString2BLen(gdprToken, out)
		if err != nil {
			return writeSize, err
		}
		writeSize += wLen
	}

	if strKVSize > 0 {
		// INFO ID TYPE(u8) + NUM HEADERS(u16)
		if err := WriteByte(byte(InfoIDKeyValue), out); err != nil {
			return writeSize, err
		}
		if err := WriteUint16(uint16(strKVSize), out); err != nil {
			return writeSize, err
		}
		writeSize += 3
		for key, val := range strKVMap {
			if key == tansmeta.GDPRToken {
				continue
			}
			keyWLen, err := WriteString2BLen(key, out)
			if err != nil {
				return writeSize, err
			}
			valWLen, err := WriteString2BLen(val, out)
			if err != nil {
				return writeSize, err
			}
			writeSize = writeSize + keyWLen + valWLen
		}
	}

	// padding = (4 - headerInfoSize%4) % 4
	padding := (4 - writeSize%4) % 4
	paddingBuf, err := out.Malloc(padding)
	if err != nil {
		return writeSize, err
	}
	for i := 0; i < len(paddingBuf); i++ {
		paddingBuf[i] = byte(0)
	}
	writeSize += padding
	return writeSize, nil
}
