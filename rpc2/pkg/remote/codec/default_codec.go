package codec

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync/atomic"

	"github.com/yamakiller/velcro-go/rpc/transport"
	"github.com/yamakiller/velcro-go/rpc/utils/verrors"
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc2/pkg/remote/codec/perrors"
	"github.com/yamakiller/velcro-go/rpc2/pkg/retry"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/rpc2/pkg/serviceinfo"
)

// The byte count of 32 and 16 integer values.
const (
	Size32 = 4
	Size16 = 2
)

const (

	// ProtobufV1Label the label code for velcro protobuf
	ProtobufV1Label = 0x40000000
	// ThriftV1Label the label code for thrift.VERSION_1
	ThriftV1Label = 0x80000000
	// ThriftFramedV1Label the label code for thrift.VERSION_1 and framed
	ThriftFramedV1Label = 0xC0000000

	// LabelProtoMask is bit mask for checking version.
	LabelProtoMask = 0xC0000000
	// LabelMask is bit mask for checking version.
	//LabelMask = 0xFFFF0000
)

var (
	velcroHeaderCodec = velcroHeader{}

	_ remote.Codec = (*defaultCodec)(nil)
)

type defaultCodec struct {
	// maxSize  limits the max size of the payload
	maxSize int
}

// DecodeMeta 解码Header、Meta
func (c *defaultCodec) DecodeMeta(ctx context.Context, message remote.Message, in remote.ByteBuffer) (err error) {
	var flagBuf []byte
	// Header -> LENGTH + Label + Flags
	if flagBuf, err = in.Peek(2 * Size32); err != nil {
		return perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("default codec read failed: %s", err.Error()))
	}

	if err = verifyRPCState(ctx, message); err != nil {
		// 重试任务中有一个调用已完成，不需要对此调用进行解码
		return err
	}

	// 解码Header
	if err = velcroHeaderCodec.decode(ctx, message, in); err != nil {
		return err
	}

	return verifyPayload(flagBuf, message, in, true, c.maxSize)
}

func (c *defaultCodec) DecodePayload(ctx context.Context, message remote.Message, in remote.ByteBuffer) error {
	// 统计信息
	defer func() {
		if rio := message.RPCInfo(); rio != nil {
			if ms := rpcinfo.AsMutableRPCStats(rio.Stats()); ms != nil {
				ms.SetRecvSize(uint64(in.ReadLen()))
			}
		}
	}()

	nRead := in.ReadLen()
	// 获取解码器
	pCodec, err := remote.GetPayloadCodec(message)
	if err != nil {
		return err
	}
	// 解码
	if err = pCodec.Unmarshal(ctx, message, in); err != nil {
		return err
	}

	if message.PayloadLen() == 0 {
		// if protocol is PurePayload, should set payload length after decoded
		message.SetPayloadLen(in.ReadLen() - nRead)
	}
	return nil
}

// Decode 解码
func (c *defaultCodec) Decode(ctx context.Context, message remote.Message, in remote.ByteBuffer) error {
	// 1.解码元信息(meta)
	if err := c.DecodeMeta(ctx, message, in); err != nil {
		return err
	}

	//2.解码数据体
	return c.DecodePayload(ctx, message, in)
}

/**
 * Velcro protobuf 字段
 * +------------------------------------------------------------+
 * |                  4Byte                 |       2Byte       |
 * +------------------------------------------------------------+
 * |   			     Length			    	|   HEADER LABEL    |
 * +------------------------------------------------------------+
 */
func isProtobufVelcro(flagBuf []byte) bool {
	return (binary.BigEndian.Uint32(flagBuf[Size32:]) & LabelProtoMask) == ProtobufV1Label
}

/**
 * +------------------------------------------------------------+
 * |                  4Byte                 |       2Byte       |
 * +------------------------------------------------------------+
 * |   			     Length			    	|   HEADER LABEL    |
 * +------------------------------------------------------------+
 */
