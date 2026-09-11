package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println("mytools " + version)
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "usage: mytools --version")
	os.Exit(2)
}
