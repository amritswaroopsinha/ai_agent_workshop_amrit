package main

import "testing"

// intersectRegion is the one comparison every off-by-one bug in this project
// lives in (see CLAUDE.md). These cases run without bedtools and are the
// direct check on that predicate that the golden tests only exercise
// indirectly through data/a.bed and data/b.bed.
func TestIntersectRegion(t *testing.T) {
	iv := func(start, end int) Interval { return Interval{Chrom: "chr1", Start: start, End: end} }

	cases := []struct {
		name      string
		a, b      Interval
		wantOK    bool
		wantStart int
		wantEnd   int
	}{
		{
			name:   "bookended: a ends where b starts, no overlap",
			a:      iv(150, 250),
			b:      iv(250, 300),
			wantOK: false,
		},
		{
			name:   "bookended: b ends where a starts, no overlap",
			a:      iv(250, 300),
			b:      iv(150, 250),
			wantOK: false,
		},
		{
			name:      "nested: b fully inside a",
			a:         iv(100, 400),
			b:         iv(180, 220),
			wantOK:    true,
			wantStart: 180,
			wantEnd:   220,
		},
		{
			name:      "identical intervals",
			a:         iv(300, 400),
			b:         iv(300, 400),
			wantOK:    true,
			wantStart: 300,
			wantEnd:   400,
		},
		{
			name:   "disjoint, far apart",
			a:      iv(0, 100),
			b:      iv(900, 1000),
			wantOK: false,
		},
		{
			name:      "position 0: touching at the origin overlaps",
			a:         iv(0, 10),
			b:         iv(5, 15),
			wantOK:    true,
			wantStart: 5,
			wantEnd:   10,
		},
		{
			name:   "position 0: bookended at the origin, non-zero-length, no overlap",
			a:      iv(0, 5),
			b:      iv(5, 10),
			wantOK: false,
		},
		{
			// a07/b02 in data/a.bed, data/b.bed: a zero-length a is allowed a
			// zero-width match against a genuinely touching, non-zero b.
			name:      "zero-length a touching non-zero b",
			a:         iv(100, 100),
			b:         iv(0, 100),
			wantOK:    true,
			wantStart: 100,
			wantEnd:   100,
		},
		{
			// Same shape as above, but at position 0 (a12 in data/a.bed).
			name:      "position 0: zero-length a touching non-zero b at the origin",
			a:         iv(0, 0),
			b:         iv(0, 100),
			wantOK:    true,
			wantStart: 0,
			wantEnd:   0,
		},
		{
			// A zero-length b reaches one base past its own coordinate on
			// both sides (empirically derived, see intersect.go's comment).
			name:      "zero-length b reaches one base either side",
			a:         iv(90, 110),
			b:         iv(100, 100),
			wantOK:    true,
			wantStart: 99,
			wantEnd:   101,
		},
		{
			name:   "zero-length a strictly outside non-zero b",
			a:      iv(50, 50),
			b:      iv(100, 200),
			wantOK: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := intersectRegion(c.a, c.b)
			if ok != c.wantOK {
				t.Fatalf("intersectRegion(%+v, %+v) ok = %v, want %v", c.a, c.b, ok, c.wantOK)
			}
			if !ok {
				return
			}
			if got.Start != c.wantStart || got.End != c.wantEnd {
				t.Fatalf("intersectRegion(%+v, %+v) = [%d,%d), want [%d,%d)",
					c.a, c.b, got.Start, got.End, c.wantStart, c.wantEnd)
			}
		})
	}
}

func TestIsZeroLength(t *testing.T) {
	if !(Interval{Start: 5, End: 5}).IsZeroLength() {
		t.Fatal("start == end should be zero-length")
	}
	if (Interval{Start: 5, End: 6}).IsZeroLength() {
		t.Fatal("start != end should not be zero-length")
	}
}
