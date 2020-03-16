package maptester

import (
	"fmt"
	"github.com/freddy33/maptester/utils"
	"github.com/google/logger"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

func Verify(name string, im *IntMapTestDataSet, result *MapTestResult) bool {
	if int32(im.size) != result.NbLines {
		logger.Errorf("Dataset %s does not have matching lines %d != %d", name, im.size, result.NbLines)
		return false
	}
	return true
}

const NbConfTest = 4

var writeThreads = [NbConfTest]int{4, 8, 16, 32}

var defaultConf = MapTestConf{
	nbReadThreads: 32,
	nbReadTest:    GenDataSize / 8,
	initRatio:     0.1,
	percentMiss:   0.3,
}

func TestAll() bool {
	allPass := true

	nbDataFileNames := len(FileNames)

	// Doing only one test for not conc map
	didNotConcurrent := false
	nbTests := 1 + NbConfTest*nbDataFileNames*(nbMapTypes-1)
	perfTests := make([]*MapPerfTestResult, 0, nbTests)

	for _, name := range FileNames {

		im, result := ReadIntData(name, GenDataSize)

		for confIdx := 0; confIdx < NbConfTest; confIdx++ {
			initSize := int(float32(result.GetNbEntries()) * defaultConf.initRatio)
			maps := CreateAllMaps(initSize, !didNotConcurrent)
			for _, m := range maps {
				var conf MapTestConf
				if m.SupportConcurrentWrite() {
					conf = NewMapTestConf(defaultConf, writeThreads[confIdx])
				} else {
					conf = NewMapTestConf(defaultConf, 1)
					didNotConcurrent = true
				}

				perf := MapPerfTestResult{
					conf: conf, dataName: name, mapTestResult: result, mapTypeName: m.Name(),
				}
				perf.testConcurrentMap(m, im)
				perfTests = append(perfTests, &perf)

				if perf.NbErrors() > 0 {
					allPass = false
				}
			}
		}
	}

	if len(perfTests) != nbTests {
		logger.Errorf("Expected to run %d tests but got only %d", nbTests, len(perfTests))
		allPass = false
	}

	fmt.Printf("All test took %v\n", perfTests[len(perfTests)-1].stopTime.Sub(perfTests[0].startTime))

	dumpPerfData(perfTests)

	return allPass
}

func dumpPerfData(perfTests []*MapPerfTestResult) {
	// Mon Jan 2 15:04:05 -0700 MST 2006
	perfOutFileName := filepath.Join(utils.GetOutPerfDir(), fmt.Sprintf("map-%02d-%02d-%d-%s.csv",
		nbMapTypes, len(FileNames), GenDataSize, time.Now().Format("2006-01-02_15_04_05")))

	outFile, err := os.OpenFile(perfOutFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0665)
	if err != nil {
		logger.Fatalf("Cannot create perf out file %s due to %v", perfOutFileName, err)
	}
	defer utils.CloseFile(outFile)

	utils.WriteNextString(outFile, "Idx;data;map;lines;entries;"+
		"init ratio;percent miss;write threads;read threads;nb read;"+
		"exec duration;mem;gc;errors\n")
	for i, perf := range perfTests {
		conf := perf.conf
		diff := perf.finalMem.Diff(perf.startMem)
		utils.WriteNextString(outFile,
			fmt.Sprintf("%2d;%s;%s;%d;%d;%f;%f;%d;%d;%d;%d;%d;%d;%d;\n",
				i, perf.dataName, perf.mapTypeName, perf.mapTestResult.NbLines, perf.nbMapEntries,
				conf.initRatio, conf.percentMiss, conf.nbWriteThreads, conf.nbReadThreads, conf.nbReadTest*conf.nbReadThreads,
				perf.execDuration().Microseconds(), diff.TotalAlloc, diff.NumGC, perf.NbErrors()))
	}
}

func (perf *MapPerfTestResult) testConcurrentMap(m ConcurrentInt3Map, im *IntMapTestDataSet) {
	perf.init()
	readWaitGroup := new(sync.WaitGroup)
	writeWaitGroup := new(sync.WaitGroup)
	doneWriting := uint32(0)
	if m.SupportConcurrentWrite() {
		size := im.size / perf.conf.nbWriteThreads
		writeWaitGroup.Add(perf.conf.nbWriteThreads)
		for i := 0; i < perf.conf.nbWriteThreads; i++ {
			offset := size * i
			go testLoadAndStore(m, im, offset, size, perf, writeWaitGroup)
		}
	} else {
		writeWaitGroup.Add(1)
		testLoadAndStore(m, im, 0, im.size, perf, writeWaitGroup)
		doneWriting = uint32(1)
	}

	readWaitGroup.Add(perf.conf.nbReadThreads)
	for i := 0; i < perf.conf.nbReadThreads; i++ {
		go testLoad(m, im, perf.conf.nbReadTest, &doneWriting, perf, readWaitGroup)
	}

	writeWaitGroup.Wait()
	atomic.AddUint32(&doneWriting, 1)
	readWaitGroup.Wait()

	perf.nbExpectedMapEntries = int(perf.mapTestResult.GetNbEntries())
	perf.nbMapEntries = m.Size()
	perf.stop()
	perf.display(m.Name())
}

func testLoadAndStore(m ConcurrentInt3Map, im *IntMapTestDataSet, offset, size int, perf *MapPerfTestResult, wg *sync.WaitGroup) {
	errorsKeyNotSame := int32(0)
	errorsValuesEqual := int32(0)
	for i := offset; i < offset+size && i < im.size; i++ {
		key := im.keys[i]
		val := &im.values[i]
		oldValue, loaded := m.LoadOrStore(key, &TestMapValue{val: val})
		if loaded {
			if im.keys[int(oldValue.val.Idx)] != im.keys[i] {
				errorsKeyNotSame++
			}
			if oldValue.val == val {
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

func testLoad(m ConcurrentInt3Map, im *IntMapTestDataSet, nbTest int, doneWritingAddr *uint32, perf *MapPerfTestResult, wg *sync.WaitGroup) {
	errorsKeyFound := int32(0)
	errorsKeyNotFound := int32(0)
	errorsValuesNotEqual := int32(0)
	errorsPointerValuesNotEqual := int32(0)
	for i := 0; i < nbTest; i++ {
		idx := int(rand.Int31n(int32(im.size)))
		var key Int3Key
		notKey := rand.Float32() < perf.conf.percentMiss
		if notKey {
			key = im.getNotKey(idx)
		} else {
			key = im.getKey(idx)
		}
		value, ok := m.Load(key)
		//doneWriting := atomic.LoadUint32(doneWritingAddr) > 0
		doneWriting := *doneWritingAddr > 0

		if notKey {
			if ok {
				errorsKeyFound++
			}
		} else {
			if doneWriting && !ok {
				errorsKeyNotFound++
			}
			if ok {
				if value.val.GetIdx() != int64(idx) {
					// It's an overwrite if done writing all
					if doneWriting && !value.IsOverwritten() {
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
	}
	atomic.AddInt32(&perf.errorsKeyFound, errorsKeyFound)
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
