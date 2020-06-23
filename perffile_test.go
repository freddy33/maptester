package maptester

import (
	"fmt"
	"github.com/freddy33/maptester/utils"
	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestReadPerfExample(t *testing.T) {
	filename := filepath.Join(utils.GetDataDir(), "perfout-examples.csv")
	perfFile, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Could not open test example file %s due to %v", filename, err)
	}
	defer utils.CloseFile(perfFile)

	results := []*PerfLine{}
	err = gocsv.UnmarshalFile(perfFile, &results)
	if err != nil {
		t.Fatalf("Unmarshall error on test example file %s due to %v", filename, err)
	}
	assert.Equal(t, 19, len(results))
	for idx, line := range results {
		assert.Equal(t, idx, line.Idx)
		assert.Equal(t, "int3d", line.KeyType)
	}
	pos, err := perfFile.Seek(0, 0)
	if err != nil {
		t.Fatalf("Seek to 0 of test example file %s due to %v", filename, err)
	}
	assert.Equal(t, int64(0), pos)
	perfPerMap := make(map[string]*AggregateMeasurement)
	aggregators := [3]*Aggregator{NewAggregator("basic"), NewAggregator("RWMutex"), NewAggregator("syncMap")}
	err = gocsv.UnmarshalToCallback(perfFile, func(line PerfLine) {
		fmt.Println(line.Idx, ":", line.Name)
		perf, found := perfPerMap[line.MapTypeName]
		if !found {
			perf = new(AggregateMeasurement)
			perfPerMap[line.MapTypeName] = perf
		}
		perf.addMeasurement(line)
		aggregators[0].addMeasurement(line)
		aggregators[1].addMeasurement(line)
		aggregators[2].addMeasurement(line)
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, perfPerMap["basic"].count)
	assert.Equal(t, 8, perfPerMap["RWMutex"].count)
	assert.Equal(t, 8, perfPerMap["syncMap"].count)
	for k, v := range perfPerMap {
		fmt.Print(k + " : ")
		v.display()
	}
	for idx, agg := range aggregators {
		fmt.Print(idx, ":")
		agg.display()
	}
}
