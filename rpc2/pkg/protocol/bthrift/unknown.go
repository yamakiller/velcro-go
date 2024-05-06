package bthrift

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/thriftgo/generator/golang/extension/unknown"
)

// UnknownField is used to describe an unknown field.
type UnknownField struct {
	Name    string
	ID      int16
	Type    int
	KeyType int
	ValType int
	Value   interface{}
}

// GetUnknownFields deserialize unknownFields stored in v to a list of *UnknownFields.
func GetUnknownFields(v interface{}) (fields []UnknownField, err error) {
	var buf []byte
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%T is not a struct type", v)
	}
	if unknownField := rv.FieldByName("_unknownFields"); !unknownField.IsValid() {
		return nil, fmt.Errorf("%T has no field named '_unknownFields'", v)
	} else {
		buf = unknownField.Bytes()
	}
	return ConvertUnknownFields(buf)
}

// ConvertUnknownFields converts buf to deserialized unknown fields.
func ConvertUnknownFields(buf unknown.Fields) (fields []UnknownField, err error) {
	if len(buf) == 0 {
		return nil, errors.New("_unknownFields is empty")
	}
	var offset int
	var l int
	var name string
	var fieldTypeId thrift.TType
	var fieldId int16
	var f UnknownField
	for {
		if offset == len(buf) {
			return
		}
		name, fieldTypeId, fieldId, l, err = Binary.ReadFieldBegin(context.Background(), buf[offset:])
		offset += l
		if err != nil {
			return nil, fmt.Errorf("read field %d begin error: %v", fieldId, err)
		}
		l, err = readUnknownField(&f, buf[offset:], name, fieldTypeId, fieldId)
		offset += l
		if err != nil {
			return nil, fmt.Errorf("read unknown field %d error: %v", fieldId, err)
		}
		fields = append(fields, f)
	}
}

func readUnknownField(f *UnknownField, buf []byte, name string, fieldType thrift.TType, id int16) (length int, err error) {
	var size int
	var l int
	f.Name = name
	f.ID = id
	f.Type = int(fieldType)
	switch fieldType {
	case thrift.BOOL:
		f.Value, l, err = Binary.ReadBool(context.Background(), buf[length:])
		length += l
	case thrift.BYTE:
		f.Value, l, err = Binary.ReadByte(context.Background(), buf[length:])
		length += l
	case thrift.I16:
		f.Value, l, err = Binary.ReadI16(context.Background(), buf[length:])
		length += l
	case thrift.I32:
		f.Value, l, err = Binary.ReadI32(context.Background(), buf[length:])
		length += l
	case thrift.I64:
		f.Value, l, err = Binary.ReadI64(context.Background(), buf[length:])
		length += l
	case thrift.DOUBLE:
		f.Value, l, err = Binary.ReadDouble(context.Background(), buf[length:])
		length += l
	case thrift.STRING:
		f.Value, l, err = Binary.ReadString(context.Background(), buf[length:])
		length += l
	case thrift.SET:
		var ttype thrift.TType
		ttype, size, l, err = Binary.ReadSetBegin(context.Background(), buf[length:])
		length += l
		if err != nil {
			return length, fmt.Errorf("read set begin error: %w", err)
		}
		f.ValType = int(ttype)
		set := make([]UnknownField, size)
		for i := 0; i < size; i++ {
			l, err2 := readUnknownField(&set[i], buf[length:], "", thrift.TType(f.ValType), int16(i))
			length += l
			if err2 != nil {
				return length, fmt.Errorf("read set elem error: %w", err2)
			}
		}
		l, err = Binary.ReadSetEnd(context.Background(), buf[length:])
		length += l
		if err != nil {
			return length, fmt.Errorf("read set end error: %w", err)
		}
		f.Value = set
	case thrift.LIST:
		var ttype thrift.TType
		ttype, size, l, err = Binary.ReadListBegin(context.Background(), buf[length:])
		length += l
		if err != nil {
			return length, fmt.Errorf("read list begin error: %w", err)
		}
		f.ValType = int(ttype)
		list := make([]UnknownField, size)
		for i := 0; i < size; i++ {
			l, err2 := readUnknownField(&list[i], buf[length:], "", thrift.TType(f.ValType), int16(i))
			length += l
			if err2 != nil {
				return length, fmt.Errorf("read list elem error: %w", err2)
			}
		}
		l, err = Binary.ReadListEnd(context.Background(), buf[length:])
		length += l
		if err != nil {
			return length, fmt.Errorf("read list end error: %w", err)
		}
		f.Value = list
	case thrift.MAP:
		var kttype, vttype thrift.TType
		kttype, vttype, size, l, err = Binary.ReadMapBegin(context.Background(), buf[length:])
		length += l
		if err != nil {
			return length, fmt.Errorf("read map begin error: %w", err)
		}
		f.KeyType = int(kttype)
		f.ValType = int(vttype)
		flatMap := make([]UnknownField, size*2)
		for i := 0; i < size; i++ {
			l, err2 := readUnknownField(&flatMap[2*i], buf[length:], "", thrift.TType(f.KeyType), int16(i))
			length += l
			if err2 != nil {
				return length, fmt.Errorf("read map key error: %w", err2)
			}
			l, err2 = readUnknownField(&flatMap[2*i+1], buf[length:], "", thrift.TType(f.ValType), int16(i))
			length += l
			if err2 != nil {
				return length, fmt.Errorf("read map value error: %w", err2)
			}
		}
		l, err = Binary.ReadMapEnd(context.Background(), buf[length:])
		length += l
		if err != nil {
			return length, fmt.Errorf("read map end error: %w", err)
		}
		f.Value = flatMap
	case thrift.STRUCT:
		_, l, err = Binary.ReadStructBegin(context.Background(), buf[length:])
		length += l
		if err != nil {
			return length, fmt.Errorf("read struct begin error: %w", err)
		}
		var field UnknownField
		var fields []UnknownField
		for {
			name, fieldTypeID, fieldID, l, err := Binary.ReadFieldBegin(context.Background(), buf[length:])
			length += l
			if err != nil {
				return length, fmt.Errorf("read field begin error: %w", err)
			}
			if fieldTypeID == thrift.STOP {
				break
			}
			l, err = readUnknownField(&field, buf[length:], name, fieldTypeID, fieldID)
			length += l
			if err != nil {
				return length, fmt.Errorf("read struct field error: %w", err)
			}
			l, err = Binary.ReadFieldEnd(context.Background(), buf[length:])
			length += l
			if err != nil {
				return length, fmt.Errorf("read field end error: %w", err)
			}
			fields = append(fields, field)
		}
		l, err = Binary.ReadStructEnd(context.Background(), buf[length:])
		length += l
		if err != nil {
			return length, fmt.Errorf("read struct end error: %w", err)
		}
		f.Value = fields
	default:
		return length, fmt.Errorf("unknown data type %d", f.Type)
	}
	if err != nil {
		return length, err
	}
	return
}

