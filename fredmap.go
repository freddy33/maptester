package maptester

import (
	"sync/atomic"
	"unsafe"
)

type hashMapEntry struct {
	key   Int3Key
	value unsafe.Pointer
	next  *hashMapEntry
}

type NonBlockConcurrentIntMap struct {
	nbElements    int32
	hashSliceSize int
	entries       []*hashMapEntry
}

func MakeNonBlockConcurrentIntMap(initSize int) *NonBlockConcurrentIntMap {
	result := new(NonBlockConcurrentIntMap)
	result.entries = make([]*hashMapEntry, initSize)
	result.hashSliceSize = len(result.entries)
	result.nbElements = 0
	return result
}

const (
	low  = 0x00000000ffffffff
	high = 0xffffffff00000000
	c1   = 0xcc9e2d51
	c2   = 0x1b873593
	r1a  = 15
	r1b  = 17
	r2a  = 13
	r2b  = 19
	m    = 4
	n    = 0xe6546b64
)

func MurmurHash(key Int3Key, size int) int {
	// Using Murmur 3 implementation
	// Found after research from https://softwareengineering.stackexchange.com/questions/49550/which-hashing-algorithm-is-best-for-uniqueness-and-speed/145633#145633?newreg=fcc6e22e2d1647e29d38f8d710248230
	h1 := uint32(0)
	for _, c := range key {
		c64 := uint64(c)
		k1 := uint32(c64 & low)
		k1 *= c1
		k1 = (k1 << r1a) | (k1 >> r1b)
		k1 *= c2
		h1 ^= k1
		h1 = (h1 << r2a) | (h1 >> r2b)
		h1 = h1*m + h1 + n

		k1 = uint32(c64 & high >> 32)
		k1 *= c1
		k1 = (k1 << r1a) | (k1 >> r1b)
		k1 *= c2
		h1 ^= k1
		h1 = (h1 << r2a) | (h1 >> r2b)
		h1 = h1*m + h1 + n
	}
	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16
	res := int(h1) % size
	if res < 0 {
		return -res
	}
	return res
}

func (n *NonBlockConcurrentIntMap) SupportConcurrentWrite() bool {
	return true
}

func (n *NonBlockConcurrentIntMap) Name() string {
	return "Non Blocking Concurrent Int Map"
}

func (n *NonBlockConcurrentIntMap) Load(key Int3Key) (*TestMapValue, bool) {
	hashIdx := MurmurHash(key, n.hashSliceSize)
	entry := n.entries[hashIdx]
	for {
		if entry == nil {
			return nil, false
		}
		if entry.key == key {
			return (*TestMapValue)(entry.value), true
		}
		entry = entry.next
	}
}

func (n *NonBlockConcurrentIntMap) Store(key Int3Key, value *TestMapValue) {
	n.internalPut(key, unsafe.Pointer(value), true)
}

func (n *NonBlockConcurrentIntMap) LoadOrStore(key Int3Key, value *TestMapValue) (*TestMapValue, bool) {
	actual, loaded := n.internalPut(key, unsafe.Pointer(value), false)
	return (*TestMapValue)(actual), loaded
}

func (n *NonBlockConcurrentIntMap) internalPut(key Int3Key, value unsafe.Pointer, overrideValue bool) (unsafe.Pointer, bool) {
	hashIdx := MurmurHash(key, n.hashSliceSize)
	for {
		actual, loaded, success := n.internalPutWithHash(hashIdx, key, value, overrideValue)
		if success {
			return actual, loaded
		}
	}
}

func (n *NonBlockConcurrentIntMap) internalPutWithHash(hashIdx int, key Int3Key, value unsafe.Pointer, overrideValue bool) (unsafe.Pointer, bool, bool) {
	entry := n.entries[hashIdx]
	entryAddr := (*unsafe.Pointer)(unsafe.Pointer(&n.entries[hashIdx]))
	for {
		if entry == nil {
			newEntry := hashMapEntry{key, value, nil}
			success := atomic.CompareAndSwapPointer(entryAddr, unsafe.Pointer(nil), unsafe.Pointer(&newEntry))
			if !success {
				return nil, false, false
			} else {
				atomic.AddInt32(&n.nbElements, 1)
				return value, false, true
			}
		} else {
			if entry.key == key {
				if overrideValue {
					success := atomic.CompareAndSwapPointer(&entry.value, entry.value, value)
					if !success {
						return nil, false, false
					} else {
						return entry.value, true, true
					}
				} else {
					return entry.value, true, true
				}
			}
			entryAddr = (*unsafe.Pointer)(unsafe.Pointer(&entry.next))
			entry = entry.next
		}
	}
}

func (n *NonBlockConcurrentIntMap) Delete(key Int3Key) {
	panic("implement me")
}

func (n *NonBlockConcurrentIntMap) Size() int {
	return int(n.nbElements)
}
