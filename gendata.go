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

var Dimensions = []string{"key type", "init ratio", "conflict ratio", "r/w threads ratio", "percent miss", "r/w nb ratio", "value size"}

// Used in data generation
var ConflictRatioValues = []float32{0.0, 0.1, 0.25, 0.5}
var ValueSize = []int{5, 12, 75}

// Used in Perf Test Execution
var KeyTypes = []string{"int3d", "string10", "string25"}

// Used in Run Configuration
var InitRatioValues = []float32{0.1, 0.25, 0.5, 0.75, 1}
var NbReadThreads = []int{32, 64}
var NbWriteThreads = []int{1, 8, 16, 32}
var PercentMissValues = []float32{0.0, 0.1, 0.5}
var NbReadWriteRatio = []int{1, 16}

// Data file aggregate key type, conflict ratio and value size
var DataConfigurations map[string]*DataConfiguration
var RunConfigurations map[string]*RunConfiguration

type DataConfiguration struct {
	dataFilename  string
	keyType       string
	conflictRatio float32
	valueSize     int
}

func (dc *DataConfiguration) fillDataFileName() {
	dc.dataFilename = fmt.Sprintf("%s-c%02d-v%02d", dc.keyType, int(dc.conflictRatio*100.0), dc.valueSize)
}

func (dc *DataConfiguration) GetDataFileName() string {
	return dc.dataFilename
}

type RunConfiguration struct {
	// Aggregate data file name and all other dimensions
	runName              string
	dataConf             *DataConfiguration
	readWriteThreadRatio float32
	readWriteNbRatio     int
	testConf             *MapTestConf
}

func (rc *RunConfiguration) fillRunName() {
	rc.runName = fmt.Sprintf("%s-ir%02d-rt%02d-wt%02d-rwr%02d-m%02d", rc.dataConf.GetDataFileName(),
		int(rc.testConf.initRatio*100.0), rc.testConf.nbReadThreads, rc.testConf.nbWriteThreads,
		rc.readWriteNbRatio, int(rc.testConf.percentMiss*100.0))
}

func (rc *RunConfiguration) GetRunName() string {
	return rc.runName
}

func init() {
	DataConfigurations = make(map[string]*DataConfiguration)
	for crIdx, cr := range ConflictRatioValues {
		for _, kt := range KeyTypes {
			for vsIdx, vs := range ValueSize {
				// Testing different value size only for highest index conflicts ratio
				if vsIdx > 0 && crIdx != len(ConflictRatioValues)-1 {
					continue
				}
				dc := DataConfiguration{
					keyType:       kt,
					conflictRatio: cr,
					valueSize:     vs,
				}
				dc.fillDataFileName()
				DataConfigurations[dc.GetDataFileName()] = &dc
			}
		}
	}
	RunConfigurations = make(map[string]*RunConfiguration)
	for _, dc := range DataConfigurations {
		for _, ir := range InitRatioValues {
			for _, nbrt := range NbReadThreads {
				for _, nbwt := range NbWriteThreads {
					for _, pm := range PercentMissValues {
						for _, rwr := range NbReadWriteRatio {
							nbReadTest := int(GenDataSize * rwr / nbrt)
							rc := RunConfiguration{
								dataConf:             dc,
								readWriteThreadRatio: float32(nbrt) / float32(nbwt),
								readWriteNbRatio:     rwr,
								testConf: &MapTestConf{
									nbWriteThreads: nbwt,
									nbReadThreads:  nbrt,
									nbReadTest:     nbReadTest,
									initRatio:      ir,
									percentMiss:    pm,
								},
							}
							rc.fillRunName()
							RunConfigurations[rc.GetRunName()] = &rc
						}
					}
				}
			}
		}
	}
}

func DisplayConfigurations() {
	fmt.Printf("Generated %d data configurations\n", len(DataConfigurations))
	fmt.Printf("Generated %d run configurations\n", len(RunConfigurations))
	allTests := getAllRunnableTests()
	fmt.Printf("With maps got %d runnable tests\n", len(allTests))
}

func DeleteAllData() {
	for name, _ := range DataConfigurations {
		DeleteDataFiles(name)
	}
}

func DeleteDataFiles(name string) {
	utils.DeleteFile(getReportFilename(name, GenDataSize))
	utils.DeleteFile(getDataFilename(name, GenDataSize))
}

func GenAllData() {
	for name, dc := range DataConfigurations {
		if dc.keyType == KeyTypes[0] {
			generateIntDataMap(name, GenDataSize, dc.conflictRatio, dc.valueSize)
		}
	}
}

func getDataFilename(name string, size int) string {
	return filepath.Join(utils.GetGenDataDir(), fmt.Sprintf("%s-%d.data", name, size))
}

func getReportFilename(name string, size int) string {
	return filepath.Join(utils.GetGenDataDir(), fmt.Sprintf("%s-%d-report.data", name, size))
}

func ReadIntData(name string, size int) (*IntMapTestDataSet, *DataFileReport) {
	dataFilename := getDataFilename(name, size)
	reportFilename := getReportFilename(name, size)

	if !utils.FileExists(dataFilename) || !utils.FileExists(reportFilename) {
		logger.Errorf("Cannot read data for %s of size %d since %s or %s does not exists!",
			name, size, reportFilename, dataFilename)
		return nil, nil
	}

	fmt.Printf("Reading int map %s of size %d\n", name, size)
	//noinspection GoBoolExpressions
	if utils.Verbose {
		fmt.Printf("Using data file %q and report file %q\n", dataFilename, reportFilename)
	}

	perf := NewStopWatch()
	result := readResults(reportFilename)
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

	perf.setNbLines(im.size)
	perf.stop()
	perf.display(name)

	return im, result
}

func readResults(resultFilename string) *DataFileReport {
	resultFile, err := os.Open(resultFilename)
	if err != nil {
		logger.Fatalf("Cannot open result file %s due to %v", resultFilename, err)
	}
	defer utils.CloseFile(resultFile)
	resultData, err := ioutil.ReadAll(resultFile)
	result := new(DataFileReport)
	err = proto.Unmarshal(resultData, result)
	if err != nil {
		fmt.Printf("Got unmarshal err with data %v\n", resultData)
		logger.Fatalf("Cannot read data in result file %s due to %v", resultFilename, err)
	}
	return result
}

func generateIntDataMap(name string, size int, conflictsRatio float32, valueStringSize int) {
	resultFilename := getReportFilename(name, size)
	dataFilename := getDataFilename(name, size)

	if utils.FileExists(dataFilename) && utils.FileExists(resultFilename) {
		logger.Infof("data for %s of size %d already done in %s and %s. Skipping generation.",
			name, size, resultFilename, dataFilename)
		return
	}

	fmt.Printf("Generating int map %s of size %d with %v conflicts ratio and %d string length\n",
		name, size, conflictsRatio, valueStringSize)

	perf := NewStopWatch()
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

func writeDataFile(dataFilename string, im *IntMapTestDataSet) *DataFileReport {
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
	mapTestResult := new(DataFileReport)
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

func writeResultFile(resultsFilename string, mapTestResult *DataFileReport) int {
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
