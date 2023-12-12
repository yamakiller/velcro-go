package files

//FileHandle doc
//@Summary file handle
//@Struct FileHandle
//@Member ([]byte) file data
//@Member (int) file size
type FileHandle struct {
	_data []byte
	_size int
}

//GetBytes doc
//@Summary Returns data
//@Method GetBytes
//@Return ([]byte)
func (slf *FileHandle) GetBytes() []byte {
	return slf._data
}

//GetSize doc
//@Summary Returns data size
//@Method GetSize
//@Return (int) size
func (slf *FileHandle) GetSize() int {
	return slf._size
}
