package circbuf

// Reader 是 nocopy 读取操作的集合.
//
// 为了方便使用, 建议将Reader实现为阻塞接口,
// 而不是简单地获取缓冲区.
// 例如, 如果少于n个字节, 调用Next(n)的返回应该被阻塞, 除非超时. 保证返回值符合要求, 否则返回错误.
type Reader interface {
	// Next 返回一个包含缓冲区中接下来的 n 个字节的切片, 推进缓冲区, 就好像这些字节已由 Read 返回一样.
	//
	// 如果缓冲区中的字节数少于n, 则Next返回将被阻塞，直到数据足够或发生错误(例如等待超时).
	//
	// 切片 p 仅在下次调用 Release 方法之前有效. Next 并不是全局最优的, 具体场景推荐使
	// 用 Skip、ReadString、ReadBinary 方法.
	//
	// Return: len(p) 必须为 n 或 0, 并且 p 和 error 不能同时为 nil
	Next(n int) (p []byte, err error)

	// Peek 返回接下来的n个字节而不推进读取器, 其它行为与Next相同.
	Peek(n int) (buf []byte, err error)

	// Skip 接下来的 n 个字节并推进读取器, 这是在不使用下一个数据时更快的 Next 实现.
	Skip(n int) (err error)

	// Until 读取直到输入中第一次出现 delim，返回一个切片在输入缓冲区中以 delim 停止.
	// 如果 Until 在找到分隔符之前遇到错误,它将返回缓冲区中的所有数据以及错误本身(通常为 ErrEOF 或 ErrConnClosed).
	// 仅当行不以 delim 结尾时, 才会返回 err != nil.
	Until(delim byte) (line []byte, err error)

	// ReadString 当需要返回字符串时, Next 是更快的实现.
	// It replaces:
	//
	//  var p, err = Next(n)
	//  return string(p), err
	//
	ReadString(n int) (s string, err error)

	// ReadBinary 当需要返回不与底层共享的切片副本时，是 Next 的更快实现.
	// It replaces:
	//
	//  var p, err = Next(n)
	//  var b = make([]byte, n)
	//  copy(b, p)
	//  return b, err
	//
	ReadBinary(n int) (p []byte, err error)

	// ReadByte 当需要返回字节时，Next 是更快的实现.
	// It replaces:
	//
	//  var p, err = Next(1)
	//  return p[0], err
	//
	ReadByte() (b byte, err error)

	// Slice 返回一个新的 Reader, 其中包含该 Reader 的接下来 n 个字节.
	//
	// 如果你想使用 Next 返回的 []byte 来创建一个新的 Reader，Slice 已经做到了, 并且该操作是零拷贝的.
	// 另外, Slice也将发布这款Reader. 逻辑伪代码类似:
	//
	//  var p, err = this.Next(n)
	//  var reader = new Reader(p) // pseudocode
	//  this.Release()
	//  return reader, err
	//
	Slice(n int) (r Reader, err error)

	// Release 所有读取的切片占用的内存空间. 在确认之前读取的数据不再使用后, 需要主动执行该方法来回收内存.
	// 调用Release后,通过Next、Peek、Skip等方法获取的切片将成为无效地址, 无法再使用.
	Release() (err error)

	// Len 返回阅读器中可读数据的总长度.
	Len() (length int)
}

// Writer is a collection of operations for nocopy writes.
//
// The usage of the design is a two-step operation, first apply for a section of memory,
// fill it and then submit. E.g:
//
//	var buf, _ = Malloc(n)
//	buf = append(buf[:0], ...)
//	Flush()
//
// Note that it is not recommended to submit self-managed buffers to Writer.
// Since the writer is processed asynchronously, if the self-managed buffer is used and recycled after submission,
// it may cause inconsistent life cycle problems. Of course this is not within the scope of the design.
type Writer interface {
	// Malloc returns a slice containing the next n bytes from the buffer,
	// which will be written after submission(e.g. Flush).
	//
	// The slice p is only valid until the next submit(e.g. Flush).
	// Therefore, please make sure that all data has been written into the slice before submission.
	Malloc(n int) (buf []byte, err error)

	// WriteString is a faster implementation of Malloc when a string needs to be written.
	// It replaces:
	//
	//  var buf, err = Malloc(len(s))
	//  n = copy(buf, s)
	//  return n, err
	//
	// The argument string s will be referenced based on the original address and will not be copied,
	// so make sure that the string s will not be changed.
	WriteString(s string) (n int, err error)

	// WriteBinary is a faster implementation of Malloc when a slice needs to be written.
	// It replaces:
	//
	//  var buf, err = Malloc(len(b))
	//  n = copy(buf, b)
	//  return n, err
	//
	// The argument slice b will be referenced based on the original address and will not be copied,
	// so make sure that the slice b will not be changed.
	WriteBinary(b []byte) (n int, err error)

	// WriteByte is a faster implementation of Malloc when a byte needs to be written.
	// It replaces:
	//
	//  var buf, _ = Malloc(1)
	//  buf[0] = b
	//
	WriteByte(b byte) (err error)

	// WriteDirect is used to insert an additional slice of data on the current write stream.
	// For example, if you plan to execute:
	//
	//  var bufA, _ = Malloc(nA)
	//  WriteBinary(b)
	//  var bufB, _ = Malloc(nB)
	//
	// It can be replaced by:
	//
	//  var buf, _ = Malloc(nA+nB)
	//  WriteDirect(b, nB)
	//
	// where buf[:nA] = bufA, buf[nA:nA+nB] = bufB.
	WriteDirect(p []byte, remainCap int) error

	// MallocAck will keep the first n malloc bytes and discard the rest.
	// The following behavior:
	//
	//  var buf, _ = Malloc(8)
	//  buf = buf[:5]
	//  MallocAck(5)
	//
	// equivalent as
	//  var buf, _ = Malloc(5)
	//
	//MallocAck(n int) (err error)

	// Append the argument writer to the tail of this writer and set the argument writer to nil,
	// the operation is zero-copy, similar to p = append(p, w.p).
	Append(w Writer) (err error)

	// Flush will submit all malloc data and must confirm that the allocated bytes have been correctly assigned.
	// Its behavior is equivalent to the io.Writer hat already has parameters(slice b).
	Flush() (err error)

	// MallocLen returns the total length of the writable data that has not yet been submitted in the writer.
	MallocLen() (length int)
}