func isThriftBinary(flagBuf []byte) bool {
	return (binary.BigEndian.Uint32(flagBuf[Size32:]) & LabelProtoMask) == ThriftV1Label
}

/**
 * +------------------------------------------------------------+
 * |                  4Byte                 |       2Byte       |
 * +------------------------------------------------------------+
 * |   			     Length			    	|   HEADER LABEL    |
 * +------------------------------------------------------------+
 */

func isThriftFramedBinary(flagBuf []byte) bool {
	return (binary.BigEndian.Uint32(flagBuf[Size32:]) & LabelProtoMask) == ThriftFramedV1Label

}

func verifyRPCState(ctx context.Context, message remote.Message) error {
	if message.RPCRole() == remote.Server {
		return nil
	}

	if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
		return verrors.ErrRPCFinish
	}
	if respOp, ok := ctx.Value(retry.ContextRespOp).(*int32); ok {
		if !atomic.CompareAndSwapInt32(respOp, retry.OpNo, retry.OpDoing) {
			// 先前的调用正在处理或完成此标志用于检查重试(回退请求)场景中的请求状态.
			return verrors.ErrRPCFinish
		}
	}
	return nil
}

// verifyPayload 校验数据体
func verifyPayload(flagBuf []byte, message remote.Message, in remote.ByteBuffer, isHeader bool, maxPayloadSize int) error {
	var (
		transProto transport.Protocol
		codecType  serviceinfo.PayloadCodec
	)

	if isThriftBinary(flagBuf) {
		codecType = serviceinfo.Thrift
		if isHeader {
			transProto = transport.VHeader
		} else {
			transProto = transport.PurePayload
		}
	} else if isThriftFramedBinary(flagBuf) {
		codecType = serviceinfo.Thrift
		if isHeader {
			transProto = transport.VHeaderFramed
		} else {
			transProto = transport.Framed
		}
		payloadLen := binary.BigEndian.Uint32((flagBuf[:Size32]))
		message.SetPayloadLen(int(payloadLen))
		if err := in.Skip(Size32); err != nil {
			return err
		}
	} else if isProtobufVelcro(flagBuf) {
		codecType = serviceinfo.Protobuf
		if isHeader {
			transProto = transport.VHeaderFramed
		} else {
			transProto = transport.Framed
		}
		payloadLen := binary.BigEndian.Uint32(flagBuf[:Size32])
		message.SetPayloadLen(int(payloadLen))
		if err := in.Skip(Size32); err != nil {
			return err
		}
	} else {
		first4Bytes := binary.BigEndian.Uint32(flagBuf[:Size32])
		second4Bytes := binary.BigEndian.Uint32(flagBuf[Size32:])
		// 0xfff4fffd is the interrupt message of telnet
		err := perrors.NewProtocolErrorWithMsg(fmt.Sprintf("invalid payload (first4Bytes=%#x, second4Bytes=%#x)", first4Bytes, second4Bytes))
		return err
	}
	if err := verifyPayloadSize(message.PayloadLen(), maxPayloadSize); err != nil {
		return err
	}

	message.SetProtocolInfo(remote.NewProtocolInfo(transProto, codecType))
	cfg := rpcinfo.AsMutableRPCConfig(message.RPCInfo().Config())
	if cfg != nil {
		tp := message.ProtocolInfo().TransProto
		cfg.SetTransportProtocol(tp)
	}
	return nil
}

func verifyPayloadSize(payloadLen, maxSize int) error {
	if maxSize > 0 && payloadLen > 0 && payloadLen > maxSize {
		return perrors.NewProtocolErrorWithType(
			perrors.InvalidData,
			fmt.Sprintf("invalid data: payload size(%d) larger than the limit(%d)", payloadLen, maxSize),
		)
	}
	return nil
}

