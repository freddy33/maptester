package maptester

type Int3Key [3]int64
type StringKey string

type IntMapTestDataSet struct {
	size   int
	keys   []Int3Key
	values []TestValue
}

type StringMapTestDataSet struct {
	size   int
	keys   []StringKey
	values []TestValue
}
