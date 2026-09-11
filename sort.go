package main

import (
	"os"
	"sort"
)

func cmdSort(args []string) error {
	inputPath := "-"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-i":
			i++
			if i >= len(args) {
				return usageError{"sort: -i requires a value"}
			}
			inputPath = args[i]
		default:
			return usageError{"sort: unknown argument " + args[i]}
		}
	}

	ivs, err := readBed(inputPath)
	if err != nil {
		return err
	}

	sort.SliceStable(ivs, func(i, j int) bool {
		if ivs[i].Chrom != ivs[j].Chrom {
			return ivs[i].Chrom < ivs[j].Chrom
		}
		if ivs[i].Start != ivs[j].Start {
			return ivs[i].Start < ivs[j].Start
		}
		return ivs[i].End < ivs[j].End
	})

	return writeLines(os.Stdout, ivs)
}
