package maptester

import (
	"fmt"
	"github.com/google/logger"
	"runtime"
	"time"
)

type MapTestConf struct {
	nbWriteThreads int
	nbReadThreads  int
	nbReadTest     int
	initRatio      float32
	percentMiss    float32
}

type MemUsage struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
}

type StopWatch interface {
	init()
	wasDone() bool
	stop()
	setNbLines(nbLines int)
	execDuration() time.Duration
	memDiff() MemUsage
	display(name string)
}

type PerfResult struct {
	startTime time.Time
	stopTime  time.Time

	startMem MemUsage
	finalMem MemUsage

	nbLines int
}

type MapPerfTestResult struct {
	runConf     *RunConfiguration
	dataReport  *DataFileReport
	mapTypeName string

	stopWatch            StopWatch
	mapInitSize          int
	nbExpectedMapEntries int
	nbMapEntries         int

	errorsKeyNotFound           int32
	errorsKeyFound              int32
	errorsKeyNotSame            int32
	errorsValuesEqual           int32
	errorsValuesNotEqual        int32
	errorsPointerValuesNotEqual int32
	errorsSizeNotMatch          int32
}

/********************************************
MemUsage Functions
*********************************************/

func (m MemUsage) Diff(o MemUsage) MemUsage {
	return MemUsage{
		m.Alloc - o.Alloc,
		m.TotalAlloc - o.TotalAlloc,
		m.Sys - o.Sys,
		m.NumGC - o.NumGC,
	}
}

func GetMemUsage() MemUsage {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemUsage{m.Alloc, m.TotalAlloc, m.Sys, m.NumGC}
}

/********************************************
PerfResult Functions
*********************************************/

var EpochZero = time.Unix(0, 0)

func NewStopWatch() StopWatch {
	perf := new(PerfResult)
	perf.init()
	return perf
}

func (pr *PerfResult) init() {
	pr.startTime = time.Now()
	pr.stopTime = EpochZero
	pr.startMem = GetMemUsage()
	pr.nbLines = 0
}

func (pr *PerfResult) setNbLines(nbLines int) {
	pr.nbLines = nbLines
}

func (pr *PerfResult) stop() {
	if pr.stopTime != EpochZero {
		logger.Warningf("Stopping stopwatch multiple time!")
	}
	pr.stopTime = time.Now()
	pr.finalMem = GetMemUsage()
}

func (pr *PerfResult) wasDone() bool {
	return pr.stopTime != EpochZero
}

func (pr *PerfResult) execDuration() time.Duration {
	if pr.stopTime == EpochZero {
		logger.Fatalf("Extracting stopwatch execution time before calling stop()!")
	}
	return pr.stopTime.Sub(pr.startTime)
}

func (pr *PerfResult) memDiff() MemUsage {
	if pr.stopTime == EpochZero {
		logger.Fatalf("Extracting stopwatch memory difference before calling stop()!")
	}
	return pr.finalMem.Diff(pr.startMem)
}

func (pr *PerfResult) display(name string) {
	if pr.stopTime == EpochZero {
		logger.Fatalf("Calling stopwatch display before calling stop()!")
	}
	fmt.Printf("%s - %d: Took %v and %d MB alloc\n",
		name, pr.nbLines, pr.execDuration(), pr.memDiff().TotalAlloc/(1024*1024))
}

/********************************************
MapPerfTestResult Functions
*********************************************/

func (mp *MapPerfTestResult) Name() string {
	return fmt.Sprintf("%s-%s", mp.runConf.GetRunName(), mp.mapTypeName)
}

func (mp *MapPerfTestResult) extractPerfLineKey() PerfLineKey {
	return PerfLineKey{
		KeyType:              mp.runConf.dataConf.keyType,
		ConflictRatio:        mp.runConf.dataConf.conflictRatio,
		ValueSize:            mp.runConf.dataConf.valueSize,
		ReadWriteThreadRatio: mp.runConf.readWriteThreadRatio,
		ReadWriteNbRatio:     mp.runConf.readWriteNbRatio,
		MapTypeName:          mp.mapTypeName,
		NbLines:              int(mp.dataReport.NbLines),
		NbMapEntries:         int(mp.dataReport.NbEntries),
		NbWriteThreads:       mp.runConf.testConf.nbWriteThreads,
		NbReadThreads:        mp.runConf.testConf.nbReadThreads,
		NbReadTest:           mp.runConf.testConf.nbReadTest,
		InitRatio:            mp.runConf.testConf.initRatio,
		PercentMiss:          mp.runConf.testConf.percentMiss,
	}
}

func readAllPerfLines(maptestsFile string) map[PerfLineKey]PerfLineMeasurement {
	result := make(map[PerfLineKey]PerfLineMeasurement)
	return result
}

func (mp *MapPerfTestResult) fill(report *DataFileReport) {
	mp.dataReport = report
	stopWatch := NewStopWatch()
	mp.stopWatch = stopWatch
	stopWatch.setNbLines(int(report.NbLines))
	mp.nbExpectedMapEntries = int(report.NbEntries)
	mp.mapInitSize = int(float32(report.NbEntries) * mp.runConf.testConf.initRatio)
}

func (mp *MapPerfTestResult) init() {
	mp.stopWatch.init()

	mp.errorsKeyNotFound = 0
	mp.errorsKeyFound = 0
	mp.errorsKeyNotSame = 0
	mp.errorsValuesEqual = 0
	mp.errorsValuesNotEqual = 0
	mp.errorsPointerValuesNotEqual = 0
	mp.errorsSizeNotMatch = 0
}

func (mp *MapPerfTestResult) wasDone() bool {
	return mp.stopWatch != nil && mp.stopWatch.wasDone()
}

func (mp *MapPerfTestResult) setNbLines(nbLines int) {
	mp.stopWatch.setNbLines(nbLines)
}

func (mp *MapPerfTestResult) stop() {
	mp.stopWatch.stop()
	if mp.nbMapEntries != mp.nbExpectedMapEntries {
		logger.Errorf("Size %d != %d\n", mp.nbMapEntries, mp.nbExpectedMapEntries)
		mp.errorsSizeNotMatch++
	}
}

func (mp *MapPerfTestResult) execDuration() time.Duration {
	return mp.stopWatch.execDuration()
}

func (mp *MapPerfTestResult) memDiff() MemUsage {
	return mp.stopWatch.memDiff()
}

func (mp *MapPerfTestResult) display(name string) {
	q := "no"
	if mp.NbErrors() > 0 {
		q = fmt.Sprintf("[nf=%d f=%d k=%d ve=%d vn=%d pvn=%d s=%d]",
			mp.errorsKeyNotFound, mp.errorsKeyFound, mp.errorsKeyNotSame,
			mp.errorsValuesEqual, mp.errorsValuesNotEqual, mp.errorsPointerValuesNotEqual,
			mp.errorsSizeNotMatch)
	}
	fmt.Printf("%s - %d: Took %v with %s error(s) and %d MB alloc\n",
		name, mp.nbMapEntries, mp.execDuration(), q, mp.memDiff().TotalAlloc/(1024*1024))
}

func (mp *MapPerfTestResult) NbErrors() int {
	return int(mp.errorsKeyNotFound + mp.errorsKeyFound + mp.errorsKeyNotSame +
		mp.errorsValuesEqual + mp.errorsValuesNotEqual + mp.errorsPointerValuesNotEqual +
		mp.errorsSizeNotMatch)
}
