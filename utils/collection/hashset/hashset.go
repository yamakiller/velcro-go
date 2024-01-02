// 在此存储库中，我们实现了一种基础数据结构: 基于 golang 中的 Map 的 Set
/* func main() {
	l := hashset.NewInt()

	for _, v := range []int{10, 12, 15} {
		if l.Add(v) {
			fmt.Println("hashset add", v)
		}
	}

	if l.Contains(10) {
		fmt.Println("hashset contains 10")
	}

	l.Range(func(value int) bool {
		fmt.Println("hashset range found ", value)
		return true
	})

	l.Remove(15)
	fmt.Printf("hashset contains %d items\r\n", l.Len())
}*/

package hashset

type Int64Set map[int64]struct{}

// NewInt64 returns an empty int64 set
func NewInt64() Int64Set {
	return make(map[int64]struct{})
}

// NewInt64WithSize returns an empty int64 set initialized with specific size
func NewInt64WithSize(size int) Int64Set {
	return make(map[int64]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Int64Set) Add(value int64) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Int64Set) Contains(value int64) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Int64Set) Remove(value int64) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Int64Set) Range(f func(value int64) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set
func (s Int64Set) Len() int {
	return len(s)
}
