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

type MapPerfResult struct {
	nbExpectedMapEntries        int
	nbMapEntries                int
	start                       time.Time
	startMem                    MemUsage
	totalExecTime               time.Duration
	errorsKeyNotFound           int32
	errorsKeyNotSame            int32
	errorsValuesEqual           int32
	errorsValuesNotEqual        int32
	errorsPointerValuesNotEqual int32
	errorsSizeNotMatch          int32
	mem                         MemUsage
}

type MemUsage struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
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
MapPerfResult Functions
*********************************************/

func NewPerfResult() *MapPerfResult {
	perf := new(MapPerfResult)
	perf.init()
	return perf
}

func (mp *MapPerfResult) init() {
	mp.start = time.Now()
	mp.startMem = GetMemUsage()
	mp.errorsKeyNotFound = 0
	mp.errorsKeyNotSame = 0
	mp.errorsValuesEqual = 0
	mp.errorsValuesNotEqual = 0
	mp.errorsPointerValuesNotEqual = 0
	mp.errorsSizeNotMatch = 0
}

func (mp *MapPerfResult) stop() {
	mp.totalExecTime = time.Now().Sub(mp.start)
	mp.mem = GetMemUsage().Diff(mp.startMem)
	if mp.nbMapEntries != mp.nbExpectedMapEntries {
		fmt.Printf("Size %d != %d\n", mp.nbMapEntries, mp.nbExpectedMapEntries)
		mp.errorsSizeNotMatch++
	}
}

func (mp *MapPerfResult) display(name string) {
	q := "no"
	if mp.NbErrors() > 0 {
		q = fmt.Sprintf("[f=%d k=%d ve=%d vn=%d pvn=%d s=%d]",
			mp.errorsKeyNotFound, mp.errorsKeyNotSame,
			mp.errorsValuesEqual, mp.errorsValuesNotEqual, mp.errorsPointerValuesNotEqual,
			mp.errorsSizeNotMatch)
	}
	fmt.Printf("%s - %d: Took %v with %s error(s) and %d MB alloc mem\n",
		name, mp.nbMapEntries, mp.totalExecTime, q, mp.mem.TotalAlloc/(1024*1024))
}

func (perf *MapPerfResult) NbErrors() int {
	return int(perf.errorsKeyNotFound + perf.errorsKeyNotSame +
		perf.errorsValuesEqual + perf.errorsValuesNotEqual + perf.errorsPointerValuesNotEqual +
		perf.errorsSizeNotMatch)
}
