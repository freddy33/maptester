package maptester

import (
	"fmt"
	"github.com/google/logger"
	"sync"
	"sync/atomic"
	"time"
)

type MapTestConf struct {
	initSizeRatio float32
	nbThreads     int
}

type MapPerfResult struct {
	totalExecTime time.Duration
	errors        bool
	mem           MemUsage
}

type TestMapValue struct {
	val   *TestValue
	count uint32
}

func Verify(name string, im *IntMapTestDataSet, result *MapTestResult) bool {
	if int32(im.size) != result.NbLines {
		logger.Errorf("Dataset %s does not have matching lines %d != %d", name, im.size, result.NbLines)
		return false
	}
	return true
}

func TestAll(im *IntMapTestDataSet, result *MapTestResult) bool {
	conf := MapTestConf{0.75, 16}
	perf1 := TestSingleThreadMap(im, result, conf)
	fmt.Println(perf1)
	perf2 := TestSyncMap(im, result, conf)
	fmt.Println(perf2)
	return !perf1.errors && !perf2.errors
}

func TestSingleThreadMap(im *IntMapTestDataSet, result *MapTestResult, conf MapTestConf) MapPerfResult {
	m1 := GetMemUsage()
	start := time.Now()

	expectedEntries := result.GetNbEntries()
	initSize := int(float32(expectedEntries) * conf.initSizeRatio)
	testMap := make(map[Int3Key]*TestMapValue, initSize)

	for i := 0; i < im.size; i++ {
		oldValue := testMap[im.keys[i]]
		if oldValue != nil {
			atomic.AddUint32(&oldValue.count, 1)
			oldValue.val = &im.values[i]
		} else {
			testMap[im.keys[i]] = &TestMapValue{&im.values[i], 0}
		}
	}

	return MapPerfResult{
		time.Now().Sub(start),
		len(testMap) != int(expectedEntries),
		GetMemUsage().Diff(m1),
	}
}

func TestSyncMap(im *IntMapTestDataSet, result *MapTestResult, conf MapTestConf) MapPerfResult {
	m1 := GetMemUsage()
	start := time.Now()

	expectedEntries := result.GetNbEntries()
	//initSize := int(float32(expectedEntries) * conf.initSizeRatio)
	var testMap sync.Map
	errors := false
	entries := uint32(0)

	for i := 0; i < im.size; i++ {
		result, loaded := testMap.LoadOrStore(im.keys[i], &TestMapValue{&im.values[i], 0})
		if loaded {
			oldValue := result.(*TestMapValue)
			atomic.AddUint32(&oldValue.count, 1)
			if oldValue.val == &im.values[i] {
				errors = true
			} else {
				oldValue.val = &im.values[i]
			}
		} else {
			atomic.AddUint32(&entries, 1)
		}
	}

	return MapPerfResult{
		time.Now().Sub(start),
		int(entries) != int(expectedEntries) || errors,
		GetMemUsage().Diff(m1),
	}
}
