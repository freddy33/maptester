package maptester

import (
	"bufio"
	"fmt"
	"github.com/freddy33/maptester/utils"
	"github.com/golang/protobuf/proto"
	"github.com/google/logger"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	MAX_CON_THREADS      = 64
	NB_LINES_PER_THREADS = 16384 * 2 // 2^15
	GEN_DATA_SIZE        = MAX_CON_THREADS * NB_LINES_PER_THREADS
)

func GenAllData() {
	initMemUsage = GetMemUsage()
	initMemUsage.Print()
	genIntDataMap("noconflicts3d", GEN_DATA_SIZE, 0.0, 12)
	genIntDataMap("10conflicts3d", GEN_DATA_SIZE, 0.1, 12)
}

func getDataFilename(name string, size int) string {
	return filepath.Join(utils.GetGenDataDir(), fmt.Sprintf("%s-%d.data", name, size))
}

func getResultsFilename(name string, size int) string {
	return filepath.Join(utils.GetGenDataDir(), fmt.Sprintf("%s-%d-results.data", name, size))
}

func ReadIntData(name string, size int) (*IntMapTestDataSet, *MapTestResult) {
	dataFilename := getDataFilename(name, size)
	resultFilename := getResultsFilename(name, size)
	fmt.Printf("Reading int map %s of size %d from data file '%s' and result file '%s'\n",
		name, size, dataFilename, resultFilename)

	start := time.Now()
	result := readResults(resultFilename)
	fmt.Println("Reading result file", resultFilename, ". Took", time.Now().Sub(start))

	m1 := GetMemUsage()
	start = time.Now()

	im := new(IntMapTestDataSet)
	im.size = int(result.NbLines)
	im.keys = make([]Int3Key, im.size)
	im.values = make([]TestValue, im.size)

	dataFile, err := os.Open(dataFilename)
	if err != nil {
		logger.Fatalf("Cannot open data file %s due to %v", dataFilename, err)
	}
	defer utils.CloseFile(dataFile)
	dataReader := bufio.NewReaderSize(dataFile, 8192)

	for i := 0; i < im.size; i++ {
		data := utils.ReadDataBlockPrefixSize(dataReader)
		if data == nil {
			logger.Errorf("Got end of file too early in %s pos %d", dataFilename, i)
			break
		}
		imLine := new(IntTestLine)
		err = proto.Unmarshal(data, imLine)
		if err != nil {
			logger.Fatalf("Cannot read line in data file %s due to %v", dataFilename, err)
		}
		for k := 0; k < 3; k++ {
			im.keys[i][k] = imLine.GetKey()[k]
		}
		im.values[i] = *imLine.GetValue()
	}

	m2 := GetMemUsage()
	fmt.Println("Finished reading", im.size, "int lines from file ", dataFilename, ". Took", time.Now().Sub(start))
	m2.Print()
	fmt.Println("Difference:")
	m2.Diff(m1).Print()

	return im, result
}

func readResults(resultFilename string) *MapTestResult {
	resultFile, err := os.Open(resultFilename)
	if err != nil {
		logger.Fatalf("Cannot open result file %s due to %v", resultFilename, err)
	}
	defer utils.CloseFile(resultFile)
	resultData, err := ioutil.ReadAll(resultFile)
	result := new(MapTestResult)
	err = proto.Unmarshal(resultData, result)
	if err != nil {
		fmt.Printf("Got unmarshal err with data %v\n", resultData)
		logger.Fatalf("Cannot read data in result file %s due to %v", resultFilename, err)
	}
	return result
}

func genIntDataMap(name string, size int, conflictsRatio float32, valueStringSize int) {
	fmt.Printf("Generating int map %s of size %d with %v conflicts ratio and %d string length\n",
		name, size, conflictsRatio, valueStringSize)

	start := time.Now()

	im := createIntMapTest(size, conflictsRatio, valueStringSize)

	m1 := GetMemUsage()
	fmt.Println("Finished creating", im.size, "int lines in memory. Took", time.Now().Sub(start))
	m1.Print()
	fmt.Println("Difference:")
	m1.Diff(initMemUsage).Print()

	start = time.Now()

	dataFilename := getDataFilename(name, size)
	fmt.Printf("Dumping int map in %s and calculating assert values\n", dataFilename)
	mapTestResult := writeDataFile(dataFilename, im)

	m2 := GetMemUsage()
	fmt.Println("Finished dumping", im.size, "int lines in file ", dataFilename, ". Took", time.Now().Sub(start))
	m2.Print()
	fmt.Println("Difference:")
	m2.Diff(m1).Print()

	resultsFilename := getResultsFilename(name, size)
	fmt.Println("Final results of", name, "with", size, "saved in", resultsFilename)
	length := writeResultFile(resultsFilename, mapTestResult)
	fmt.Println("Result file", resultsFilename, "saved with", length)
}

func createIntMapTest(size int, conflictsRatio float32, valueStringSize int) *IntMapTestDataSet {
	im := IntMapTestDataSet{
		size,
		make([]Int3Key, size),
		make([]TestValue, size)}
	for i := 0; i < im.size; i++ {
		// Each line is a different value
		im.values[i] = TestValue{S: randomString(valueStringSize), I: rand.Int63()}

		if i > int(float32(size)*conflictsRatio)/2 && rand.Float32() < conflictsRatio {
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
	return &im
}

func writeDataFile(dataFilename string, im *IntMapTestDataSet) *MapTestResult {
	result := make(map[Int3Key]int, im.size)
	dataFile, err := os.OpenFile(dataFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0665)
	if err != nil {
		logger.Fatalf("Cannot open data file %s due to %v", dataFilename, err)
	}
	defer utils.CloseFile(dataFile)
	countsSize := make(map[byte]int, 5)
	offsetsPerThreads := make([]int32, MAX_CON_THREADS)
	currentPos := int32(0)
	currentThread := 0
	for i := 0; i < im.size; i++ {
		if i%NB_LINES_PER_THREADS == 0 {
			offsetsPerThreads[currentThread] = currentPos
			currentThread++
		}
		line := IntTestLine{Key: im.keys[i][:], Value: &im.values[i]}
		data, err := proto.Marshal(&line)
		if err != nil {
			logger.Fatalf("Failed to marshall %v due to %v", line, err)
		}
		length := utils.WriteDataBlockPrefixSize(dataFile, data)
		currentPos += int32(length) + 1
		countsSize[length]++
		result[im.keys[i]]++
	}
	fmt.Println(countsSize)

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
	mapTestResult.OffsetsPerThreads = make([]int32, len(offsetsPerThreads))
	for i, offset := range offsetsPerThreads {
		mapTestResult.OffsetsPerThreads[i] = offset
	}

	return mapTestResult
}

func writeResultFile(resultsFilename string, mapTestResult *MapTestResult) int {
	resultFile, err := os.OpenFile(resultsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0665)
	if err != nil {
		logger.Fatalf("Cannot open data file %s due to %v", resultsFilename, err)
	}
	defer utils.CloseFile(resultFile)
	data, err := proto.Marshal(mapTestResult)
	if err != nil {
		logger.Fatalf("Failed to marshall results in %s due to %v", resultsFilename, err)
	}
	return utils.WriteDataBlock(resultFile, data)
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
