package maptester

import (
	"fmt"
	"runtime"
	"time"
)

type MapTestConf struct {
	nbWriteThreads int
	nbReadThreads  int
	nbReadTest     int
}

type MemUsage struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
}

type MapPerfTestResult struct {
	conf          MapTestConf
	dataName      string
	mapTestResult *MapTestResult
	mapTypeName   string

	nbExpectedMapEntries int
	nbMapEntries         int

	startTime time.Time
	stopTime  time.Time

	startMem MemUsage
	finalMem MemUsage

	errorsKeyNotFound           int32
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
MapPerfTestResult Functions
*********************************************/

func NewPerfResult() *MapPerfTestResult {
	perf := new(MapPerfTestResult)
	perf.init()
	return perf
}

func (mp *MapPerfTestResult) init() {
	mp.startTime = time.Now()
	mp.startMem = GetMemUsage()
	mp.errorsKeyNotFound = 0
	mp.errorsKeyNotSame = 0
	mp.errorsValuesEqual = 0
	mp.errorsValuesNotEqual = 0
	mp.errorsPointerValuesNotEqual = 0
	mp.errorsSizeNotMatch = 0
}

func (mp *MapPerfTestResult) stop() {
	mp.stopTime = time.Now()
	mp.finalMem = GetMemUsage()
	if mp.nbMapEntries != mp.nbExpectedMapEntries {
		fmt.Printf("Size %d != %d\n", mp.nbMapEntries, mp.nbExpectedMapEntries)
		mp.errorsSizeNotMatch++
	}
}

func (mp *MapPerfTestResult) execDuration() time.Duration {
	return mp.stopTime.Sub(mp.startTime)
}

func (mp *MapPerfTestResult) memDiff() MemUsage {
	return mp.finalMem.Diff(mp.startMem)
}

func (mp *MapPerfTestResult) display(name string) {
	q := "no"
	if mp.NbErrors() > 0 {
		q = fmt.Sprintf("[f=%d k=%d ve=%d vn=%d pvn=%d s=%d]",
			mp.errorsKeyNotFound, mp.errorsKeyNotSame,
			mp.errorsValuesEqual, mp.errorsValuesNotEqual, mp.errorsPointerValuesNotEqual,
			mp.errorsSizeNotMatch)
	}
	fmt.Printf("%s - %d: Took %v with %s error(s) and %d MB alloc\n",
		name, mp.nbMapEntries, mp.execDuration(), q, mp.memDiff().TotalAlloc/(1024*1024))
}

func (perf *MapPerfTestResult) NbErrors() int {
	return int(perf.errorsKeyNotFound + perf.errorsKeyNotSame +
		perf.errorsValuesEqual + perf.errorsValuesNotEqual + perf.errorsPointerValuesNotEqual +
		perf.errorsSizeNotMatch)
}
