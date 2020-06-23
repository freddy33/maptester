package maptester

import (
	"encoding/csv"
	"fmt"
	"github.com/freddy33/maptester/utils"
	"github.com/gocarina/gocsv"
	"github.com/google/logger"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type PerfLineIdx struct {
	Idx  int    `csv:"idx"`
	Name string `csv:"name"`
}

type PerfLineKey struct {
	KeyType              string  `csv:"key type"`
	InitRatio            float32 `csv:"init ratio"`
	ConflictRatio        float32 `csv:"conflict ratio"`
	ReadWriteThreadRatio float32 `csv:"r/w threads ratio"`
	PercentMiss          float32 `csv:"percent miss"`
	ReadWriteNbRatio     int     `csv:"r/w nb ratio"`
	ValueSize            int     `csv:"value size"`
	MapTypeName          string  `csv:"map type"`
	NbLines              int     `csv:"nb lines"`
	NbMapEntries         int     `csv:"nb map entries"`
	NbWriteThreads       int     `csv:"nb write threads"`
	NbReadThreads        int     `csv:"nb read threads"`
	NbReadTest           int     `csv:"nb read done"`
}

type PerfLineMeasurement struct {
	ExecDuration int64 `csv:"exec duration"`
	MemoryUsage  int64 `csv:"memory usage"`
	GCDone       int   `csv:"GC Done"`
	Errors       int   `csv:"errors"`
}

type PerfLine struct {
	PerfLineIdx
	PerfLineKey
	PerfLineMeasurement
}

type AggregateMeasurement struct {
	count           int
	totalLines      int64
	totalMapEntries int64
	totalReadDone   int64
	PerfLineMeasurement
}

func init() {
	gocsv.SetCSVReader(func(reader io.Reader) gocsv.CSVReader {
		r := csv.NewReader(reader)
		r.Comma = ';'
		return r
	})
}

func (agg *AggregateMeasurement) addMeasurement(line PerfLine) {
	agg.count++
	agg.totalLines += int64(line.NbLines)
	agg.totalMapEntries += int64(line.NbMapEntries)
	agg.totalReadDone += int64(line.NbReadTest)
	agg.ExecDuration += line.ExecDuration
	agg.MemoryUsage += line.MemoryUsage
	agg.GCDone += line.GCDone
	agg.Errors += line.Errors
}

func (agg *AggregateMeasurement) avgExec() float32 {
	return float32(agg.ExecDuration) / float32(agg.totalReadDone+agg.totalLines)
}

func (agg *AggregateMeasurement) avgMem() float32 {
	return float32(agg.MemoryUsage) / float32(agg.totalMapEntries)
}

func (agg *AggregateMeasurement) display() {
	fmt.Println(agg.count, agg.avgExec(), agg.avgMem())
}

type Aggregator struct {
	mapType           string
	total             *AggregateMeasurement
	perInitRatio      map[float32]*AggregateMeasurement
	perConflictRatio  map[float32]*AggregateMeasurement
	perNbWriteThreads map[int]*AggregateMeasurement
	perNbReadThreads  map[int]*AggregateMeasurement
}

func NewAggregator(mapTypeName string) *Aggregator {
	agg := new(Aggregator)
	agg.mapType = mapTypeName
	agg.total = new(AggregateMeasurement)
	agg.perInitRatio = make(map[float32]*AggregateMeasurement, 10)
	agg.perConflictRatio = make(map[float32]*AggregateMeasurement, 10)
	agg.perNbWriteThreads = make(map[int]*AggregateMeasurement, 10)
	agg.perNbReadThreads = make(map[int]*AggregateMeasurement, 10)
	return agg
}

func (agg *Aggregator) addMeasurement(line PerfLine) {
	if line.MapTypeName != agg.mapType {
		return
	}

	agg.total.addMeasurement(line)

	perf, found := agg.perInitRatio[line.InitRatio]
	if !found {
		perf = new(AggregateMeasurement)
		agg.perInitRatio[line.InitRatio] = perf
	}
	perf.addMeasurement(line)

	perf, found = agg.perConflictRatio[line.ConflictRatio]
	if !found {
		perf = new(AggregateMeasurement)
		agg.perConflictRatio[line.ConflictRatio] = perf
	}
	perf.addMeasurement(line)

	perf, found = agg.perNbWriteThreads[line.NbWriteThreads]
	if !found {
		perf = new(AggregateMeasurement)
		agg.perNbWriteThreads[line.NbWriteThreads] = perf
	}
	perf.addMeasurement(line)

	perf, found = agg.perNbReadThreads[line.NbReadThreads]
	if !found {
		perf = new(AggregateMeasurement)
		agg.perNbReadThreads[line.NbReadThreads] = perf
	}
	perf.addMeasurement(line)
}

func (agg *Aggregator) display() {
	fmt.Print(agg.mapType, ":")
	agg.total.display()
	for k, v := range agg.perInitRatio {
		fmt.Print(k, ":")
		v.display()
	}
	for k, v := range agg.perConflictRatio {
		fmt.Print(k, ":")
		v.display()
	}
	for k, v := range agg.perNbWriteThreads {
		fmt.Print(k, ":")
		v.display()
	}
	for k, v := range agg.perNbReadThreads {
		fmt.Print(k, ":")
		v.display()
	}
}

func AnalyzePerfFiles(fileNames []string) {
	aggregators := [3]*Aggregator{NewAggregator("basic"), NewAggregator("RWMutex"), NewAggregator("syncMap")}
	for _, filename := range fileNames {
		file := filepath.Join(utils.GetOutPerfDir(), filename)
		addFileMeasurements(file, aggregators)
	}

	analysisOutFile := filepath.Join(utils.GetOutPerfDir(), fmt.Sprintf("analysis-%s.csv",
		time.Now().Format("2006-01-02_15_04_05")))
	outFile, err := os.OpenFile(analysisOutFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0665)
	if err != nil {
		logger.Fatalf("cannot create analysis out file %q due to %v", analysisOutFile, err)
	}
	defer utils.CloseFile(outFile)

	utils.WriteNextString(outFile, "map type,total,avg exec,avg mem\n")
	for _, agg := range aggregators {
		utils.WriteNextString(outFile, fmt.Sprintf("%s,total,%f,%f\n", agg.mapType, agg.total.avgExec(), agg.total.avgMem()))
	}
	utils.WriteNextString(outFile, "\n")

	keysNbWriteThreads := make([]int, 0, len(aggregators[1].perNbWriteThreads))
	for _, agg := range aggregators {
		for k, _ := range agg.perNbWriteThreads {
			keysNbWriteThreads = appendIfNotPresent(keysNbWriteThreads, k)
		}
	}
	sort.Ints(keysNbWriteThreads)
	utils.WriteNextString(outFile, "map type nb write threads")
	for k := range keysNbWriteThreads {
		utils.WriteNextString(outFile, fmt.Sprintf(",%d", k))
	}
	utils.WriteNextString(outFile, "\n")
	for _, agg := range aggregators {
		utils.WriteNextString(outFile, fmt.Sprintf("%s", agg.mapType))
		for k := range keysNbWriteThreads {
			val, ok := agg.perNbWriteThreads[k]
			if ok {
				utils.WriteNextString(outFile, fmt.Sprintf(",%f", val.avgExec()))
			} else {
				utils.WriteNextString(outFile, ",")
			}
		}
		utils.WriteNextString(outFile, "\n")
	}
	utils.WriteNextString(outFile, "\n")
}

func appendIfNotPresent(slice []int, val int) []int {
	for k := range slice {
		if k == val {
			return slice
		}
	}
	return append(slice, val)
}

func addFileMeasurements(file string, aggregators [3]*Aggregator) {
	perfFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer utils.CloseFile(perfFile)
	err = gocsv.UnmarshalToCallback(perfFile, func(line PerfLine) {
		aggregators[0].addMeasurement(line)
		aggregators[1].addMeasurement(line)
		aggregators[2].addMeasurement(line)
	})
	if err != nil {
		log.Fatal(err)
	}
}
