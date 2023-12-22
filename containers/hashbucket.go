package containers

func NewHashBucket(size uint64) *Uint64Hash {
	hb := &Uint64Hash{
		tableLength:    size,
		hashIndexTable: make([]hashTable, size),
	}

	hb.initCryptTable()
	hb.initTable()

	return hb
}

// hashTable 表
type hashTable struct {
	a        uint64
	b        uint64
	isExists bool
}

type Uint64Hash struct {
	cryptTable     [0x500]uint64
	tableLength    uint64
	hashIndexTable []hashTable
}

func (h *Uint64Hash) initCryptTable() {
	seed := uint64(0x00100001)
	index1 := uint64(0)
	index2 := uint64(0)

	for index1 = 0; index1 < 0x100; index1++ {
		index2 = index1
		for i := 0; i < 5; i += 1 {
			defer func() {
				index2 += 0x100
			}()

			var (
				temp1 uint64
				temp2 uint64
			)

			seed = (seed*125 + 3) % 0x2AAAAB
			temp1 = (seed & 0xFFFF) << 0x10
			seed = (seed*125 + 3) % 0x2AAAAB
			temp2 = (seed & 0xFFFF)
			h.cryptTable[index2] = (temp1 | temp2)
		}
	}
}

func (h *Uint64Hash) initTable() {
	for i := uint64(0); i < h.tableLength; i++ {
		h.hashIndexTable[i].a = 0
		h.hashIndexTable[i].b = 0
		h.hashIndexTable[i].isExists = false
	}
}

func (h *Uint64Hash) hashUint64(key uint64) uint64 {
	key = (^key) + (key << 21)
	key = key ^ (key >> 24)
	key = (key + (key << 3)) + (key << 8)
	key = key ^ (key >> 14)
	key = (key + (key << 2)) + (key << 14)
	key = key ^ (key >> 28)
	key = key + (key << 31)
}

func (h *Uint64Hash) hashUint64A(key uint64) uint64 {
	key = (^key) + (key << 18)
	key = key ^ (key >> 31)
	key = key * 21
	key = key ^ (key >> 11)
	key = key + (key << 6)
	key = key ^ (key >> 22)
	return key
}

func (h *Uint64Hash) hashUint64B(key uint64) uint64 {
	key = (^key) + (key << 15)
	key = key ^ (key >> 12)
	key = key + (key << 2)
	key = key ^ (key >> 4)
	key = key * 2057
	key = key ^ (key >> 16)

	return key
}

func (h *Uint64Hash) Get(key uint64) (int32, bool) {
	nHash := h.hashUint64(key)
	nHashA := h.hashUint64A(key)
	nHashB := h.hashUint64B(key)

	nHashStart := (nHash & (h.tableLength - 1))
	nHashPos := nHashStart

	for {
		if !h.hashIndexTable[nHashPos].isExists {
			break
		}

		if h.hashIndexTable[nHashPos].a == nHashA &&
			h.hashIndexTable[nHashPos].b == nHashB {
			return int32(nHashPos), true
		} else {
			nHashPos = ((nHash + 1) & (h.tableLength - 1))
		}

		if nHashPos == nHashStart {
			break
		}

	}

	//未找到
	return -1, false
}

func (h *Uint64Hash) Alloc(key uint64) (int32, bool) {
	nHash := h.hashUint64(key)
	nHashA := h.hashUint64A(key)
	nHashB := h.hashUint64B(key)

	nHashStart := (nHash & (h.tableLength - 1))
	nHashPos := nHashStart

	for {
		if !h.hashIndexTable[nHashPos].isExists {
			break
		}
		nHashPos := (nHash & (h.tableLength - 1))
		if nHashPos == nHashStart {
			return -1, false
		}
	}

	h.hashIndexTable[nHashPos].isExists = true
	h.hashIndexTable[nHashPos].a = nHashA
	h.hashIndexTable[nHashPos].b = nHashB

	return int32(nHashPos), true
}
