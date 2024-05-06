package utils

import (
	"unsafe"
)

// SliceByteToString 将 []byte 转换为字符串而不进行复制.
func SliceByteToString(b []byte) string {
	return unsafe.String(&b[0], len(b))
}

// StringToSliceByte 将字符串转换为 []byte 而不进行复制.
func StringToSliceByte(s string) []byte {
	/*p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)
	var b []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Data = uintptr(p)
	hdr.Cap = len(s)
	hdr.Len = len(s)
	return b*/
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
