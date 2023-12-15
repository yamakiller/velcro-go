package metric

// Bucket包含多个float64点.
type Bucket struct {
	Points []float64
	Count  int64
	next   *Bucket
}

// Append 将给定值追加到存储桶中.
func (b *Bucket) Append(val float64) {
	b.Points = append(b.Points, val)
	b.Count++
}

// Add adds 点的给定值.
func (b *Bucket) Add(offset int, val float64) {
	b.Points[offset] += val
	b.Count++
}

// Reset 清空桶.
func (b *Bucket) Reset() {
	b.Points = b.Points[:0]
	b.Count = 0
}

// Next 返回下一个桶.
func (b *Bucket) Next() *Bucket {
	return b.next
}

// Window 包含多个桶.
type Window struct {
	window []Bucket
	size   int
}

// WindowOpts 包含创建窗口的参数.
type WindowOpts struct {
	Size int
}

// NewWindow creates a new Window based on WindowOpts.
func NewWindow(opts WindowOpts) *Window {
	buckets := make([]Bucket, opts.Size)
	for offset := range buckets {
		buckets[offset] = Bucket{Points: make([]float64, 0)}
		nextOffset := offset + 1
		if nextOffset == opts.Size {
			nextOffset = 0
		}
		buckets[offset].next = &buckets[nextOffset]
	}
	return &Window{window: buckets, size: opts.Size}
}

// ResetWindow empties all buckets within the window.
func (w *Window) ResetWindow() {
	for offset := range w.window {
		w.ResetBucket(offset)
	}
}

// ResetBucket empties the bucket based on the given offset.
func (w *Window) ResetBucket(offset int) {
	w.window[offset].Reset()
}

// ResetBuckets empties the buckets based on the given offsets.
func (w *Window) ResetBuckets(offsets []int) {
	for _, offset := range offsets {
		w.ResetBucket(offset)
	}
}

// Append appends the given value to the bucket where index equals the given offset.
func (w *Window) Append(offset int, val float64) {
	w.window[offset].Append(val)
}

// Add adds the given value to the latest point within bucket where index equals the given offset.
func (w *Window) Add(offset int, val float64) {
	if w.window[offset].Count == 0 {
		w.window[offset].Append(val)
		return
	}
	w.window[offset].Add(0, val)
}

// Bucket returns the bucket where index equals the given offset.
func (w *Window) Bucket(offset int) Bucket {
	return w.window[offset]
}

// Size returns the size of the window.
func (w *Window) Size() int {
	return w.size
}

// Iterator returns the bucket iterator.
func (w *Window) Iterator(offset int, count int) Iterator {
	return Iterator{
		count: count,
		cur:   &w.window[offset],
	}
}
