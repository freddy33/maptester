package main

import (
	"fmt"
	"github.com/freddy33/maptester"
	"os"
)

func main() {
	c := "all"
	if len(os.Args) > 1 {
		c = os.Args[1]
	}
	switch c {
	case "all":
		maptester.GenAllData()
	default:
		fmt.Printf("Command '%s' unknown", c)
		os.Exit(1)
	}
}
