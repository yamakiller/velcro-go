package thrift

import (
	"encoding/base64"
	"errors"
)

// Thrift Protocol exception
type TProtocolException interface {
	TException
	TypeId() int
}

const (
	UNKNOWN_PROTOCOL_EXCEPTION = 0
	INVALID_DATA               = 1
	NEGATIVE_SIZE              = 2
	SIZE_LIMIT                 = 3
	BAD_VERSION                = 4
	NOT_IMPLEMENTED            = 5
	DEPTH_LIMIT                = 6
)

type tProtocolException struct {
	typeId int
	err    error
	msg    string
}

var _ TProtocolException = (*tProtocolException)(nil)

func (tProtocolException) TExceptionType() TExceptionType {
	return TExceptionTypeProtocol
}

func (p *tProtocolException) TypeId() int {
	return p.typeId
}

func (p *tProtocolException) String() string {
	return p.msg
}

func (p *tProtocolException) Error() string {
	return p.msg
}

func (p *tProtocolException) Unwrap() error {
	return p.err
}

func NewTProtocolException(err error) TProtocolException {
	if err == nil {
		return nil
	}

	if e, ok := err.(TProtocolException); ok {
		return e
	}

	if errors.As(err, new(base64.CorruptInputError)) {
		return NewTProtocolExceptionWithType(INVALID_DATA, err)
	}

	return NewTProtocolExceptionWithType(UNKNOWN_PROTOCOL_EXCEPTION, err)
}

func NewTProtocolExceptionWithType(errType int, err error) TProtocolException {
	if err == nil {
		return nil
	}
	return &tProtocolException{
		typeId: errType,
		err:    err,
		msg:    err.Error(),
	}
}

func prependTProtocolException(prepend string, err TProtocolException) TProtocolException {
	return &tProtocolException{
		typeId: err.TypeId(),
		err:    err,
		msg:    prepend + err.Error(),
	}
}
