# Benchmarks

**OS**: Linux 5.13

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Software versions**:

Dud version: v0.2.0

DVC version: 2.6.4 (pip)
## Few large files

This dataset consists of four 1 GB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.017 ± 0.008 | 0.012 | 0.025 | 1.00 |
| `DVC` | 0.396 ± 0.012 | 0.385 | 0.408 | 23.98 ± 11.18 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.873 ± 0.030 | 0.838 | 0.894 | 1.00 |
| `DVC` | 3.691 ± 0.030 | 3.666 | 3.725 | 4.23 ± 0.15 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.009 ± 0.002 | 0.007 | 0.011 | 1.00 |
| `DVC` | 0.342 ± 0.015 | 0.325 | 0.352 | 39.15 ± 9.57 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.642 ± 0.097 | 1.543 | 1.736 | 1.00 |
| `DVC` | 4.752 ± 0.050 | 4.699 | 4.798 | 2.89 ± 0.17 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.511 ± 0.034 | 1.472 | 1.534 | 1.00 |
| `DVC` | 67.207 ± 1.881 | 65.036 | 68.358 | 44.47 ± 1.60 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.436 ± 0.035 | 0.405 | 0.474 | 1.00 |
| `DVC` | 1.640 ± 0.032 | 1.608 | 1.671 | 3.76 ± 0.31 |
