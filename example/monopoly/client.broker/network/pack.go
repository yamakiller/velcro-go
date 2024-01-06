package network

type Pack struct{

}
func (p *Pack) copyBytes(source []byte, sourceOffset int32, dest []byte, destOffset, len int32) int32 {
	index := int32(0)
	for i := sourceOffset; i < sourceOffset+len; i++ {
		dest[destOffset+index] = source[sourceOffset+index]
		index++
	}
	return destOffset + len
}
func (p *Pack) writeUint8(v uint8, dest []byte, destOffset int32) int32 {
	dest[destOffset] = byte(v)
	return destOffset + 1
}
func (p *Pack) writeUint16(v uint16, dest []byte, destOffset int32) int32 {
	tmpv := v
	dest[destOffset] = byte(tmpv & 0xFF)
	dest[destOffset+1] = byte((tmpv & 0xFF00) >> 8)
	return destOffset + 2
}
func (p *Pack) writeUint32(v uint32, dest []byte, destOffset int32) int32 {
	tmpv := v
	dest[destOffset] = (byte)(tmpv & 0xFF)
	dest[destOffset+1] = (byte)((tmpv & 0xFF00) >> 8)
	dest[destOffset+2] = (byte)((int32(tmpv) & int32(0xFF0000)) >> 16)
	dest[destOffset+3] = (byte)((tmpv >> 24) & 0xFF)
	return destOffset + 4
}
func (p *Pack) writeUint64(v uint64, dest []byte, destOffset int32) int32 {
	tmpv := v
	dest[destOffset] = (byte)(tmpv & 0xFF)
	dest[destOffset+1] = (byte)((tmpv & 0xFF00) >> 8)
	dest[destOffset+2] = (byte)((int32(tmpv) & int32(0xFF0000)) >> 16)
	dest[destOffset+3] = (byte)((tmpv >> 24) & 0xFF)
	dest[destOffset+4] = (byte)((tmpv >> 32) & 0xFF)
	dest[destOffset+5] = (byte)((tmpv >> 40) & 0xFF)
	dest[destOffset+6] = (byte)((tmpv >> 48) & 0xFF)
	dest[destOffset+7] = (byte)((tmpv >> 56) & 0xFF)
	return destOffset + 8
}

func (p *Pack) Marshal(tts int64,flatDsc string,flatData []byte)([]byte,int32)  {
	buf := make([]byte, 10+len(flatDsc)+2+len(flatData))
	offset := int32(0)
	offset = p.writeUint8(1, buf, offset)
	offset = p.writeUint64(uint64(tts), buf, offset)
	offset = p.writeUint8(uint8(len(flatDsc)), buf, offset)
	offset = p.copyBytes([]byte(flatDsc), 0, buf, offset, int32(len(flatDsc)))
	offset = p.writeUint16(uint16(len(flatData)), buf, offset)
	if len(flatData) > 0 {
		offset = p.copyBytes(flatData, 0, buf, offset, int32(len(flatData)))
	}
	return buf,offset
}