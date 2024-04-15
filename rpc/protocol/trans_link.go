package protocol

import "github.com/yamakiller/velcro-go/utils/circbuf"

func NewTransLink()*TransLink{
	return &TransLink{
		LinkBuffer: circbuf.NewLinkBuffer(4096),
	}
}

type TransLink struct {
	*circbuf.LinkBuffer
}

func (t *TransLink) Write(b []byte) (int,error){
	n  ,err := t.WriteBinary(b)
	if err !=nil{
		return n, err
	}
	return n,t.Flush()
}