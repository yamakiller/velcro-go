package circbuf

type Buffer interface {
	Get(p []byte) (n int, err error)
	GetByte() (b byte, err error)
	Read(p []byte) (n int, err error)
	ReadByte() (b byte, err error)
	Write(p []byte) (n int, err error)
	WriteByte(c byte) error
	Peek(num int) (n int, err error)
	Length() int
}
