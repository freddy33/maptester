package maptester

import (
	"fmt"
	"runtime"
	"time"
)

type MapTestConf struct {
	initSizeRatio  float32
	nbWriteThreads int
	nbReadThreads  int
}

type MapPerfResult struct {
	start         time.Time
	startMem      MemUsage
	totalExecTime time.Duration
	errors        bool
	mem           MemUsage
}

type TestMapValue struct {
	val   *TestValue
	count uint32
}

type MemUsage struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
}

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

func NewPerfResult() *MapPerfResult {
	perf := new(MapPerfResult)
	perf.init()
	return perf
}

func (mp *MapPerfResult) init() {
	mp.start = time.Now()
	mp.startMem = GetMemUsage()
	mp.errors = false
}

func (mp *MapPerfResult) stop(error bool) {
	if error {
		mp.errors = error
	}
	mp.totalExecTime = time.Now().Sub(mp.start)
	mp.mem = GetMemUsage().Diff(mp.startMem)
}

func (mp *MapPerfResult) display(name string) {
	q := "no"
	if mp.errors {
		q = "some"
	}
	fmt.Printf("%s: Took %v with %s error(s) and %d MB alloc mem\n",
		name, mp.totalExecTime, q, mp.mem.TotalAlloc/(1024*1024))
}
