# Benchmarks

**OS**: Linux 5.11

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Software versions**:

dud version v0.0.1-21-g0bee52f

DVC version: 2.0.6 (pip)
## Few large files

This dataset consists of four 1 GB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.008 ± 0.000 | 0.008 | 0.008 | 1.00 |
| `DVC` | 0.280 ± 0.020 | 0.260 | 0.301 | 36.31 ± 2.65 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.722 ± 0.021 | 0.701 | 0.742 | 1.00 |
| `DVC` | 3.913 ± 0.279 | 3.720 | 4.232 | 5.42 ± 0.42 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.026 ± 0.014 | 0.011 | 0.040 | 1.00 |
| `DVC` | 0.292 ± 0.030 | 0.273 | 0.326 | 11.42 ± 6.53 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.024 ± 0.170 | 0.860 | 1.200 | 1.00 |
| `DVC` | 8.540 ± 0.240 | 8.308 | 8.787 | 8.34 ± 1.41 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.737 ± 0.163 | 1.572 | 1.898 | 1.00 |
| `DVC` | 39.000 ± 6.423 | 32.868 | 45.678 | 22.45 ± 4.26 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.330 ± 0.005 | 0.326 | 0.335 | 1.00 |
| `DVC` | 1.413 ± 0.017 | 1.394 | 1.425 | 4.28 ± 0.08 |
