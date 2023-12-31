package hashset

type Float32Set map[float32]struct{}

// NewFloat32 returns an empty float32 set
func NewFloat32() Float32Set {
	return make(map[float32]struct{})
}

// NewFloat32WithSize returns an empty float32 set initialized with specific size
func NewFloat32WithSize(size int) Float32Set {
	return make(map[float32]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Float32Set) Add(value float32) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Float32Set) Contains(value float32) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Float32Set) Remove(value float32) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Float32Set) Range(f func(value float32) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s Float32Set) Len() int {
	return len(s)
}

type Float64Set map[float64]struct{}

// NewFloat64 returns an empty float64 set
func NewFloat64() Float64Set {
	return make(map[float64]struct{})
}

// NewFloat64WithSize returns an empty float64 set initialized with specific size
func NewFloat64WithSize(size int) Float64Set {
	return make(map[float64]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Float64Set) Add(value float64) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Float64Set) Contains(value float64) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Float64Set) Remove(value float64) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Float64Set) Range(f func(value float64) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s Float64Set) Len() int {
	return len(s)
}

type Int32Set map[int32]struct{}

// NewInt32 returns an empty int32 set
func NewInt32() Int32Set {
	return make(map[int32]struct{})
}

// NewInt32WithSize returns an empty int32 set initialized with specific size
func NewInt32WithSize(size int) Int32Set {
	return make(map[int32]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Int32Set) Add(value int32) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Int32Set) Contains(value int32) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Int32Set) Remove(value int32) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Int32Set) Range(f func(value int32) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s Int32Set) Len() int {
	return len(s)
}

type Int16Set map[int16]struct{}

// NewInt16 returns an empty int16 set
func NewInt16() Int16Set {
	return make(map[int16]struct{})
}

// NewInt16WithSize returns an empty int16 set initialized with specific size
func NewInt16WithSize(size int) Int16Set {
	return make(map[int16]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Int16Set) Add(value int16) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Int16Set) Contains(value int16) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Int16Set) Remove(value int16) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Int16Set) Range(f func(value int16) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s Int16Set) Len() int {
	return len(s)
}

type IntSet map[int]struct{}

// NewInt returns an empty int set
func NewInt() IntSet {
	return make(map[int]struct{})
}

// NewIntWithSize returns an empty int set initialized with specific size
func NewIntWithSize(size int) IntSet {
	return make(map[int]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s IntSet) Add(value int) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s IntSet) Contains(value int) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s IntSet) Remove(value int) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s IntSet) Range(f func(value int) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s IntSet) Len() int {
	return len(s)
}

type Uint64Set map[uint64]struct{}

// NewUint64 returns an empty uint64 set
func NewUint64() Uint64Set {
	return make(map[uint64]struct{})
}

// NewUint64WithSize returns an empty uint64 set initialized with specific size
func NewUint64WithSize(size int) Uint64Set {
	return make(map[uint64]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Uint64Set) Add(value uint64) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Uint64Set) Contains(value uint64) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Uint64Set) Remove(value uint64) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Uint64Set) Range(f func(value uint64) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s Uint64Set) Len() int {
	return len(s)
}

type Uint32Set map[uint32]struct{}

// NewUint32 returns an empty uint32 set
func NewUint32() Uint32Set {
	return make(map[uint32]struct{})
}

// NewUint32WithSize returns an empty uint32 set initialized with specific size
func NewUint32WithSize(size int) Uint32Set {
	return make(map[uint32]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Uint32Set) Add(value uint32) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Uint32Set) Contains(value uint32) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Uint32Set) Remove(value uint32) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Uint32Set) Range(f func(value uint32) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s Uint32Set) Len() int {
	return len(s)
}

type Uint16Set map[uint16]struct{}

// NewUint16 returns an empty uint16 set
func NewUint16() Uint16Set {
	return make(map[uint16]struct{})
}

// NewUint16WithSize returns an empty uint16 set initialized with specific size
func NewUint16WithSize(size int) Uint16Set {
	return make(map[uint16]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Uint16Set) Add(value uint16) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Uint16Set) Contains(value uint16) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Uint16Set) Remove(value uint16) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Uint16Set) Range(f func(value uint16) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s Uint16Set) Len() int {
	return len(s)
}

type UintSet map[uint]struct{}

// NewUint returns an empty uint set
func NewUint() UintSet {
	return make(map[uint]struct{})
}

// NewUintWithSize returns an empty uint set initialized with specific size
func NewUintWithSize(size int) UintSet {
	return make(map[uint]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s UintSet) Add(value uint) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s UintSet) Contains(value uint) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s UintSet) Remove(value uint) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s UintSet) Range(f func(value uint) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set

func (s UintSet) Len() int {
	return len(s)
}
