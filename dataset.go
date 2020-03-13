package maptester

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

type MapKey interface {
	Hash() int
	Equal(o MapKey) bool
}

func dummy() {
	d := proto.String("Hello")
	fmt.Println(d)
}

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
