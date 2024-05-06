package utils

// GetUIntLen 返回给定值 n 的字符串表示形式的长度.
func GetUIntLen(n uint64) int {
	cnt := 1
	for n >= 10 {
		cnt++
		n /= 10
	}
	return cnt
}
