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

func Verify(name string, im *IntMapTestDataSet, result *DataFileReport) bool {
	if int32(im.size) != result.NbLines {
		logger.Errorf("Dataset %s does not have matching lines %d != %d", name, im.size, result.NbLines)
		return false
	}
	return true
}

func getAllRunnableTests() []*MapPerfTestResult {
	// Filter key types and concurrent write for non concurrent maps
	result := make([]*MapPerfTestResult, 0, len(RunConfigurations)*2)
	for _, rc := range RunConfigurations {
		for _, mt := range MapTypes {
			if !mt.isConcurrentWrite && rc.testConf.nbWriteThreads > 1 {
				// skip cannot be used
				continue
			}
			mp := MapPerfTestResult{
				runConf:     rc,
				mapTypeName: mt.name,
			}
			result = append(result, &mp)
		}
	}
	return result
}

func TestAll() bool {
	globalStopWatch := NewStopWatch()
	globalStopWatch.init()
	globalLines := 0

	allPass := true
	perfTests := getAllRunnableTests()
	for _, dc := range DataConfigurations {
		currentDataName := dc.GetDataFileName()
		im, report := ReadIntData(currentDataName, GenDataSize)
		for _, perfTest := range perfTests {
			if perfTest.runConf.dataConf.GetDataFileName() != currentDataName {
				// Not here
				continue
			}
			perfTest.fill(report)
			perfTest.testConcurrentMap(im)
			if perfTest.NbErrors() > 0 {
				allPass = false
			}
			globalLines += perfTest.nbMapEntries
		}
	}

	for _, perfTest := range perfTests {
		if !perfTest.wasDone() {
			logger.Errorf("Expected to run %s - %s test", perfTest.runConf.GetRunName(), perfTest.mapTypeName)
		}
	}

	globalStopWatch.stop()
	globalStopWatch.setNbLines(globalLines)
	globalStopWatch.display("All tests")

	dumpPerfData(perfTests)

	return allPass
}

func dumpPerfData(perfTests []*MapPerfTestResult) {
	// Mon Jan 2 15:04:05 -0700 MST 2006
	perfOutFileName := filepath.Join(utils.GetOutPerfDir(), fmt.Sprintf("maptests-%03d-%08d-%s.csv",
		len(perfTests), GenDataSize, time.Now().Format("2006-01-02_15_04_05")))

	outFile, err := os.OpenFile(perfOutFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0665)
	if err != nil {
		logger.Fatalf("Cannot create perf out file %s due to %v", perfOutFileName, err)
	}
	defer utils.CloseFile(outFile)

	utils.WriteNextString(outFile, "Idx;data;map;lines;entries;"+
		"init ratio;percent miss;write threads;read threads;nb read;"+
		"exec duration;mem;gc;errors\n")
	for i, perf := range perfTests {
		conf := perf.runConf.testConf
		diff := perf.memDiff()
		utils.WriteNextString(outFile,
			fmt.Sprintf("%2d;%s;%s;%d;%d;%f;%f;%d;%d;%d;%d;%d;%d;%d;\n",
				i, perf.runConf.GetRunName(), perf.mapTypeName, perf.dataReport.NbLines, perf.nbMapEntries,
				conf.initRatio, conf.percentMiss, conf.nbWriteThreads, conf.nbReadThreads, conf.nbReadTest*conf.nbReadThreads,
				perf.execDuration().Microseconds(), diff.TotalAlloc, diff.NumGC, perf.NbErrors()))
	}
}

func (mp *MapPerfTestResult) testConcurrentMap(im *IntMapTestDataSet) {
	m := mp.CreateMap()
	conf := mp.runConf.testConf

	mp.init()
	readWaitGroup := new(sync.WaitGroup)
	writeWaitGroup := new(sync.WaitGroup)
	doneWriting := uint32(0)
	if m.SupportConcurrentWrite() {
		size := im.size / conf.nbWriteThreads
		writeWaitGroup.Add(conf.nbWriteThreads)
		for i := 0; i < conf.nbWriteThreads; i++ {
			offset := size * i
			go testLoadAndStore(m, im, offset, size, mp, writeWaitGroup)
		}
	} else {
		writeWaitGroup.Add(1)
		testLoadAndStore(m, im, 0, im.size, mp, writeWaitGroup)
		doneWriting = uint32(1)
	}

	readWaitGroup.Add(conf.nbReadThreads)
	for i := 0; i < conf.nbReadThreads; i++ {
		go testLoad(m, im, conf.nbReadTest, &doneWriting, mp, readWaitGroup)
	}

	writeWaitGroup.Wait()
	atomic.AddUint32(&doneWriting, 1)
	readWaitGroup.Wait()

	mp.nbExpectedMapEntries = int(mp.dataReport.GetNbEntries())
	mp.nbMapEntries = m.Size()
	mp.stop()
	mp.display(mp.Name())
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
		notKey := rand.Float32() < perf.runConf.testConf.percentMiss
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
