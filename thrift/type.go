package thrift

// Type constants in the Thrift protocol
type TType byte

const (
	STOP   = 0
	VOID   = 1
	BOOL   = 2
	BYTE   = 3
	I08    = 3
	DOUBLE = 4
	I16    = 6
	I32    = 8
	I64    = 10
	STRING = 11
	UTF7   = 11
	STRUCT = 12
	MAP    = 13
	SET    = 14
	LIST   = 15
	UUID   = 16
)

var typeNames = map[int]string{
	STOP:   "STOP",
	VOID:   "VOID",
	BOOL:   "BOOL",
	BYTE:   "BYTE",
	DOUBLE: "DOUBLE",
	I16:    "I16",
	I32:    "I32",
	I64:    "I64",
	STRING: "STRING",
	STRUCT: "STRUCT",
	MAP:    "MAP",
	SET:    "SET",
	LIST:   "LIST",
	UUID:   "UUID",
}

func (p TType) String() string {
	if s, ok := typeNames[int(p)]; ok {
		return s
	}
	return "Unknown"
}