/*const (
	// ThriftV1Magic is the magic code for thrift.VERSION_1
	ThriftV1Magic = 0x80010000
	// ProtobufV1Magic is the magic code for kitex protobuf
	ProtobufV1Magic = 0x90010000

	// MagicMask is bit mask for checking version.
	MagicMask = 0xffff0000
)

var (
	ttHeaderCodec   = ttHeader{}
	meshHeaderCodec = meshHeader{}

	_ remote.Codec       = (*defaultCodec)(nil)
	_ remote.MetaDecoder = (*defaultCodec)(nil)
)

// NewDefaultCodec creates the default protocol sniffing codec supporting thrift and protobuf.
func NewDefaultCodec() remote.Codec {
	// No size limit by default
	return &defaultCodec{
		maxSize: 0,
	}
}

// NewDefaultCodecWithSizeLimit creates the default protocol sniffing codec supporting thrift and protobuf but with size limit.
// maxSize is in bytes
func NewDefaultCodecWithSizeLimit(maxSize int) remote.Codec {
	return &defaultCodec{
		maxSize: maxSize,
	}
}

type defaultCodec struct {
	// maxSize limits the max size of the payload
	maxSize int
}

// EncodePayload encode payload
func (c *defaultCodec) EncodePayload(ctx context.Context, message remote.Message, out remote.ByteBuffer) error {
	defer func() {
		// notice: mallocLen() must exec before flush, or it will be reset
		if ri := message.RPCInfo(); ri != nil {
			if ms := rpcinfo.AsMutableRPCStats(ri.Stats()); ms != nil {
				ms.SetSendSize(uint64(out.MallocLen()))
			}
		}
	}()
	var err error
	var framedLenField []byte
	headerLen := out.MallocLen()
	tp := message.ProtocolInfo().TransProto

	// 1. malloc framed field if needed
	if tp&transport.Framed == transport.Framed {
		if framedLenField, err = out.Malloc(Size32); err != nil {
			return err
		}
		headerLen += Size32
	}

	// 2. encode payload
	if err = c.encodePayload(ctx, message, out); err != nil {
		return err
	}

	// 3. fill framed field if needed
	var payloadLen int
	if tp&transport.Framed == transport.Framed {
		if framedLenField == nil {
			return perrors.NewProtocolErrorWithMsg("no buffer allocated for the framed length field")
		}
		payloadLen = out.MallocLen() - headerLen
		binary.BigEndian.PutUint32(framedLenField, uint32(payloadLen))
	} else if message.ProtocolInfo().CodecType == serviceinfo.Protobuf {
		return perrors.NewProtocolErrorWithMsg("protobuf just support 'framed' trans proto")
	}
	if tp&transport.TTHeader == transport.TTHeader {
		payloadLen = out.MallocLen() - Size32
	}
	err = checkPayloadSize(payloadLen, c.maxSize)
	return err
}

// EncodeMetaAndPayload encode meta and payload
func (c *defaultCodec) EncodeMetaAndPayload(ctx context.Context, message remote.Message, out remote.ByteBuffer, me remote.MetaEncoder) error {
	var err error
	var totalLenField []byte
	tp := message.ProtocolInfo().TransProto

	// 1. encode header and return totalLenField if needed
	// totalLenField will be filled after payload encoded
	if tp&transport.TTHeader == transport.TTHeader {
		if totalLenField, err = ttHeaderCodec.encode(ctx, message, out); err != nil {
			return err
		}
	}
	// 2. encode payload
	if err = me.EncodePayload(ctx, message, out); err != nil {
		return err
	}
	// 3. fill totalLen field for header if needed
	if tp&transport.TTHeader == transport.TTHeader {
		if totalLenField == nil {
			return perrors.NewProtocolErrorWithMsg("no buffer allocated for the header length field")
		}
		payloadLen := out.MallocLen() - Size32
		binary.BigEndian.PutUint32(totalLenField, uint32(payloadLen))
	}
	return nil
}

// Encode implements the remote.Codec interface, it does complete message encode include header and payload.
func (c *defaultCodec) Encode(ctx context.Context, message remote.Message, out remote.ByteBuffer) (err error) {
	return c.EncodeMetaAndPayload(ctx, message, out, c)
}

// DecodeMeta decode header
func (c *defaultCodec) DecodeMeta(ctx context.Context, message remote.Message, in remote.ByteBuffer) (err error) {
	var flagBuf []byte
	if flagBuf, err = in.Peek(2 * Size32); err != nil {
		return perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("default codec read failed: %s", err.Error()))
	}

	if err = checkRPCState(ctx, message); err != nil {
		// there is one call has finished in retry task, it doesn't need to do decode for this call
		return err
	}
	isTTHeader := IsTTHeader(flagBuf)
	// 1. decode header
	if isTTHeader {
		// TTHeader
		if err = ttHeaderCodec.decode(ctx, message, in); err != nil {
			return err
		}
		if flagBuf, err = in.Peek(2 * Size32); err != nil {
			return perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("ttheader read payload first 8 byte failed: %s", err.Error()))
		}
	} else if isMeshHeader(flagBuf) {
		message.Tags()[remote.MeshHeader] = true
		// MeshHeader
		if err = meshHeaderCodec.decode(ctx, message, in); err != nil {
			return err
		}
		if flagBuf, err = in.Peek(2 * Size32); err != nil {
			return perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("meshHeader read payload first 8 byte failed: %s", err.Error()))
		}
	}
	return checkPayload(flagBuf, message, in, isTTHeader, c.maxSize)
}

// DecodePayload decode payload
func (c *defaultCodec) DecodePayload(ctx context.Context, message remote.Message, in remote.ByteBuffer) error {
	defer func() {
		if ri := message.RPCInfo(); ri != nil {
			if ms := rpcinfo.AsMutableRPCStats(ri.Stats()); ms != nil {
				ms.SetRecvSize(uint64(in.ReadLen()))
			}
		}
	}()

	hasRead := in.ReadLen()
	pCodec, err := remote.GetPayloadCodec(message)
	if err != nil {
		return err
	}
	if err = pCodec.Unmarshal(ctx, message, in); err != nil {
		return err
	}
	if message.PayloadLen() == 0 {
		// if protocol is PurePayload, should set payload length after decoded
		message.SetPayloadLen(in.ReadLen() - hasRead)
	}
	return nil
}

// Decode implements the remote.Codec interface, it does complete message decode include header and payload.
func (c *defaultCodec) Decode(ctx context.Context, message remote.Message, in remote.ByteBuffer) (err error) {
	// 1. decode meta
	if err = c.DecodeMeta(ctx, message, in); err != nil {
		return err
	}

	// 2. decode payload
	return c.DecodePayload(ctx, message, in)
}

func (c *defaultCodec) Name() string {
	return "default"
}

// Select to use thrift or protobuf according to the protocol.
func (c *defaultCodec) encodePayload(ctx context.Context, message remote.Message, out remote.ByteBuffer) error {
	pCodec, err := remote.GetPayloadCodec(message)
	if err != nil {
		return err
	}
	return pCodec.Marshal(ctx, message, out)
}*/

