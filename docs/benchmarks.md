# Reproducible scanner benchmarks

Benchmarks are evidence for regressions and engineering decisions, not a
promise about a user's filesystem. Filesystem cache state, storage, directory
shape, antivirus software, and the number of workers can materially change the
result.

## Fixture

`BenchmarkAnalyzeWithOptions` in `internal/scanner/bench_test.go` creates 40
synthetic repositories. Each contains `.git` and a `node_modules/pkg` directory
with 25 files of 4,096 bytes: 1,000 generated files and 4,096,000 payload bytes
in total. The benchmark uses the default eight-worker scheduler and a minimum
candidate size of one byte.

The other benchmarks use deterministic in-memory fixtures:

- `BenchmarkPushTop`: 10,000 entries with a retained top-20;
- `BenchmarkSummarizeGroups`: 5,000 candidates across 250 groups with a
  retained top-20.

## Baseline from 2026-08-14

Environment:

- commit: `08d3fb687a83ab02a9e0fab55777bcea3e7c2e22`;
- Go: `go1.25.12 linux/amd64`;
- CPU exposed to the runner: AMD EPYC 9V74, 9 logical CPUs visible and
  `GOMAXPROCS=8` for the benchmark;
- kernel: Linux 6.18.35 x86_64 under KVM;
- fixture filesystem reported by `stat -f`: `ext2/ext3`;
- five benchmark samples, with the median shown below.

| Benchmark | Median time/op | Bytes/op | Allocations/op |
| --- | ---: | ---: | ---: |
| Analyze synthetic filesystem | 3,915,957 ns | 916,358 B | 7,652 |
| Maintain bounded top-20 | 51,245 ns | 1,640 B | 6 |
| Summarize 5,000 candidates | 180,518 ns | 59,584 B | 25 |

Exact command:

```sh
go test -run '^$' \
  -bench 'Benchmark(AnalyzeWithOptions|PushTop|SummarizeGroups)$' \
  -benchmem -count=5 ./internal/scanner
```

Keep all raw samples when publishing a comparison. Do not compare these values
with another machine, fixture, Go version, or cache state as if they were the
same experiment.
