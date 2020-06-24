package maptester

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFredMapBasic(t *testing.T) {
	m := MakeNonBlockConcurrentIntMap(10)
	assert.Equal(t, 0, m.Size())
	key := Int3Key{1, 2, 3}
	val, ok := m.Load(key)
	assert.False(t, ok)
	assert.Nil(t, val)
	val = new(TestMapValue)
	val.count = 1
	val.overwritten = false
	val.val = new(TestValue)
	val.val.Idx = 45
	val.val.SVal = "test value"
	m.Store(key, val)
	assert.Equal(t, 1, m.Size())

	key2 := Int3Key{1, 2, 3}
	ret, ok := m.Load(key2)
	assert.True(t, ok)
	assert.NotNil(t, ret)
	assert.Equal(t, "test value", ret.val.SVal)
	assert.Equal(t, uint32(1), ret.count)
	assert.Equal(t, 1, m.Size())

	key3 := Int3Key{34567, 76543, 987643257}
	val2 := new(TestMapValue)
	val2.count = 1
	val2.overwritten = false
	val2.val = new(TestValue)
	val2.val.Idx = 456789
	val2.val.SVal = "test value 2"
	ret2, loaded := m.LoadOrStore(key3, val2)
	assert.False(t, loaded)
	assert.NotNil(t, ret2)
	assert.Equal(t, "test value 2", ret2.val.SVal)
	assert.Equal(t, int64(456789), ret2.val.Idx)
	assert.Equal(t, 2, m.Size())

	ret3, loaded := m.LoadOrStore(key2, val2)
	assert.True(t, loaded)
	assert.NotNil(t, ret3)
	assert.Equal(t, "test value", ret3.val.SVal)
	assert.Equal(t, uint32(1), ret3.count)
	assert.Equal(t, 2, m.Size())

	m.Store(key2, val2)
	assert.Equal(t, 2, m.Size())
	ret4, ok := m.Load(key)
	assert.True(t, ok)
	assert.NotNil(t, ret4)
	assert.Equal(t, "test value 2", ret4.val.SVal)
	assert.Equal(t, int64(456789), ret4.val.Idx)
}

func TestFredMapManyEntries(t *testing.T) {
	m := MakeNonBlockConcurrentIntMap(10)
	assert.Equal(t, 0, m.Size())
	for i := int64(2); i < 102; i++ {
		key := Int3Key{i * 2, i * i, i * i * 4}
		val := new(TestMapValue)
		val.val = new(TestValue)
		val.val.Idx = i
		val.val.SVal = fmt.Sprintf("val of i^2=%d", i*i)
		actual, loaded := m.LoadOrStore(key, val)
		assert.False(t, loaded, "key %v should not been there already", key)
		assert.Equal(t, val, actual, "key %v should have value", key)
	}
	assert.Equal(t, 100, m.Size())
}
