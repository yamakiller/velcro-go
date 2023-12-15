package metric

import "fmt"

// Iterator 迭代窗口内的桶.
type Iterator struct {
	count         int
	iteratedCount int
	cur           *Bucket
}

// Next 返回 true util 所有存储桶都已被迭代.
func (i *Iterator) Next() bool {
	return i.count != i.iteratedCount
}

// Bucket 获取当前桶.
func (i *Iterator) Bucket() Bucket {
	if !(i.Next()) {
		panic(fmt.Errorf("stat/metric: iteration out of range iteratedCount: %d count: %d", i.iteratedCount, i.count))
	}
	bucket := *i.cur
	i.iteratedCount++
	i.cur = i.cur.Next()
	return bucket
}
