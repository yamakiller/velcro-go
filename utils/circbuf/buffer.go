package circbuf

// Reader 是非复制读取操作的集合
// 为了方便使用,建议将Reader实现为阻塞接口,而不是简单地获取缓冲区.
// 例如, 如果少于 n 个字节，则调用 Next(n) 的返回应该被阻塞,除非超时.
// 保证返回值符合要求, 否则返回错误.
type Reader interface {
	// Next 返回包含缓冲区中接下来的 n 个字节的切片,
	// 推进缓冲区，就好像字节已由 Read 返回一样.
	//
	// 如果缓冲区中的字节数少于 n,下次返回将被阻止,直到数据足够或发生错误
	// (比如等待超时)
	//
	// 切片 p 仅在下次调用 Release 方法之前有效.
	// Next不是全局最优的,还有Skip、ReadString,对于特定场景,建议使用 ReadBinary 方法.
	//
	// Return: 在err为空时,buf不为空.
	Next(n int) (p []byte, err error)

	// Peek 返回接下来的 n 个字节而不推进读取器.
	// 其他行为与Next相同.
	Peek(n int) (buf []byte, err error)

	// 跳过接下来的 n 个字节并推进Reader,当不使用下一个数据时,这是 Next 的更快实现.
	Skip(n int) (err error)

	// Until 读取直到输入中第一次出现 delim,返回切片在输入缓冲区中以 delim 停止.
	// 如果 Until 在查找分隔符之前遇到错误, 它返回缓冲区中的所有数据以及错误本身(经常出现 ErrEOF 或 ErrConnClosed),
	// 直到仅当行不以 delim 结尾时才返回 err != nil.
	Until(delim byte) (line []byte, err error)

	// ReadString 当需要返回字符串时，Next 是更快的实现
	// 类似于:
	// 	var p, err = Next(n)
	// 	return string(p), err
	ReadString(n int) (s string, err error)
	// ReadBinary 当需要返回不与底层共享的切片副本时,是 比Next 的更快实现.
	// 类似于:
	// 	var p, err = Next(n)
	// 	var b = make([]byte, n)
	// 	copy(b, p)
	// 	return b, err
	ReadBinary(n int) (p []byte, err error)
	// Readbyte 当需要返回字节时,是比Next 的更快实现.
	// 类似于:
	//  var p, err = Next(1)
	//  return p[0], err
	ReadByte() (b byte, err error)
	// Slice 返回一个新的 Reader，其中包含该 Reader 的接下来 n 个字.
	// 类似于:
	//  var p, err = this.Next(n)
	//  var reader = new Reader(p)
	//  this.Release()
	//  return reader, err
	//
	Slice(n int) (r Reader, err error)
	// 释放内存资源
	Release() (err error)
	// Length 返回Reader总字节数.
	Length() (length int)
}

// Writer 是 nocopy 写入操作的集合.
//
// 该设计的使用是两步操作，首先申请一段内存, 然后填充例如:
//	var buf, _ = Malloc(n)
//	buf = append(buf[:0], ...)
//	Flush()
//
// 注意, 不建议将自管理缓冲区提交给Writer.
// 由于writer是异步处理的, 如果使用了自管理的buffer并在提交后回收,
// 它可能会导致不一致的生命周期问题。当然这不在设计范围之内.
type Writer interface {
	// Malloc 返回包含缓冲区中接下来的 n 个字节的切片,
	// 将在写入后提交.
	//
	// 注意, 切片 p 仅在下一次提交之前有效(如 Flush).
	// 因此, 提交前请确保所有数据已写入切片
	Malloc(n int) (buf []byte, err error)

	// WriteString 当需要写入字符串时，是 Malloc 的更快实现.
	// 类似于:
	//	var buf, err = Malloc(len(s))
	//	n = copy(buf, s)
	//  return n, err
	//
	// 注意, 参数字符串 s 将根据原始地址进行引用, 不会被复制
	// 所以要确保字符串 s 不会被改变.
	WriteString(s string) (n int, err error)

	// WriteBinary 当需要写入切片时，是 Malloc 的更快实现.
	// 类似于:
	//	var buf, err = Malloc(len(b))
	//	n = copy(buf, b)
	// return n, err
	//
	// 注意, 参数切片 b 将根据原始地址进行引用, 不会被复制
	// 所以要确保切片 b 不会被改变
	WriteBinary(b []byte) (n int, err error)

	// WriteByte 当需要写入字节时，是 Malloc 的更快实现.
	// 类似于:
	// 	var buf, _ = Malloc(1)
	//  buf[0] = b
	WriteByte(b byte) (err error)

	// WriteDirect 用于在当前写入流上插入额外的数据片.
	// 例如:
	//	var bufA, _ = Malloc(ALen)
	//  WriteBinary(b)
	//  var bufB, _ = Malloc(BLen)
	//
	// 可以替换为一下方式:
	//  var buf, _= Malloc(ALen + BLen)
	//  WriteDirect(b, BLen)
	//
	//  当 buf[:ALen] == bufA, buf[Alen:ALen + BLen] = bufB
	WriteDirect(p []byte, remainCap int) error

	// MallocAck 将保留前 n 个 malloc 字节并丢弃其余的.
	// 例如:
	//	var buf, _ = Malloc(8)
	//  buf = buf[:6]
	//  MallockAck(6)
	//
	//  与下列方式相同
	//  var buf, _ = Malloc(6)
	MallocAck(n int) (err error)

	//  Append 将参数 writer 追加到该 writer 的尾部, 并将参数 writer 设置为 nil,
	// 该操作是零拷贝的，类似于 p =append(p, w.p).
	Append(w Writer) (err error)

	// Flush 将提交所有malloc数据并且必须确认分配的字节已被正确分配.
	// 并将缓冲区数据输出到 io.Writer
	Flush() (err error)

	// MallocLen 返回writer中尚未提交的可写数据的总长度.
	MallocLen() (length int)
}

type ReadWriter interface {
	Reader
	Writer
}
