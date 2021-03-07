# Benchmarks

**OS**: Linux 5.10.16-arch1-1 x86_64 GNU/Linux

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Dud version**: dud version ad31ccb

**DVC version**: DVC version: 2.0.3 (pip)
## Few large files

This dataset consists of four 1 GB files in a single directory.

### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.772 ± 0.060 | 0.735 | 0.841 | 1.00 |
| `DVC` | 3.831 ± 0.055 | 3.783 | 3.891 | 4.96 ± 0.39 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.644 ± 0.098 | 1.556 | 1.749 | 1.00 |
| `DVC` | 31.733 ± 2.838 | 29.915 | 35.003 | 19.30 ± 2.07 |
