package maptester

type Int3Key [3]int64
type StringKey string

type IntMapTestDataSet struct {
	size   int
	keys   []Int3Key
	values []TestValue
}

func (im *IntMapTestDataSet) getKey(i int) Int3Key {
	return im.keys[i]
}

func (im *IntMapTestDataSet) getNotKey(i int) Int3Key {
	return Int3Key{
		im.keys[i][0] + 1,
		im.keys[i][1],
		im.keys[i][2] - 1,
	}
}