// UnknownFieldsLength returns the length of fs.
func UnknownFieldsLength(fs []UnknownField) (int, error) {
	l := 0
	for _, f := range fs {
		l += Binary.FieldBeginLength(context.Background(), f.Name, thrift.TType(f.Type), f.ID)
		ll, err := unknownFieldLength(&f)
		l += ll
		if err != nil {
			return l, err
		}
		l += Binary.FieldEndLength(context.Background())
	}
	return l, nil
}

func unknownFieldLength(f *UnknownField) (length int, err error) {
	// use constants to avoid some type assert
	switch f.Type {
	case unknown.TBool:
		length += Binary.BoolLength(context.Background(), false)
	case unknown.TByte:
		length += Binary.ByteLength(context.Background(), 0)
	case unknown.TDouble:
		length += Binary.DoubleLength(0)
	case unknown.TI16:
		length += Binary.I16Length(context.Background(), 0)
	case unknown.TI32:
		length += Binary.I32Length(context.Background(), 0)
	case unknown.TI64:
		length += Binary.I64Length(context.Background(), 0)
	case unknown.TString:
		length += Binary.StringLength(f.Value.(string))
	case unknown.TSet:
		vs := f.Value.([]UnknownField)
		length += Binary.SetBeginLength(context.Background(), thrift.TType(f.ValType), len(vs))
		for _, v := range vs {
			l, err := unknownFieldLength(&v)
			length += l
			if err != nil {
				return length, err
			}
		}
		length += Binary.SetEndLength(context.Background())
	case unknown.TList:
		vs := f.Value.([]UnknownField)
		length += Binary.ListBeginLength(context.Background(), thrift.TType(f.ValType), len(vs))
		for _, v := range vs {
			l, err := unknownFieldLength(&v)
			length += l
			if err != nil {
				return length, err
			}
		}
		length += Binary.ListEndLength(context.Background())
	case unknown.TMap:
		kvs := f.Value.([]UnknownField)
		length += Binary.MapBeginLength(context.Background(), thrift.TType(f.KeyType), thrift.TType(f.ValType), len(kvs)/2)
		for i := 0; i < len(kvs); i += 2 {
			l, err := unknownFieldLength(&kvs[i])
			length += l
			if err != nil {
				return length, err
			}
			l, err = unknownFieldLength(&kvs[i+1])
			length += l
			if err != nil {
				return length, err
			}
		}
		length += Binary.MapEndLength(context.Background())
	case unknown.TStruct:
		fs := f.Value.([]UnknownField)
		length += Binary.StructBeginLength(f.Name)
		l, err := UnknownFieldsLength(fs)
		length += l
		if err != nil {
			return length, err
		}
		length += Binary.FieldStopLength(context.Background())
		length += Binary.StructEndLength()
	default:
		return length, fmt.Errorf("unknown data type %d", f.Type)
	}
	return
}

