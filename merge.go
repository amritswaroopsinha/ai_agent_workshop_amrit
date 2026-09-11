package main

import (
	"os"
)

func cmdMerge(args []string) error {
	inputPath := "-"
	d := 0
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-i":
			i++
			if i >= len(args) {
				return usageError{"merge: -i requires a value"}
			}
			inputPath = args[i]
		case "-d":
			i++
			if i >= len(args) {
				return usageError{"merge: -d requires a value"}
			}
			n, err := parseInt(args[i])
			if err != nil {
				return usageError{"merge: -d requires an integer"}
			}
			d = n
		default:
			return usageError{"merge: unknown argument " + args[i]}
		}
	}

	ivs, err := readBed(inputPath)
	if err != nil {
		return err
	}

	if err := checkSorted(ivs); err != nil {
		return err
	}

	out := mergeIntervals(ivs, d)
	return writeLines(os.Stdout, out)
}

func checkSorted(ivs []Interval) error {
	seenChrom := map[string]bool{}
	lastChrom := ""
	lastStart := -1
	for i, iv := range ivs {
		if i == 0 || iv.Chrom != lastChrom {
			if seenChrom[iv.Chrom] {
				return dataError{"merge: input not sorted (chrom " + iv.Chrom + " out of order)"}
			}
			if i > 0 {
				seenChrom[lastChrom] = true
			}
			lastChrom = iv.Chrom
			lastStart = iv.Start
			continue
		}
		if iv.Start < lastStart {
			return dataError{"merge: input not sorted (start out of order on " + iv.Chrom + ")"}
		}
		lastStart = iv.Start
	}
	return nil
}

// mergeIntervals implements bedtools' merge -d semantics, including its
// zero-length-feature quirk: a zero-length record extends its reach by one
// base past its own coordinate, both for deciding whether the next record
// joins the group (eligibility) and, if it ends up defining the group's min
// start or max end, for the printed boundary itself. Verified empirically
// against real bedtools rather than derived from the BED spec, per
// tests/README.md's guidance on zero-length features.
func mergeIntervals(ivs []Interval, d int) []Interval {
	var out []Interval
	open := false
	var groupChrom string
	var groupStart, groupEnd int
	var groupStartIsZero, groupEndIsZero bool
	groupCount := 0

	flush := func() {
		if !open {
			return
		}
		s, e := groupStart, groupEnd
		if groupCount > 1 {
			if groupStartIsZero {
				s--
			}
			if groupEndIsZero {
				e++
			}
		}
		out = append(out, Interval{Chrom: groupChrom, Start: s, End: e})
		open = false
	}

	startGroup := func(iv Interval) {
		groupChrom = iv.Chrom
		groupStart = iv.Start
		groupEnd = iv.End
		groupStartIsZero = iv.IsZeroLength()
		groupEndIsZero = iv.IsZeroLength()
		groupCount = 1
		open = true
	}

	for _, iv := range ivs {
		if !open {
			startGroup(iv)
			continue
		}
		if iv.Chrom != groupChrom {
			flush()
			startGroup(iv)
			continue
		}
		tolerance := 0
		if iv.IsZeroLength() {
			tolerance++
		}
		if groupEndIsZero {
			tolerance++
		}
		if iv.Start-groupEnd <= d+tolerance {
			groupCount++
			if iv.End > groupEnd {
				groupEnd = iv.End
				groupEndIsZero = iv.IsZeroLength()
			} else if iv.End == groupEnd && iv.IsZeroLength() {
				groupEndIsZero = true
			}
			continue
		}
		flush()
		startGroup(iv)
	}
	flush()

	return out
}
