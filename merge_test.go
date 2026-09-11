package main

import "testing"

// mergeIntervals is where the zero-length-feature quirk lives (see the
// comment on the function). These cases pin the additive tolerance rule —
// each zero-length side of a candidate join adds 1, not "any zero-length
// side adds 1" — verified empirically against real bedtools (fuzzed against
// it directly; a capped, non-additive tolerance passed data/a.bed's golden
// cases but broke on inputs with two zero-length features in a row).
func TestMergeIntervals(t *testing.T) {
	iv := func(chrom string, start, end int) Interval { return Interval{Chrom: chrom, Start: start, End: end} }

	cases := []struct {
		name string
		in   []Interval
		d    int
		want []Interval
	}{
		{
			// Per SPEC.md: bookended intervals do NOT overlap, but they DO
			// merge at d=0 (gap 0 <= d).
			name: "bookended, d=0: merges",
			in:   []Interval{iv("chr1", 100, 200), iv("chr1", 200, 300)},
			d:    0,
			want: []Interval{iv("chr1", 100, 300)},
		},
		{
			name: "bookended, d=-1: requires genuine overlap, does not merge",
			in:   []Interval{iv("chr1", 100, 200), iv("chr1", 200, 300)},
			d:    -1,
			want: []Interval{iv("chr1", 100, 200), iv("chr1", 200, 300)},
		},
		{
			name: "lone zero-length feature is left untouched",
			in:   []Interval{iv("chr1", 500, 500)},
			d:    0,
			want: []Interval{iv("chr1", 500, 500)},
		},
		{
			name: "two zero-length features, gap 4: needs d=2 (1+1 tolerance), not d=1",
			in:   []Interval{iv("chr1", 10, 10), iv("chr1", 14, 14)},
			d:    1,
			want: []Interval{iv("chr1", 10, 10), iv("chr1", 14, 14)},
		},
		{
			name: "two zero-length features, gap 4, d=2: merges and both edges expand",
			in:   []Interval{iv("chr1", 10, 10), iv("chr1", 14, 14)},
			d:    2,
			want: []Interval{iv("chr1", 9, 15)},
		},
		{
			name: "zero-length then normal, gap 3, d=2: only one side's tolerance applies",
			in:   []Interval{iv("chr1", 10, 10), iv("chr1", 13, 20)},
			d:    2,
			want: []Interval{iv("chr1", 9, 20)},
		},
		{
			name: "zero-length then normal, gap 3, d=1: not enough tolerance",
			in:   []Interval{iv("chr1", 10, 10), iv("chr1", 13, 20)},
			d:    1,
			want: []Interval{iv("chr1", 10, 10), iv("chr1", 13, 20)},
		},
		{
			name: "chain of two zero-length features bridging into a normal one, d=2",
			in:   []Interval{iv("chr1", 14, 14), iv("chr1", 18, 18), iv("chr1", 29, 38)},
			d:    2,
			want: []Interval{iv("chr1", 13, 19), iv("chr1", 29, 38)},
		},
		{
			name: "nested: b fully inside a",
			in:   []Interval{iv("chr1", 300, 400), iv("chr1", 320, 350)},
			d:    0,
			want: []Interval{iv("chr1", 300, 400)},
		},
		{
			name: "position 0: interval touching the origin",
			in:   []Interval{iv("chr1", 0, 100), iv("chr1", 100, 200)},
			d:    0,
			want: []Interval{iv("chr1", 0, 200)},
		},
		{
			name: "chrom boundary: never merges across chroms regardless of d",
			in:   []Interval{iv("chr1", 0, 100), iv("chr2", 100, 200)},
			d:    1000,
			want: []Interval{iv("chr1", 0, 100), iv("chr2", 100, 200)},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := mergeIntervals(c.in, c.d)
			if len(got) != len(c.want) {
				t.Fatalf("mergeIntervals(%v, %d) = %v, want %v", c.in, c.d, got, c.want)
			}
			for i := range got {
				if got[i].Chrom != c.want[i].Chrom || got[i].Start != c.want[i].Start || got[i].End != c.want[i].End {
					t.Fatalf("mergeIntervals(%v, %d) = %v, want %v", c.in, c.d, got, c.want)
				}
			}
		})
	}
}
