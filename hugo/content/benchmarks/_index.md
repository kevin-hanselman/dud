# Benchmarks

**OS**: Linux 5.10.12-arch1-1 x86_64 GNU/Linux

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB
## Few large files

This dataset consists of four 1 GB files in a single directory.

### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.774 ± 0.006 | 0.767 | 0.777 | 1.00 |
| `DVC` | 3.933 ± 0.057 | 3.877 | 3.990 | 5.08 ± 0.08 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.603 ± 0.070 | 1.540 | 1.679 | 1.00 |
| `DVC` | 16.046 ± 0.025 | 16.030 | 16.074 | 10.01 ± 0.44 |
