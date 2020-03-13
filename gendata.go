package maptester

import (
	"fmt"
	"github.com/freddy33/maptester/utils"
	"github.com/golang/protobuf/proto"
	"github.com/google/logger"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

//const GEN_DATA_SIZE = 10_000_000
const GEN_DATA_SIZE = 1000000

func GenAllData() {
	initMemUsage = GetMemUsage()
	initMemUsage.Print()
	genIntDataMap("noconflicts3d", GEN_DATA_SIZE, 0.0, 12)
	genIntDataMap("10conflicts3d", GEN_DATA_SIZE, 0.1, 12)
}

func genIntDataMap(name string, size int, conflictsRatio float32, valueStringSize int) {
	fmt.Printf("Generating int map %s of size %d with %v conflicts ratio and %d string length\n",
		name, size, conflictsRatio, valueStringSize)

	start := time.Now()
	im := IntMapTestDataSet{
		size,
		make([]Int3Key, size),
		make([]TestValue, size)}
	for i := 0; i < im.size; i++ {
		// Each line is a different value
		im.values[i] = TestValue{S: randomString(valueStringSize), I: rand.Int63()}

		if i > 1 && rand.Float32() < conflictsRatio {
			// Let's generate a conflict
			previousKeyIndex := int(rand.Int31n(int32(i)))
			im.keys[i] = im.keys[previousKeyIndex]
		} else {
			im.keys[i] = [3]int64{}
			for k := 0; k < 3; k++ {
				im.keys[i][k] = rand.Int63()
			}
		}
	}
	m1 := GetMemUsage()
	fmt.Println("Finished creating", im.size, "int lines in memory. Took", time.Now().Sub(start))
	m1.Print()
	fmt.Println("Difference:")
	m1.Diff(initMemUsage).Print()

	filename := filepath.Join(utils.GetGenDataDir(), fmt.Sprintf("%s-%d.data", name, size))
	fmt.Printf("Dumping int map in %s and calculating assert values\n", filename)

	start = time.Now()
	result := make(map[Int3Key]int, im.size)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0665)
	if err != nil {
		logger.Fatalf("Cannot open data file %s due to %v", filename, err)
	}
	defer file.Close()
	for i := 0; i < im.size; i++ {
		line := IntTestLine{Key: im.keys[i][:], Value: &im.values[i]}
		data, err := proto.Marshal(&line)
		if err != nil {
			logger.Fatalf("Failed to marshall %v due to %v", line, err)
		}
		utils.WriteNextBytes(file, data)
		result[im.keys[i]]++
	}
	m2 := GetMemUsage()
	fmt.Println("Finished dumping", im.size, "int lines in file ", filename, ". Took", time.Now().Sub(start))
	m2.Print()
	fmt.Println("Difference:")
	m2.Diff(m1).Print()

	max := 0
	sameKeysCount := make(map[int]int32, 5)
	for _, v := range result {
		sameKeysCount[v]++
		if v > max {
			max = v
		}
	}
	mapTestResult := new(MapTestResult)
	mapTestResult.NbLines = int32(im.size)
	mapTestResult.NbEntries = int32(len(result))
	mapTestResult.NbSameKeys = mapTestResult.NbLines - mapTestResult.NbEntries
	if max > 1 {
		mapTestResult.NbOfTimesSameKey = make([]int32, max-1)
		for k, v := range sameKeysCount {
			if k > 1 {
				mapTestResult.NbOfTimesSameKey[k-2] = v
			}
		}
	}
	fmt.Println(proto.MarshalTextString(mapTestResult))

}

func randomString(size int) string {
	cb := make([]byte, size)
	for i := 0; i < size; i++ {
		cb[i] = randomChar()
	}
	return string(cb)
}

func randomChar() byte {
	var result byte
	// 10% capital letter, 20% space, 70% lowercase
	t := rand.Float32()
	if t < 0.1 {
		result = 0x20
	} else if t < 0.3 {
		result = byte(65 + rand.Int31n(26))
	} else {
		result = byte(97 + rand.Int31n(26))
	}
	return result
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

func (m MemUsage) Print() {
	fmt.Printf("Alloc = %v", convertIfNeeded(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v", convertIfNeeded(m.TotalAlloc))
	fmt.Printf("\tSys = %v", convertIfNeeded(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

var initMemUsage MemUsage

func GetMemUsage() MemUsage {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemUsage{m.Alloc, m.TotalAlloc, m.Sys, m.NumGC}
}

func convertIfNeeded(b uint64) uint64 {
	return b
}
