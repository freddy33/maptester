package main

import (
	"fmt"
	"github.com/freddy33/maptester"
	"os"
	"runtime"
)

func main() {
	c := ""
	if len(os.Args) > 1 {
		c = os.Args[1]
	}
	switch c {
	case "help":
		usage()
	case "show":
		maptester.DisplayConfigurations()
	case "clean":
		maptester.DeleteAllData()
	case "regen":
		maptester.DeleteAllData()
		maptester.GenAllData()
	case "gen":
		maptester.GenAllData()
	case "analyze":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		filenames := make([]string, len(os.Args)-2)
		for idx, _ := range filenames {
			filenames[idx] = os.Args[idx+2]
		}
		maptester.AnalyzePerfFiles(filenames)
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
		runtime.GOMAXPROCS(maptester.MaxConThreads * 2)
		if !maptester.TestAll() {
			os.Exit(4)
		}
	default:
		fmt.Printf("Command %q unknown\n", c)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf("Usage: $ maptester [command] (name) (options)\n" +
		"\tcommand: help, show, clean, gen, regen, read [name], test, analyze [list of file names]\n")
}
