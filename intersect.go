package main

import (
	"os"
)

func cmdIntersect(args []string) error {
	var aPath, bPath string
	modeU, modeV, modeWA := false, false, false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-a":
			i++
			if i >= len(args) {
				return usageError{"intersect: -a requires a value"}
			}
			aPath = args[i]
		case "-b":
			i++
			if i >= len(args) {
				return usageError{"intersect: -b requires a value"}
			}
			bPath = args[i]
		case "-u":
			modeU = true
		case "-v":
			modeV = true
		case "-wa":
			modeWA = true
		default:
			return usageError{"intersect: unknown argument " + args[i]}
		}
	}
	if aPath == "" || bPath == "" {
		return usageError{"intersect: -a and -b are both required"}
	}
	if boolCount(modeU, modeV, modeWA) > 1 {
		return usageError{"intersect: -u, -v and -wa are mutually exclusive"}
	}
	if bPath == "-" {
		return usageError{"intersect: -b must be a file, not stdin"}
	}

	aIvs, err := readBed(aPath)
	if err != nil {
		return err
	}
	bIvs, err := readBed(bPath)
	if err != nil {
		return err
	}

	byChrom := map[string][]Interval{}
	for _, b := range bIvs {
		byChrom[b.Chrom] = append(byChrom[b.Chrom], b)
	}

	var out []Interval
	for _, a := range aIvs {
		var overlaps []Interval
		for _, b := range byChrom[a.Chrom] {
			region, ok := intersectRegion(a, b)
			if ok {
				overlaps = append(overlaps, region)
			}
		}

		switch {
		case modeV:
			if len(overlaps) == 0 {
				out = append(out, a)
			}
		case modeU:
			if len(overlaps) > 0 {
				out = append(out, a)
			}
		case modeWA:
			for range overlaps {
				out = append(out, a)
			}
		default:
			for _, region := range overlaps {
				out = append(out, Interval{Chrom: a.Chrom, Start: region.Start, End: region.End, Rest: a.Rest})
			}
		}
	}

	return writeLines(os.Stdout, out)
}

// intersectRegion reports whether a and b overlap and, if so, the overlapping
// span. A zero-length record (start == end) is treated by real bedtools as
// reaching one base further than its own coordinate when it is the side
// being queried against (b here) - encoded as effB below - and a zero-length
// a is allowed to report a zero-width match against a genuinely touching b.
// Both quirks were derived empirically (see tests/README.md); plain BED
// semantics alone do not predict them.
func intersectRegion(a, b Interval) (Interval, bool) {
	effBStart, effBEnd := b.Start, b.End
	if b.IsZeroLength() {
		effBStart--
		effBEnd++
	}
	start := max(a.Start, effBStart)
	end := min(a.End, effBEnd)
	eligible := start < end || (a.IsZeroLength() && start <= end)
	if !eligible {
		return Interval{}, false
	}
	return Interval{Chrom: a.Chrom, Start: start, End: end}, true
}

func boolCount(bs ...bool) int {
	n := 0
	for _, b := range bs {
		if b {
			n++
		}
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
