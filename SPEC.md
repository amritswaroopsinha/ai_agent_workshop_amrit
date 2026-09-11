# SPEC.md — mytools

## 1. Scope

v1 ships four subcommands:

| Subcommand  | Flags in v1              |
|-------------|--------------------------|
| `sort`      | (none)                   |
| `merge`     | `-d <int>`               |
| `intersect` | `-u`, `-v`, `-wa`        |
| `subtract`  | (none)                   |

Not in v1: `closest`, `-s`/`-S` strand awareness, `-f` minimum overlap, BED12,
GFF3/VCF, compressed input, `-header`. `closest` is a stretch goal, added later.

## 2. Invocation

    mytools sort      -i <file|->
    mytools merge     -i <file|-> [-d N]
    mytools intersect -a <file|-> -b <file>
    mytools subtract  -a <file|-> -b <file>
    mytools --version

Flag names and meanings match bedtools exactly. `-` means stdin; exactly one input
may be `-`. `-b` must be a real file (we read it fully; see §6).

## 3. Input format

BED3 through BED6, tab-separated:

    chrom  start  end  [name  score  strand]

- Column count may vary between lines; treat missing trailing columns as absent.
- Lines beginning with `#`, `track`, or `browser` are skipped silently.
- Blank lines are skipped.
- `start` and `end` are non-negative integers. `start > end` is an error (§7).
- `start == end` (zero-length) is **legal**. See §4.

## 4. Interval semantics

BED is **0-based, half-open**. `chr1 100 200` covers bases 100..199.

- Overlap: `a.start < b.end AND b.start < a.end`. Strict `<` on both sides.
- Bookended intervals (`a.end == b.start`) do **not** overlap.
- Bookended intervals **do** merge at `-d 0` (bedtools' default), because `merge`
  joins features whose gap is `<= d`, and the gap here is 0.
- Zero-length intervals are legal. Match whatever real `bedtools` prints on
  `data/a.bed` rather than reasoning about them from first principles.

## 5. Output

- Tab-separated, LF line endings, trailing newline on the final line.
- Empty result: print nothing, exit 0.
- `sort`: all input columns preserved. Order is by chrom (lexicographic — `chr17`
  before `chr7`), then `start`, then `end`.
- `merge`: BED3 only (`chrom start end`). Input columns are dropped.
- `intersect`: default prints the intersected region carrying `-a`'s trailing
  columns. `-wa` prints `-a`'s original interval, once per overlapping `-b` feature.
  `-u` prints each `-a` feature at most once. `-v` prints `-a` features with no
  overlap. `-u`, `-v` and `-wa` are mutually exclusive.
- `subtract`: `-a` features with `-b` regions removed. One feature may become two,
  or vanish entirely.
- Input order is preserved for `intersect` and `subtract` (bedtools does not sort
  `-a` for you, and neither do we).

## 6. Memory model

- `sort` holds the whole input in memory. Acceptable at our sizes.
- `merge` streams, assuming sorted input. It does **not** sort for you; unsorted
  input is an error (§7), same as bedtools.
- `intersect` and `subtract` load `-b` into memory (grouped by chrom, sorted by
  start, binary-searched) and stream `-a`.
- Target: inputs up to ~10^6 intervals. No mmap, no index files, no threads.

## 7. Errors and exit codes

Errors go to **stderr**. stdout carries data only.

| Situation                             | exit |
|---------------------------------------|------|
| Success (including empty output)       | 0   |
| Malformed line / non-integer coords    | 1   |
| `start > end`                          | 1   |
| Unsorted input to `merge`              | 1   |
| Unknown flag / missing required arg    | 2   |
| Input file does not exist              | 2   |
| `--version`, `--help`                  | 0   |

Data problems are 1. Caller problems are 2. Messages name the file and line number:
`a.bed:14: start > end (500 > 400)`.

## 8. Correctness

Real `bedtools` is the oracle. Every one of these must produce byte-identical output
to its bedtools equivalent on the files in `data/`:

    mytools sort -i data/a.bed                    == bedtools sort -i data/a.bed
    mytools merge -i <sorted a.bed>               == bedtools merge -i <sorted a.bed>
    mytools merge -d 10 -i <sorted a.bed>         == bedtools merge -d 10 -i <sorted a.bed>
    mytools intersect -a data/a.bed -b data/b.bed == bedtools intersect -a ... -b ...
    mytools intersect -u  ...                     == bedtools intersect -u ...
    mytools intersect -v  ...                     == bedtools intersect -v ...
    mytools intersect -wa ...                     == bedtools intersect -wa ...
    mytools subtract -a data/a.bed -b data/b.bed  == bedtools subtract -a ... -b ...

Also required: `mytools --version` prints a version and exits 0.

Accepted deviations from bedtools: none. If one is found that cannot be fixed, it
gets written down here with the reason.

## 9. Language and layout

- Implementation language: Go. One codebase, one language — see `CLAUDE.md`.
- `mytools` is a single Go module built to a `mytools` binary, on `PATH` for tests.
- Tests live in `tests/`, driven by `tests/run_golden.sh`.
- No third-party runtime dependencies (standard library only).
