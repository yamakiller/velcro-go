package circbuf

import "testing"

func TestBuffer(t *testing.T) {
	ab := NewLinkBuffer(1024)

	wb := make([]byte, 8)
	wb[0] = 1
	wb[1] = 2
	wb[5] = 6
	wb[7] = 6

	ab.WriteBinary(wb)
}