/**
 * +------------------------------------------------------------+
 * |                  4Byte                 |       2Byte       |
 * +------------------------------------------------------------+
 * |   			     Length			    	|   HEADER MAGIC    |
 * +------------------------------------------------------------+
 */
//func IsTTHeader(flagBuf []byte) bool {
//	return binary.BigEndian.Uint32(flagBuf[Size32:])&MagicMask == TTHeaderMagic
//}

/**
 * +----------------------------------------+
 * |       2Byte        |       2Byte       |
 * +----------------------------------------+
 * |    HEADER MAGIC    |   HEADER SIZE     |
 * +----------------------------------------+
 */
//func isMeshHeader(flagBuf []byte) bool {
//	return binary.BigEndian.Uint32(flagBuf[:Size32])&MagicMask == MeshHeaderMagic
//}

/**
 * Velcro protobuf has framed field
 * +------------------------------------------------------------+
 * |                  4Byte                 |       2Byte       |
 * +------------------------------------------------------------+
 * |   			     Length			    	|   HEADER MAGIC    |
 * +------------------------------------------------------------+
 */
//func isProtobufVelcro(flagBuf []byte) bool {
//	return binary.BigEndian.Uint32(flagBuf[Size32:])&MagicMask == ProtobufV1Magic
//}

/**
 * +-------------------+
 * |       2Byte       |
 * +-------------------+
 * |   HEADER MAGIC    |
 * +-------------------
 */