// WriteUnknownFields writes fs into buf, and return written offset of the buf.
func WriteUnknownFields(buf []byte, fs []UnknownField) (offset int, err error) {
	for _, f := range fs {
		offset += Binary.WriteFieldBegin(context.Background(), buf[offset:], f.Name, thrift.TType(f.Type), f.ID)
		l, err := writeUnknownField(buf[offset:], &f)
		offset += l
		if err != nil {
			return offset, err
		}
		offset += Binary.WriteFieldEnd(context.Background(), buf[offset:])
	}
	return offset, nil
}

func writeUnknownField(buf []byte, f *UnknownField) (offset int, err error) {
	switch f.Type {
	case unknown.TBool:
		offset += Binary.WriteBool(context.Background(), buf, f.Value.(bool))
	case unknown.TByte:
		offset += Binary.WriteByte(context.Background(), buf, f.Value.(int8))
	case unknown.TDouble:
		offset += Binary.WriteDouble(context.Background(), buf, f.Value.(float64))
	case unknown.TI16:
		offset += Binary.WriteI16(context.Background(), buf, f.Value.(int16))
	case unknown.TI32:
		offset += Binary.WriteI32(context.Background(), buf, f.Value.(int32))
	case unknown.TI64:
		offset += Binary.WriteI64(context.Background(), buf, f.Value.(int64))
	case unknown.TString:
		offset += Binary.WriteString(context.Background(), buf, f.Value.(string))
	case unknown.TSet:
		vs := f.Value.([]UnknownField)
		offset += Binary.WriteSetBegin(context.Background(), buf, thrift.TType(f.ValType), len(vs))
		for _, v := range vs {
			l, err := writeUnknownField(buf[offset:], &v)
			offset += l
			if err != nil {
				return offset, err
			}
		}
		offset += Binary.WriteSetEnd(context.Background(), buf[offset:])
	case unknown.TList:
		vs := f.Value.([]UnknownField)
		offset += Binary.WriteListBegin(context.Background(), buf, thrift.TType(f.ValType), len(vs))
		for _, v := range vs {
			l, err := writeUnknownField(buf[offset:], &v)
			offset += l
			if err != nil {
				return offset, err
			}
		}
		offset += Binary.WriteListEnd(context.Background(), buf[offset:])
	case unknown.TMap:
		kvs := f.Value.([]UnknownField)
		offset += Binary.WriteMapBegin(context.Background(), buf, thrift.TType(f.KeyType), thrift.TType(f.ValType), len(kvs)/2)
		for i := 0; i < len(kvs); i += 2 {
			l, err := writeUnknownField(buf[offset:], &kvs[i])
			offset += l
			if err != nil {
				return offset, err
			}
			l, err = writeUnknownField(buf[offset:], &kvs[i+1])
			offset += l
			if err != nil {
				return offset, err
			}
		}
		offset += Binary.WriteMapEnd(context.Background(), buf[offset:])
	case unknown.TStruct:
		fs := f.Value.([]UnknownField)
		offset += Binary.WriteStructBegin(context.Background(), buf, f.Name)
		l, err := WriteUnknownFields(buf[offset:], fs)
		offset += l
		if err != nil {
			return offset, err
		}
		offset += Binary.WriteFieldStop(context.Background(), buf[offset:])
		offset += Binary.WriteStructEnd(context.Background(), buf[offset:])
	default:
		return offset, fmt.Errorf("unknown data type %d", f.Type)
	}
	return
}
