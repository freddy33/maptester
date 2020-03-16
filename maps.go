package maptester

import (
	"sync"
	"sync/atomic"
)

type MapKey interface {
	Hash() int
	Equal(o MapKey) bool
}

type TestMapValue struct {
	val         *TestValue
	count       uint32
	overwritten bool
}

type ConcurrentInt3Map interface {
	SupportConcurrentWrite() bool
	Name() string
	Load(key Int3Key) (*TestMapValue, bool)
	Store(key Int3Key, value *TestMapValue)
	LoadOrStore(key Int3Key, value *TestMapValue) (actual *TestMapValue, loaded bool)
	Delete(key Int3Key)
	Size() int
}

var nbMapTypes = 3

func CreateAllMaps(initSize int, withNonConcurrent bool) []ConcurrentInt3Map {
	nbMaps := nbMapTypes
	if !withNonConcurrent {
		nbMaps--
	}
	res := make([]ConcurrentInt3Map, nbMaps)

	idx := 0
	if withNonConcurrent {
		res[idx] = &BasicNonConcurrentIntMap{m: make(map[Int3Key]*TestMapValue, initSize)}
		idx++
	}
	res[idx] = &BasicConcurrentIntMap{m: make(map[Int3Key]*TestMapValue, initSize)}
	idx++
	res[idx] = &SyncIntMap{}
	idx++
	return res
}

/********************************************
TestMapValue Functions
*********************************************/

func (tmv *TestMapValue) IsOverwritten() bool {
	return tmv.overwritten
}

func (tmv *TestMapValue) overwriteVal(newVal *TestValue) {
	// Add info of overwrite count
	tmv.overwritten = true
	atomic.AddUint32(&tmv.count, 1)
	tmv.val = newVal
}

/********************************************
Non concurrent basic map
*********************************************/

type BasicNonConcurrentIntMap struct {
	m map[Int3Key]*TestMapValue
}

func (b *BasicNonConcurrentIntMap) SupportConcurrentWrite() bool {
	return false
}

func (b *BasicNonConcurrentIntMap) Name() string {
	return "Basic Map No Concurrency"
}

func (b *BasicNonConcurrentIntMap) Load(key Int3Key) (*TestMapValue, bool) {
	val, ok := b.m[key]
	return val, ok
}

func (b *BasicNonConcurrentIntMap) Store(key Int3Key, value *TestMapValue) {
	b.m[key] = value
}

func (b *BasicNonConcurrentIntMap) LoadOrStore(key Int3Key, value *TestMapValue) (actual *TestMapValue, loaded bool) {
	oldValue, ok := b.m[key]
	if ok {
		return oldValue, true
	} else {
		b.m[key] = value
		return value, false
	}
}

func (b *BasicNonConcurrentIntMap) Delete(key Int3Key) {
	delete(b.m, key)
}

func (b *BasicNonConcurrentIntMap) Size() int {
	return len(b.m)
}

/********************************************
Concurrent basic map using RWMutex
*********************************************/

type BasicConcurrentIntMap struct {
	mutex sync.RWMutex
	m     map[Int3Key]*TestMapValue
}

func (b *BasicConcurrentIntMap) SupportConcurrentWrite() bool {
	return true
}

func (b *BasicConcurrentIntMap) Name() string {
	return "Basic Concurrent Map using RWMutex"
}

func (b *BasicConcurrentIntMap) Load(key Int3Key) (*TestMapValue, bool) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	val, ok := b.m[key]
	return val, ok
}

func (b *BasicConcurrentIntMap) Store(key Int3Key, value *TestMapValue) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.m[key] = value
}

func (b *BasicConcurrentIntMap) LoadOrStore(key Int3Key, value *TestMapValue) (actual *TestMapValue, loaded bool) {
	oldValue, ok := b.Load(key)
	if ok {
		return oldValue, true
	} else {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		oldValue, ok := b.m[key]
		if ok {
			return oldValue, true
		} else {
			b.m[key] = value
			return value, false
		}
	}
}

func (b *BasicConcurrentIntMap) Delete(key Int3Key) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.m, key)
}

func (b *BasicConcurrentIntMap) Size() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return len(b.m)
}

/********************************************
Concurrent map using sync.Map
*********************************************/

type SyncIntMap struct {
	m       sync.Map
	entries uint32
}

func (s *SyncIntMap) SupportConcurrentWrite() bool {
	return true
}

func (s *SyncIntMap) Name() string {
	return "Concurrent map using sync.Map"
}

func (s *SyncIntMap) Load(key Int3Key) (*TestMapValue, bool) {
	val, ok := s.m.Load(key)
	if !ok {
		return nil, false
	}
	return val.(*TestMapValue), true
}

func (s *SyncIntMap) Store(key Int3Key, value *TestMapValue) {
	s.m.Store(key, value)
}

func (s *SyncIntMap) LoadOrStore(key Int3Key, value *TestMapValue) (*TestMapValue, bool) {
	actualVal, loaded := s.m.LoadOrStore(key, value)
	if !loaded {
		atomic.AddUint32(&s.entries, 1)
	}
	return actualVal.(*TestMapValue), loaded
}

func (s *SyncIntMap) Delete(key Int3Key) {
	s.m.Delete(key)
}

func (s *SyncIntMap) Size() int {
	if s.entries > 0 {
		return int(s.entries)
	}
	s.m.Range(func(key, value interface{}) bool {
		s.entries++
		return true
	})
	return int(s.entries)
}
