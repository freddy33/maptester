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
)

const (
	MaxConThreads     = 64
	NbLinesPerThreads = 32768
	GenDataSize       = MaxConThreads * NbLinesPerThreads
)

var FileNames = [4]string{"noconflicts3d", "10conflicts3d", "25conflicts3d", "50conflicts3d"}

func DeleteAllData() {
	for _, name := range FileNames {
		DeleteDataFiles(name)
	}
}

func DeleteDataFiles(name string) {
	utils.DeleteFile(getResultsFilename(name, GenDataSize))
	utils.DeleteFile(getDataFilename(name, GenDataSize))
}

func GenAllData() {
	generateIntDataMap(FileNames[0], GenDataSize, 0.0, 10)
	generateIntDataMap(FileNames[1], GenDataSize, 0.1, 10)
	generateIntDataMap(FileNames[2], GenDataSize, 0.25, 10)
	generateIntDataMap(FileNames[3], GenDataSize, 0.5, 10)
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

	if !utils.FileExists(dataFilename) || !utils.FileExists(resultFilename) {
		logger.Errorf("Cannot read data for %s of size %d since %s or %s does not exists!",
			name, size, resultFilename, dataFilename)
		return nil, nil
	}

	fmt.Printf("Reading int map %s of size %d\n", name, size)
	//noinspection GoBoolExpressions
	if utils.Verbose {
		fmt.Printf("Using data file '%s' and result file '%s'\n", dataFilename, resultFilename)
	}

	perf := NewPerfResult()
	result := readResults(resultFilename)
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

	// It's not a map yet just lines
	perf.nbExpectedMapEntries = im.size
	perf.nbMapEntries = im.size
	perf.stop()
	perf.display(name)

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

func generateIntDataMap(name string, size int, conflictsRatio float32, valueStringSize int) {
	resultFilename := getResultsFilename(name, size)
	dataFilename := getDataFilename(name, size)

	if utils.FileExists(dataFilename) && utils.FileExists(resultFilename) {
		logger.Infof("data for %s of size %d already done in %s and %s. Skipping generation.",
			name, size, resultFilename, dataFilename)
		return
	}

	fmt.Printf("Generating int map %s of size %d with %v conflicts ratio and %d string length\n",
		name, size, conflictsRatio, valueStringSize)

	perf := NewPerfResult()
	im := createIntMapTest(size, conflictsRatio, valueStringSize)
	perf.stop()
	perf.display(fmt.Sprintf("%s in memory %d lines", name, size))

	perf.init()
	fmt.Printf("Dumping int map in %s and calculating assert values\n", dataFilename)
	mapTestResult := writeDataFile(dataFilename, im)
	length := writeResultFile(resultFilename, mapTestResult)
	fmt.Println("Result file", resultFilename, "saved with", length)
	perf.stop()
	perf.display(fmt.Sprintf("%s saved %d lines", name, size))
}

func createIntMapTest(size int, conflictsRatio float32, valueStringSize int) *IntMapTestDataSet {
	im := IntMapTestDataSet{
		size,
		make([]Int3Key, size),
		make([]TestValue, size)}
	for i := 0; i < im.size; i++ {
		// Each line is a different value
		im.values[i] = TestValue{SVal: randomString(valueStringSize), Idx: int64(i)}

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
	offsetsPerThreads := make([]int32, MaxConThreads)
	currentPos := int32(0)
	currentThread := 0
	for i := 0; i < im.size; i++ {
		if i%NbLinesPerThreads == 0 {
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

	// Verify all the not key are not present
	for i := 0; i < im.size; i++ {
		notKey := im.getNotKey(i)
		_, ok := result[notKey]
		if ok {
			logger.Fatalf("Found a not key %v!", notKey)
		}
	}

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