//func isThriftBinary(flagBuf []byte) bool {
//	return binary.BigEndian.Uint32(flagBuf[:Size32])&MagicMask == ThriftV1Magic
//}

/**
 * +------------------------------------------------------------+
 * |                  4Byte                 |       2Byte       |
 * +------------------------------------------------------------+
 * |   			     Length			    	|   HEADER MAGIC    |
 * +------------------------------------------------------------+
 */
/*func isThriftFramedBinary(flagBuf []byte) bool {
	return binary.BigEndian.Uint32(flagBuf[Size32:])&MagicMask == ThriftV1Magic
}

func checkRPCState(ctx context.Context, message remote.Message) error {
	if message.RPCRole() == remote.Server {
		return nil
	}
	if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
		return verrors.ErrRPCFinish
	}
	if respOp, ok := ctx.Value(retry.CtxRespOp).(*int32); ok {
		if !atomic.CompareAndSwapInt32(respOp, retry.OpNo, retry.OpDoing) {
			// previous call is being handling or done
			// this flag is used to check request status in retry(backup request) scene
			return kerrors.ErrRPCFinish
		}
	}
	return nil
}

func checkPayload(flagBuf []byte, message remote.Message, in remote.ByteBuffer, isTTHeader bool, maxPayloadSize int) error {
	var transProto transport.Protocol
	var codecType serviceinfo.PayloadCodec
	if isThriftBinary(flagBuf) {
		codecType = serviceinfo.Thrift
		if isTTHeader {
			transProto = transport.TTHeader
		} else {
			transProto = transport.PurePayload
		}
	} else if isThriftFramedBinary(flagBuf) {
		codecType = serviceinfo.Thrift
		if isTTHeader {
			transProto = transport.TTHeaderFramed
		} else {
			transProto = transport.Framed
		}
		payloadLen := binary.BigEndian.Uint32(flagBuf[:Size32])
		message.SetPayloadLen(int(payloadLen))
		if err := in.Skip(Size32); err != nil {
			return err
		}
	} else if isProtobufKitex(flagBuf) {
		codecType = serviceinfo.Protobuf
		if isTTHeader {
			transProto = transport.TTHeaderFramed
		} else {
			transProto = transport.Framed
		}
		payloadLen := binary.BigEndian.Uint32(flagBuf[:Size32])
		message.SetPayloadLen(int(payloadLen))
		if err := in.Skip(Size32); err != nil {
			return err
		}
	} else {
		first4Bytes := binary.BigEndian.Uint32(flagBuf[:Size32])
		second4Bytes := binary.BigEndian.Uint32(flagBuf[Size32:])
		// 0xfff4fffd is the interrupt message of telnet
		err := perrors.NewProtocolErrorWithMsg(fmt.Sprintf("invalid payload (first4Bytes=%#x, second4Bytes=%#x)", first4Bytes, second4Bytes))
		return err
	}
	if err := checkPayloadSize(message.PayloadLen(), maxPayloadSize); err != nil {
		return err
	}
	message.SetProtocolInfo(remote.NewProtocolInfo(transProto, codecType))
	cfg := rpcinfo.AsMutableRPCConfig(message.RPCInfo().Config())
	if cfg != nil {
		tp := message.ProtocolInfo().TransProto
		cfg.SetTransportProtocol(tp)
	}
	return nil
}

func checkPayloadSize(payloadLen, maxSize int) error {
	if maxSize > 0 && payloadLen > 0 && payloadLen > maxSize {
		return perrors.NewProtocolErrorWithType(
			perrors.InvalidData,
			fmt.Sprintf("invalid data: payload size(%d) larger than the limit(%d)", payloadLen, maxSize),
		)
	}
	return nil
}
*/
