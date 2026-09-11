package main

import (
	"os"
	"sort"
)

func cmdSubtract(args []string) error {
	var aPath, bPath string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-a":
			i++
			if i >= len(args) {
				return usageError{"subtract: -a requires a value"}
			}
			aPath = args[i]
		case "-b":
			i++
			if i >= len(args) {
				return usageError{"subtract: -b requires a value"}
			}
			bPath = args[i]
		default:
			return usageError{"subtract: unknown argument " + args[i]}
		}
	}
	if aPath == "" || bPath == "" {
		return usageError{"subtract: -a and -b are both required"}
	}
	if bPath == "-" {
		return usageError{"subtract: -b must be a file, not stdin"}
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
	for chrom := range byChrom {
		bs := byChrom[chrom]
		sort.Slice(bs, func(i, j int) bool { return bs[i].Start < bs[j].Start })
		byChrom[chrom] = bs
	}

	var out []Interval
	for _, a := range aIvs {
		out = append(out, subtractOne(a, byChrom[a.Chrom])...)
	}

	return writeLines(os.Stdout, out)
}

// subtractOne removes, from a, every span covered by any overlapping b. A
// zero-length b (start == end) is treated as if it actually spans one base
// on each side of its point - effStart-1..effEnd+1 - matching real bedtools,
// which erodes a's boundary by one base wherever such a feature touches it.
// This does not come from the BED spec; it was derived by diffing against
// real bedtools on cases like a01/b02 in data/, per tests/README.md.
func subtractOne(a Interval, bs []Interval) []Interval {
	// A zero-length a has no BED-legal way to be "partly" covered - it is
	// removed entirely or not at all. It disappears when strictly inside a
	// normal b, or when it sits at the exact same point as a zero-length b;
	// both were confirmed empirically (a07 in data/a.bed against data/b.bed).
	if a.IsZeroLength() {
		p := a.Start
		for _, b := range bs {
			if b.IsZeroLength() {
				if p == b.Start {
					return nil
				}
				continue
			}
			if b.Start < p && p < b.End {
				return nil
			}
		}
		return []Interval{a}
	}

	type span struct{ start, end int }
	var removals []span
	for _, b := range bs {
		effStart, effEnd := b.Start, b.End
		if b.IsZeroLength() {
			effStart--
			effEnd++
		}
		start := max(a.Start, effStart)
		end := min(a.End, effEnd)
		if start < end {
			removals = append(removals, span{start, end})
		}
	}
	if len(removals) == 0 {
		return []Interval{a}
	}

	sort.Slice(removals, func(i, j int) bool { return removals[i].start < removals[j].start })
	merged := removals[:0:0]
	for _, r := range removals {
		if n := len(merged); n > 0 && r.start <= merged[n-1].end {
			if r.end > merged[n-1].end {
				merged[n-1].end = r.end
			}
			continue
		}
		merged = append(merged, r)
	}

	var out []Interval
	cursor := a.Start
	for _, r := range merged {
		if cursor < r.start {
			out = append(out, Interval{Chrom: a.Chrom, Start: cursor, End: r.start, Rest: a.Rest})
		}
		if r.end > cursor {
			cursor = r.end
		}
	}
	if cursor < a.End {
		out = append(out, Interval{Chrom: a.Chrom, Start: cursor, End: a.End, Rest: a.Rest})
	}
	return out
}
