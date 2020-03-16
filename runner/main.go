package main

import (
	"fmt"
	"github.com/freddy33/maptester"
	"os"
)

func main() {
	c := ""
	if len(os.Args) > 1 {
		c = os.Args[1]
	}
	switch c {
	case "help":
		usage()
	case "clean":
		maptester.DeleteAllData()
	case "regen":
		maptester.DeleteAllData()
		maptester.GenAllData()
	case "gen":
		maptester.GenAllData()
	case "read":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		name := os.Args[2]
		im, res := maptester.ReadIntData(name, maptester.GenDataSize)
		goodData := maptester.Verify(name, im, res)
		if !goodData {
			os.Exit(3)
		}
	case "test":
		if !maptester.TestAll() {
			os.Exit(4)
		}
	default:
		fmt.Printf("Command '%s' unknown", c)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf("Usage: $ maptester [command] (name) (options)\n" +
		"\tcommand: help, gen, read [name], test [name]")
}
