package network

type cidKey struct {
	_address string
	_id      string
}

type ClientIDSet struct {
	_cids   []*ClientID
	_lookup map[cidKey]int
}

func (cis *ClientIDSet) key(cid *ClientID) cidKey {
	return cidKey{_address: cid.Address, _id: cid.Id}
}

func (cis *ClientIDSet) ensureInit() {
	if cis._lookup == nil {
		cis._lookup = make(map[cidKey]int)
	}
}

func (cis *ClientIDSet) indexOf(v *ClientID) int {
	if idx, ok := cis._lookup[cis.key(v)]; ok {
		return idx
	}

	return -1
}

func (cis *ClientIDSet) Contains(v *ClientID) bool {
	_, ok := cis._lookup[cis.key(v)]
	return ok
}

func (cis *ClientIDSet) Push(v *ClientID) {

	cis.ensureInit()
	if cis.Contains(v) {
		return
	}

	cis._cids = append(cis._cids, v)
	cis._lookup[cis.key(v)] = len(cis._cids) - 1
}

func (cis *ClientIDSet) Erase(v *ClientID) bool {

	cis.ensureInit()
	i := cis.indexOf(v)
	if i == -1 {
		return false
	}

	delete(cis._lookup, cis.key(v))
	if i < len(cis._cids)-1 {
		lastPID := cis._cids[len(cis._cids)-1]

		cis._cids[i] = lastPID
		cis._lookup[cis.key(lastPID)] = i
	}

	cis._cids = cis._cids[:len(cis._cids)-1]

	return true
}

func (cis *ClientIDSet) Len() int {

	return len(cis._cids)
}

// Clear removes all the elements in the set.
func (cis *ClientIDSet) Clear() {
	cis._cids = cis._cids[:0]
	cis._lookup = make(map[cidKey]int)
}

// Empty reports whether the set is empty.
func (cis *ClientIDSet) Empty() bool {

	return cis.Len() == 0
}

// Values returns all the elements of the set as a slice.
func (cis *ClientIDSet) Values() []*ClientID {

	return cis._cids
}

// ForEach invokes f for every element of the set.
func (cis *ClientIDSet) ForEach(f func(i int, cid *ClientID)) {

	for i, cid := range cis._cids {
		f(i, cid)
	}
}

func (cis *ClientIDSet) Get(index int) *ClientID {

	return cis._cids[index]
}

//
