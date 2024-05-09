package protobuf

import (
	"context"
	"errors"
	"fmt"

	"github.com/yamakiller/velcro-go/rpc/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc/pkg/remote/codec"
	"github.com/yamakiller/velcro-go/rpc/pkg/remote/codec/perrors"
	"github.com/yamakiller/velcro-go/utils/velcropb"
)

/**
 * Velcro Protobuf Protocol
 *    0 1 2 3 4 5 6 8 9 a b c d e f   0 1 2 3 4 5 6 7 8 9 a b c d e f
 *  +-------------2Byte--------------|-------------2Byte-------------+
 *  +----------------------------------------------------------------+
 *  | 0|				        PAYLOAD LENGTH 						 |
 *  +----------------------------------------------------------------+
 *  | 0|				       	SERIAL NUMBER						 |
 *  +----------------------------------------------------------------+
 *  | 0|        VG1                                  |  	MAGIC	 |
 *  +----------------------------------------------------------------+
 *  | 0|       MSG-TYPE             |       METHOD-NAME(string)      |
 */

const (
	VelcroProtobufVersion = "VP1"
)

const (
	headerFixedLength = 14
)

type protobufCodec struct{}

func (c protobufCodec) Marshal(ctx context.Context, message remote.Message, out remote.ByteBuffer) error {
	// 1.prepare info
	methodName := message.RPCInfo().Invocation().MethodName()
	if methodName == "" {
		return errors.New("methodName is empty in protobuf Marshal")
	}
	// 2.构建消息数据.
	data, err := genValidData(methodName, message)
	if err != nil {
		return err
	}

	// 3. encode header and meta
	// 3.1 sn
	if err := codec.WriteUint32(uint32(message.RPCInfo().Invocation().SeqID()), out); err != nil {
		return perrors.NewProtocolErrorWithMsg(fmt.Sprintf("protobuf marshal, write seqID failed: %s", err.Error()))
	}

	// 3.2 version and magic
	var versionAndMaigc [4]byte
	copy(versionAndMaigc[:3], []byte(VelcroProtobufVersion))
	versionAndMaigc[3] = codec.ProtobufV1Magic
	if err = codec.WriteBytes(versionAndMaigc[:], out); err != nil {
		return perrors.NewProtocolErrorWithMsg(fmt.Sprintf("protobuf marshal, write version and magic info failed: %s", err.Error()))
	}

	// 3.3 message type
	if err = codec.WriteUint16(uint16(message.MessageType()), out); err != nil {
		return perrors.NewProtocolErrorWithMsg(fmt.Sprintf("protobuf marshal, write message type info failed: %s", err.Error()))
	}

	// 3.4 methodName
	if _, err := codec.WriteString(methodName, out); err != nil {
		return perrors.NewProtocolErrorWithMsg(fmt.Sprintf("protobuf marshal, write method name failed: %s", err.Error()))
	}

	// 4. write actual message buf
	msg, ok := data.(velcropb.Writer)
	if !ok {
		return remote.NewTransErrorWithMsg(remote.ProtocolError, "velcro protobuf marshal non-velcropb")
	}

	msgsize := msg.Size()
	actualMsgBuf, err := out.Malloc(msgsize)
	if err != nil {
		perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("protobuf malloc size %d failed: %s", msgsize, err.Error()))
	}
	msg.Write(actualMsgBuf)
	return nil
}

func (c protobufCodec) Unmarshal(ctx context.Context, message remote.Message, in remote.ByteBuffer) error {
	// 在上一层已读取长度信息
	payloadLength := message.PayloadLen()
	seqID, err := codec.ReadUint32(in)
	if err != nil {
		return perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("velcro protobuf unmarshal, read seqID failed: %s", err.Error()))
	}

	versionAndMagic, err := in.Next(4)
	if err != nil {
		return err
	}

	if string(versionAndMagic[:3]) != VelcroProtobufVersion {
		return perrors.NewProtocolErrorWithType(perrors.BadVersion, "Bad version in protobuf Unmarshal")
	}

	if (versionAndMagic[3] & codec.MagicMask) != codec.ProtobufV1Magic {
		return perrors.NewProtocolErrorWithType(perrors.BadVersion, "Bad version in protobuf Unmarshal")
	}

	msgType, err := codec.ReadUint16(in)
	if err != nil {
		return perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("velcro protobuf unmarshal, read message type failed: %s", err.Error()))
	}

	if err := codec.UpdateMsgType(uint32(msgType), message); err != nil {
		return err
	}

	if err = codec.SetOrCheckSeqID(int32(seqID), message); err != nil && msgType != uint16(remote.Exception) {
		return err
	}

	methodName, methodFieldLen, err := codec.ReadString(in)
	if err != nil {
		return perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("velcro protobuf unmarshal, read method name failed: %s", err.Error()))
	}

	if err = codec.SetOrCheckMethodName(methodName, message); err != nil && msgType != uint16(remote.Exception) {
		return err
	}

	actualMsgLen := payloadLength - headerFixedLength - methodFieldLen
	actualMsgBuf, err := in.Next(actualMsgLen)
	if err != nil {
		return perrors.NewProtocolErrorWithErrMsg(err, fmt.Sprintf("velcro protobuf unmarshal, read message buffer failed: %s", err.Error()))
	}

	// exception  message
	if message.MessageType() == remote.Exception {
		var exception pbError
		if err := exception.Unmarshal(actualMsgBuf); err != nil {
			return perrors.NewProtocolErrorWithMsg(fmt.Sprintf("velcro protobuf umarshal Exception failed: %s", err.Error()))
		}
		return remote.NewTransError(exception.TypeID(), &exception)
	}

	if err = codec.NewDataIfNeeded(methodName, message); err != nil {
		return err
	}
	data := message.Data()
	msg, ok := data.(velcropb.Reader)
	if !ok {
		return remote.NewTransErrorWithMsg(remote.ProtocolError, "velcro protobuf umarshal non-velcropb")
	}

	if len(actualMsgBuf) != 0 {
		_, err := velcropb.ReadMessage(actualMsgBuf, velcropb.SkipTypeCheck, msg)
		if err != nil {
			return remote.NewTransErrorWithMsg(remote.ProtocolError, err.Error())
		}
		return nil
	}

	// 没有数据不需要解码
	return nil
}

func genValidData(methodName string, message remote.Message) (interface{}, error) {
	if err := codec.NewDataIfNeeded(methodName, message); err != nil {
		return nil, err
	}

	data := message.Data()
	if message.MessageType() != remote.Exception {
		return data, nil
	}
	transErr, isTransErr := data.(*remote.TransError)
	if !isTransErr {
		if err, isError := data.(error); isError {
			encodeErr := NewPbError(remote.InternalError, err.Error())
			return encodeErr, nil
		}
		return nil, errors.New("exception relay need error type data")
	}

	encodeErr := NewPbError(transErr.TypeID(), transErr.Error())
	return encodeErr, nil
}
