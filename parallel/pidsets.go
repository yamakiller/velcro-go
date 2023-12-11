package parallel

// NewPIDSet returns a new PIDSet with the given _pids.
func NewPIDSet(_pids ...*PID) *PIDSet {
	p := &PIDSet{}
	for _, pid := range _pids {
		p.Add(pid)
	}
	return p
}

type pidKey struct {
	_address string
	_id      string
}

type PIDSet struct {
	_pids   []*PID
	_lookup map[pidKey]int
}

func (pds *PIDSet) key(pid *PID) pidKey {
	return pidKey{_address: pid.Address, _id: pid.Id}
}

func (pds *PIDSet) ensureInit() {
	if pds._lookup == nil {
		pds._lookup = make(map[pidKey]int)
	}
}

func (pds *PIDSet) indexOf(v *PID) int {
	if idx, ok := pds._lookup[pds.key(v)]; ok {
		return idx
	}

	return -1
}

func (pds *PIDSet) Contains(v *PID) bool {
	_, ok := pds._lookup[pds.key(v)]
	return ok
}

// Add adds the element v to the set.
func (pds *PIDSet) Add(v *PID) {
	pds.ensureInit()
	if pds.Contains(v) {
		return
	}

	pds._pids = append(pds._pids, v)
	pds._lookup[pds.key(v)] = len(pds._pids) - 1
}

// Remove removes v from the set and returns true if them element existed.
func (pds *PIDSet) Remove(v *PID) bool {
	pds.ensureInit()
	i := pds.indexOf(v)
	if i == -1 {
		return false
	}

	delete(pds._lookup, pds.key(v))
	if i < len(pds._pids)-1 {
		lastPID := pds._pids[len(pds._pids)-1]

		pds._pids[i] = lastPID
		pds._lookup[pds.key(lastPID)] = i
	}

	pds._pids = pds._pids[:len(pds._pids)-1]

	return true
}

// Len returns the number of elements in the set.
func (pds *PIDSet) Len() int {
	return len(pds._pids)
}

// Clear removes all the elements in the set.
func (pds *PIDSet) Clear() {
	pds._pids = pds._pids[:0]
	pds._lookup = make(map[pidKey]int)
}

// Empty reports whether the set is empty.
func (pds *PIDSet) Empty() bool {
	return pds.Len() == 0
}

// Values returns all the elements of the set as a slice.
func (pds *PIDSet) Values() []*PID {
	return pds._pids
}

// ForEach invokes f for every element of the set.
func (pds *PIDSet) ForEach(f func(i int, pid *PID)) {
	for i, pid := range pds._pids {
		f(i, pid)
	}
}

func (pds *PIDSet) Get(index int) *PID {
	return pds._pids[index]
}

func (pds *PIDSet) Clone() *PIDSet {
	return NewPIDSet(pds._pids...)
}

/*type PIDSet struct {
	_pids   []PID
	_lookup map[pidKey]uint32
	_seq    int
	_sz     int
	//_syn    mutex.SpinLock
}

type pidKey struct {
	address string
	tag     string
}



func (pds *PIDSet) toKey(address, tag string) pidKey {
	return pidKey{address: address, tag: tag}
}

func (pds *PIDSet) ensureInit() {
	if pds._lookup == nil {
		pds._lookup = make(map[pidKey]uint32)
	}
}

func NewPIDSet() *PIDSet {
	return &PIDSet{}
}

// Malloc 分配一个PID
func (pds *PIDSet) Malloc(address, tag string) (*PID, error) {
	pds.ensureInit()
	if pds.Contains(address, tag) {
		return nil, ErrorPIDExisted
	}

	pid, err := pds.next()
	if err != nil {

		return nil, err
	}

	pid.Address = address
	pid.Tag = tag
	pds._lookup[pds.toKey(address, tag)] = pid.Id

	return pid, nil
}

// Free 释放PID
func (pds *PIDSet) Free(pid *PID) error {
	pds.ensureInit()
	i, err := pds.indexOf(pid.Id)
	if err != nil {
		return err
	}

	delete(pds._lookup, pds.toKey(pid.Address, pid.Tag))

	pds._pids[i].Id = 0
	pds._pids[i].Address = ""
	pds._pids[i].Tag = ""
	pds._sz--

	return nil
}

func (pds *PIDSet) Contains(address, tag string) bool {
	_, ok := pds._lookup[pds.toKey(address, tag)]
	return ok
}

func (pds *PIDSet) Count() int {
	return pds._sz
}

func (pds *PIDSet) IsEmpty() bool {
	return pds.Count() == 0
}

// Values 在数据量大时比较低性能
func (pds *PIDSet) Values() []*PID {
	//slf.lock()
	//defer slf.unlock()

	var (
		e    error
		pid  *PID
		icur int = 0
	)

	result := make([]*PID, pds._sz)
	for _, id := range pds._lookup {
		pid, e = pds.indexOfObj(id)
		if e != nil {
			continue
		}

		result[icur] = pid
		icur++
	}

	return result
}

func (pds *PIDSet) ForEach(f func(pid *PID)) {
	var (
		e   error
		pid *PID
	)

	for _, id := range pds._lookup {
		pid, e = pds.indexOfObj(id)
		if e != nil {
			continue
		}

		f(pid)
	}
}

func (pds *PIDSet) Get(id uint32) *PID {
	pid, err := pds.indexOfObj(id)
	if err != nil {
		return nil
	}

	return pid
}

func (pds *PIDSet) Clear() {
	var (
		i uint32
		e error
	)
	for _, v := range pds._lookup {
		i, e = pds.indexOf(v)
		if e != nil {
			continue
		}
		pds._pids[i].Id = 0
		pds._pids[i].Address = ""
		pds._pids[i].Tag = ""
		pds._sz--
	}

	pds._lookup = make(map[pidKey]uint32)
}

// Next 生成新的PID
func (pds *PIDSet) next() (*PID, error) {
	for i := 0; i < math.MaxUint16; i++ {
		key := ((i + pds._seq) & math.MaxUint16)
		hash := key & (math.MaxUint16 - 1)
		if pds._pids[hash].Id == 0 {
			pds._seq = key + 1
			pds._pids[hash].Id = uint32(key)
			pds._sz++
			return &pds._pids[hash], nil
		}
	}

	return &PID{}, ErrorPIDSetFulled
}

func (pds *PIDSet) indexOf(id uint32) (uint32, error) {
	hash := uint32(id) & uint32(math.MaxUint16-1)
	if pds._pids[hash].Id != 0 && pds._pids[hash].Id == id {
		return hash, nil
	}

	return 0, ErrorPIDNotExist
}

func (pds *PIDSet) indexOfObj(id uint32) (*PID, error) {
	hash := uint32(id) & uint32(math.MaxUint16-1)
	if pds._pids[hash].Id != 0 && pds._pids[hash].Id == id {
		return &pds._pids[hash], nil
	}

	return nil, ErrorPIDNotExist
}*/

/*func (pds *PIDSet) lock() {
	try := 0
	timeDelay := time.Millisecond
	for {
		if !pds._syn.Trylock() {
			try++
			if try > 6 {
				time.Sleep(timeDelay)
				timeDelay = timeDelay * 2
				if max := 100 * time.Microsecond; timeDelay > max {
					timeDelay = max
				}
			}
			continue
		}
		break
	}
}

func (pds *PIDSet) unlock() {
	pds._syn.Unlock()
}*/
