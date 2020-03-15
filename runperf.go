package maptester

import (
	"fmt"
	"github.com/google/logger"
	"math/rand"
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
	initSizeRatio := float32(0.75)
	initSize := int(float32(result.GetNbEntries()) * initSizeRatio)
	maps := CreateAllMaps(initSize)

	conf := MapTestConf{nbWriteThreads: 4, nbReadThreads: 32, nbReadTest: im.size / 4}
	errors := false
	for _, m := range maps {
		p := testConcurrentMap(m, im, result, conf)
		if p.NbErrors() > 0 {
			errors = true
		}
	}
	return errors
}

func testConcurrentMap(m ConcurrentInt3Map, im *IntMapTestDataSet, result *MapTestResult, conf MapTestConf) *MapPerfResult {
	perf := NewPerfResult()

	readWaitGroup := new(sync.WaitGroup)
	writeWaitGroup := new(sync.WaitGroup)
	doneWriting := false
	if m.SupportConcurrentWrite() {
		size := im.size / conf.nbWriteThreads
		writeWaitGroup.Add(conf.nbWriteThreads)
		for i := 0; i < conf.nbWriteThreads; i++ {
			offset := size * i
			go testLoadAndStore(m, im, offset, size, perf, writeWaitGroup)
		}
	} else {
		writeWaitGroup.Add(1)
		testLoadAndStore(m, im, 0, im.size, perf, writeWaitGroup)
		doneWriting = true
	}

	readWaitGroup.Add(conf.nbReadThreads)
	for i := 0; i < conf.nbReadThreads; i++ {
		go testLoad(m, im, conf.nbReadTest, &doneWriting, perf, readWaitGroup)
	}

	writeWaitGroup.Wait()
	doneWriting = true
	readWaitGroup.Wait()

	perf.nbExpectedMapEntries = int(result.GetNbEntries())
	perf.nbMapEntries = m.Size()
	perf.stop()
	perf.display(m.Name())
	return perf
}

func testLoadAndStore(m ConcurrentInt3Map, im *IntMapTestDataSet, offset, size int, perf *MapPerfResult, wg *sync.WaitGroup) {
	errorsKeyNotSame := int32(0)
	errorsValuesEqual := int32(0)
	for i := offset; i < offset+size && i < im.size; i++ {
		key := im.keys[i]
		val := &im.values[i]
		oldValue, loaded := m.LoadOrStore(key, &TestMapValue{val: val})
		if loaded {
			if im.keys[int(oldValue.val.Idx)] != im.keys[i] {
				fmt.Println("error key conflict but not same key")
				errorsKeyNotSame++
			}
			if oldValue.val == val {
				fmt.Println("error key conflict with same value")
				errorsValuesEqual++
			} else {
				oldValue.overwriteVal(val)
			}
		}
	}
	atomic.AddInt32(&perf.errorsKeyNotSame, errorsKeyNotSame)
	atomic.AddInt32(&perf.errorsValuesEqual, errorsValuesEqual)
	wg.Done()
}

func testLoad(m ConcurrentInt3Map, im *IntMapTestDataSet, nbTest int, doneWriting *bool, perf *MapPerfResult, wg *sync.WaitGroup) {
	errorsKeyNotFound := int32(0)
	errorsValuesNotEqual := int32(0)
	errorsPointerValuesNotEqual := int32(0)
	for i := 0; i < nbTest; i++ {
		idx := int(rand.Int31n(int32(im.size)))
		value, ok := m.Load(im.keys[idx])
		if *doneWriting && !ok {
			errorsKeyNotFound++
		}
		if ok {
			if value.val.GetIdx() != int64(idx) {
				// It's an overwrite
				if !value.IsOverwriten() {
					errorsValuesNotEqual++
				}
			} else {
				// Make sure same pointer
				if value.val != &(im.values[idx]) {
					errorsPointerValuesNotEqual++
				}
			}
		}
	}
	atomic.AddInt32(&perf.errorsKeyNotFound, errorsKeyNotFound)
	atomic.AddInt32(&perf.errorsValuesNotEqual, errorsValuesNotEqual)
	atomic.AddInt32(&perf.errorsPointerValuesNotEqual, errorsPointerValuesNotEqual)
	wg.Done()
}

// Testing different scenario:
// 1. First populate, then read
//   1.1 standard map can only single thread populate but can do parallel read
//   1.2 All concurrent maps can parallel populate then parallel reads
// 2. Concurrent read write
