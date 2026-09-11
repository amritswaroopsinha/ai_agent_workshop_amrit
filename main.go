package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

const usage = `usage: mytools <command> [args]

commands:
  sort      -i <file|->
  merge     -i <file|-> [-d N]
  intersect -a <file|-> -b <file> [-u|-v|-wa]
  subtract  -a <file|-> -b <file>
  --version
`

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println("mytools " + version)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "sort":
		err = cmdSort(os.Args[2:])
	case "merge":
		err = cmdMerge(os.Args[2:])
	case "intersect":
		err = cmdIntersect(os.Args[2:])
	case "subtract":
		err = cmdSubtract(os.Args[2:])
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "mytools: "+err.Error())
		switch err.(type) {
		case usageError:
			os.Exit(2)
		default:
			os.Exit(1)
		}
	}
}
