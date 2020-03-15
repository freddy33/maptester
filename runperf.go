package maptester

import (
	"github.com/google/logger"
	"sync"
	"sync/atomic"
)

func Verify(name string, im *IntMapTestDataSet, result *MapTestResult) bool {
	if int32(im.size) != result.NbLines {
		logger.Errorf("Dataset %s does not have matching lines %d != %d", name, im.size, result.NbLines)
		return false
	}
	return true
}

func TestAll(name string) bool {
	if name == "all" {
		allPass := true
		for _, n := range FileNames {
			p := TestName(n)
			if !p {
				allPass = false
			}
		}
		return allPass
	} else {
		return TestName(name)
	}
}

func TestName(name string) bool {
	im, res := ReadIntData(name, GEN_DATA_SIZE)
	return TestMapTypes(im, res)
}

func TestMapTypes(im *IntMapTestDataSet, result *MapTestResult) bool {
	conf := MapTestConf{0.75, 4, 16}
	perf1 := TestSingleThreadMap(im, result, conf)
	perf1.display("single")
	perf2 := TestSyncMap(im, result, conf)
	perf2.display("sync.Map")
	return !perf1.errors && !perf2.errors
}

func TestSingleThreadMap(im *IntMapTestDataSet, result *MapTestResult, conf MapTestConf) *MapPerfResult {
	perf := NewPerfResult()

	expectedEntries := result.GetNbEntries()
	initSize := int(float32(expectedEntries) * conf.initSizeRatio)
	testMap := make(map[Int3Key]*TestMapValue, initSize)

	errors := 0
	for i := 0; i < im.size; i++ {
		oldValue := testMap[im.keys[i]]
		if oldValue != nil {
			atomic.AddUint32(&oldValue.count, 1)
			if im.keys[int(oldValue.val.Idx)] != im.keys[i] {
				errors++
			}
			oldValue.val = &im.values[i]
		} else {
			testMap[im.keys[i]] = &TestMapValue{&im.values[i], 0}
		}
	}
	perf.stop(len(testMap) != int(expectedEntries) || errors > 0)
	return perf
}

func TestSyncMap(im *IntMapTestDataSet, result *MapTestResult, conf MapTestConf) *MapPerfResult {
	perf := NewPerfResult()

	expectedEntries := result.GetNbEntries()
	//initSize := int(float32(expectedEntries) * conf.initSizeRatio)
	var testMap sync.Map
	errors := 0
	entries := uint32(0)

	for i := 0; i < im.size; i++ {
		result, loaded := testMap.LoadOrStore(im.keys[i], &TestMapValue{&im.values[i], 0})
		if loaded {
			oldValue := result.(*TestMapValue)
			atomic.AddUint32(&oldValue.count, 1)
			if im.keys[int(oldValue.val.Idx)] != im.keys[i] {
				errors++
			}
			if oldValue.val == &im.values[i] {
				errors++
			} else {
				oldValue.val = &im.values[i]
			}
		} else {
			atomic.AddUint32(&entries, 1)
		}
	}
	perf.stop(int(entries) != int(expectedEntries) || errors > 0)
	return perf
}
