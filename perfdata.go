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
	"strings"
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

const (
	Float32Multiplier = 1000
	NbAggregatorMaps  = 4
	InitRatioMap      = 0
	ConflictRatioMap  = 1
	NbWriteThreadsMap = 2
	NbReadThreadsMap  = 3
)

func (line *PerfLine) getMapKey(mapIdx int) int {
	switch mapIdx {
	case InitRatioMap:
		return int(line.InitRatio * Float32Multiplier)
	case ConflictRatioMap:
		return int(line.ConflictRatio * Float32Multiplier)
	case NbWriteThreadsMap:
		return line.NbWriteThreads
	case NbReadThreadsMap:
		return line.NbReadThreads
	default:
		log.Fatalf("map index %d not supported", mapIdx)
	}
	return -1
}

func getDisplayName(mapIdx int) string {
	switch mapIdx {
	case InitRatioMap:
		return "init ratio"
	case ConflictRatioMap:
		return "conflict ratio"
	case NbWriteThreadsMap:
		return "nb write threads"
	case NbReadThreadsMap:
		return "nb read threads"
	default:
		log.Fatalf("map index %d not supported", mapIdx)
	}
	return ""
}

func getDisplayFromKey(mapIdx int, key int) string {
	switch mapIdx {
	case InitRatioMap:
		return fmt.Sprintf("%.2f", float32(key)/float32(Float32Multiplier))
	case ConflictRatioMap:
		return fmt.Sprintf("%.2f", float32(key)/float32(Float32Multiplier))
	case NbWriteThreadsMap:
		return fmt.Sprintf("%d", key)
	case NbReadThreadsMap:
		return fmt.Sprintf("%d", key)
	default:
		log.Fatalf("map index %d not supported", mapIdx)
	}
	return ""
}

type Aggregator struct {
	mapType string
	total   *AggregateMeasurement
	maps    [NbAggregatorMaps]map[int]*AggregateMeasurement
}

func NewAggregator(mapTypeName string) *Aggregator {
	agg := new(Aggregator)
	agg.mapType = mapTypeName
	agg.total = new(AggregateMeasurement)
	for idx := range agg.maps {
		agg.maps[idx] = make(map[int]*AggregateMeasurement, 10)
	}
	return agg
}

func (agg *Aggregator) addMeasurement(line PerfLine) {
	if line.MapTypeName != agg.mapType {
		return
	}
	agg.total.addMeasurement(line)
	for idx, aggMap := range agg.maps {
		key := line.getMapKey(idx)
		perf, found := aggMap[key]
		if !found {
			perf = new(AggregateMeasurement)
			aggMap[key] = perf
		}
		perf.addMeasurement(line)
	}
}

func (agg *Aggregator) display() {
	fmt.Print(agg.mapType, ":")
	agg.total.display()
	for idx, aggMap := range agg.maps {
		fmt.Println(idx, len(aggMap))
		for k, v := range aggMap {
			fmt.Print(k, ":")
			v.display()
		}
	}
}

func AnalyzePerfFiles(fileNames []string) {
	aggregators := [4]*Aggregator{NewAggregator("basic"),
		NewAggregator("RWMutex"),
		NewAggregator("syncMap"),
		NewAggregator("fredMap"),
	}
	for _, filename := range fileNames {
		var file string
		if strings.ContainsRune(filename, '/') {
			file = filename
		} else {
			file = filepath.Join(utils.GetOutPerfDir(), filename)
		}
		addFileMeasurements(file, aggregators)
	}

	analysisOutFile := filepath.Join(utils.GetOutPerfDir(), fmt.Sprintf("analysis-%s.csv",
		time.Now().Format("2006-01-02_15_04_05")))
	outFile, err := os.OpenFile(analysisOutFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0665)
	if err != nil {
		logger.Fatalf("cannot create analysis out file %q due to %v", analysisOutFile, err)
	}
	defer utils.CloseFile(outFile)
	fmt.Println("Generating analysis out put in", analysisOutFile)

	utils.WriteNextString(outFile, "Analysis for")
	for _, filename := range fileNames {
		utils.WriteNextString(outFile, fmt.Sprintf(",%s", filename))
	}
	utils.WriteNextString(outFile, "\n")

	utils.WriteNextString(outFile, "map type,total,avg exec,avg mem\n")
	for _, agg := range aggregators {
		utils.WriteNextString(outFile, fmt.Sprintf("%s,total,%f,%f\n", agg.mapType, agg.total.avgExec(), agg.total.avgMem()))
	}
	utils.WriteNextString(outFile, "\n")

	for idx := InitRatioMap; idx < NbAggregatorMaps; idx++ {
		keys := make([]int, 0, len(aggregators[1].maps[idx]))
		for _, agg := range aggregators {
			for k := range agg.maps[idx] {
				keys = appendIfNotPresentInt(keys, k)
			}
		}
		sort.Ints(keys)
		utils.WriteNextString(outFile, "map type ")
		utils.WriteNextString(outFile, getDisplayName(idx))
		for _, k := range keys {
			utils.WriteNextString(outFile, fmt.Sprintf(",%s", getDisplayFromKey(idx, k)))
		}
		utils.WriteNextString(outFile, "\n")

		for _, agg := range aggregators {
			utils.WriteNextString(outFile, fmt.Sprintf("%s", agg.mapType))
			for _, k := range keys {
				val, ok := agg.maps[idx][k]
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
}

func appendIfNotPresentInt(slice []int, val int) []int {
	for _, k := range slice {
		if k == val {
			return slice
		}
	}
	return append(slice, val)
}

func addFileMeasurements(file string, aggregators [4]*Aggregator) {
	perfFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer utils.CloseFile(perfFile)
	err = gocsv.UnmarshalToCallback(perfFile, func(line PerfLine) {
		aggregators[0].addMeasurement(line)
		aggregators[1].addMeasurement(line)
		aggregators[2].addMeasurement(line)
		aggregators[3].addMeasurement(line)
	})
	if err != nil {
		log.Fatal(err)
	}
}
