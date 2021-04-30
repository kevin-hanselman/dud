# Benchmarks

**OS**: Linux 5.11

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Software versions**:

dud version v0.0.1-26-g301044c

DVC version: 2.0.6 (pip)
## Few large files

This dataset consists of four 1 GB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.023 ± 0.015 | 0.007 | 0.038 | 1.00 |
| `DVC` | 0.297 ± 0.003 | 0.295 | 0.300 | 12.68 ± 8.35 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.798 ± 0.030 | 0.774 | 0.832 | 1.00 |
| `DVC` | 3.728 ± 0.058 | 3.668 | 3.784 | 4.67 ± 0.19 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.020 ± 0.020 | 0.007 | 0.043 | 1.00 |
| `DVC` | 0.288 ± 0.013 | 0.275 | 0.301 | 14.18 ± 13.89 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.021 ± 0.331 | 0.818 | 1.403 | 1.00 |
| `DVC` | 8.438 ± 0.322 | 8.244 | 8.810 | 8.27 ± 2.70 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.681 ± 0.047 | 1.649 | 1.735 | 1.00 |
| `DVC` | 36.496 ± 1.676 | 34.965 | 38.287 | 21.71 ± 1.16 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.317 ± 0.036 | 0.294 | 0.359 | 1.00 |
| `DVC` | 1.312 ± 0.020 | 1.297 | 1.336 | 4.14 ± 0.48 |
