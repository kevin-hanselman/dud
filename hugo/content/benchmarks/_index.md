# Benchmarks

**OS**: Linux 5.15

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Software versions**:

Dud version: v0.2.0-10-gbda3fc3

DVC version: 2.6.4 (pip)

DVC non-default configuration:

    core.analytics=false
    core.check_update=false
    cache.type=symlink
## Few large files

This dataset consists of four 1 GB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.021 ± 0.016 | 0.009 | 0.039 | 1.00 |
| `DVC` | 0.383 ± 0.028 | 0.357 | 0.413 | 18.65 ± 15.06 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.919 ± 0.255 | 0.723 | 1.208 | 1.00 |
| `DVC` | 3.804 ± 0.082 | 3.753 | 3.899 | 4.14 ± 1.16 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.018 ± 0.011 | 0.010 | 0.030 | 1.00 |
| `DVC` | 0.334 ± 0.009 | 0.326 | 0.343 | 18.76 ± 11.21 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.530 ± 0.119 | 1.428 | 1.661 | 1.00 |
| `DVC` | 4.961 ± 0.024 | 4.935 | 4.980 | 3.24 ± 0.25 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.785 ± 0.102 | 1.669 | 1.858 | 1.00 |
| `DVC` | 68.034 ± 2.807 | 65.276 | 70.887 | 38.10 ± 2.69 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.330 ± 0.029 | 0.308 | 0.363 | 1.00 |
| `DVC` | 1.626 ± 0.037 | 1.586 | 1.659 | 4.93 ± 0.45 |
