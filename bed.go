package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Interval is one BED3-BED6 record. Rest holds any columns after end, verbatim,
// for pass-through in commands that preserve them (sort, intersect default/-wa).
type Interval struct {
	Chrom string
	Start int
	End   int
	Rest  []string
}

func (iv Interval) IsZeroLength() bool { return iv.Start == iv.End }

func (iv Interval) Fields() []string {
	f := make([]string, 0, 3+len(iv.Rest))
	f = append(f, iv.Chrom, strconv.Itoa(iv.Start), strconv.Itoa(iv.End))
	return append(f, iv.Rest...)
}

func (iv Interval) String() string {
	return strings.Join(iv.Fields(), "\t")
}

// usageError formats a bad-flag/missing-file condition; callers exit 2 on it.
type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }

// dataError formats a malformed-input condition; callers exit 1 on it.
type dataError struct{ msg string }

func (e dataError) Error() string { return e.msg }

func openInput(path string) (io.ReadCloser, string, error) {
	if path == "-" {
		return io.NopCloser(os.Stdin), "-", nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "", usageError{fmt.Sprintf("cannot open %s: %v", path, err)}
	}
	return f, path, nil
}

// readBed reads and parses every record from path ("-" for stdin), skipping
// blank lines and track/browser/# comment lines.
func readBed(path string) ([]Interval, error) {
	f, name, err := openInput(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []Interval
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 64*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "track") || strings.HasPrefix(line, "browser") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			return nil, dataError{fmt.Sprintf("%s:%d: expected at least 3 tab-separated fields, got %d", name, lineNo, len(fields))}
		}
		start, err := strconv.Atoi(fields[1])
		if err != nil || start < 0 {
			return nil, dataError{fmt.Sprintf("%s:%d: invalid start %q", name, lineNo, fields[1])}
		}
		end, err := strconv.Atoi(fields[2])
		if err != nil || end < 0 {
			return nil, dataError{fmt.Sprintf("%s:%d: invalid end %q", name, lineNo, fields[2])}
		}
		if start > end {
			return nil, dataError{fmt.Sprintf("%s:%d: start > end (%d > %d)", name, lineNo, start, end)}
		}
		rest := append([]string{}, fields[3:]...)
		out = append(out, Interval{Chrom: fields[0], Start: start, End: end, Rest: rest})
	}
	if err := sc.Err(); err != nil {
		return nil, dataError{fmt.Sprintf("%s: %v", name, err)}
	}
	return out, nil
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func writeLines(w io.Writer, ivs []Interval) error {
	bw := bufio.NewWriter(w)
	for _, iv := range ivs {
		if _, err := bw.WriteString(iv.String()); err != nil {
			return err
		}
		if _, err := bw.WriteString("\n"); err != nil {
			return err
		}
	}
	return bw.Flush()
}
