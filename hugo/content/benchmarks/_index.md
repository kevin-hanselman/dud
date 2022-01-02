# Benchmarks

**OS**: Linux 5.15

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Software versions**:

Dud version: v0.3.0

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
| `Dud` | 0.014 ± 0.004 | 0.009 | 0.017 | 1.00 |
| `DVC` | 0.359 ± 0.006 | 0.354 | 0.365 | 25.18 ± 7.70 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.760 ± 0.072 | 0.707 | 0.842 | 1.00 |
| `DVC` | 3.759 ± 0.032 | 3.724 | 3.784 | 4.95 ± 0.47 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.010 ± 0.000 | 0.010 | 0.010 | 1.00 |
| `DVC` | 0.324 ± 0.004 | 0.321 | 0.328 | 32.22 ± 0.57 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.249 ± 0.271 | 1.088 | 1.562 | 1.00 |
| `DVC` | 5.042 ± 0.015 | 5.030 | 5.060 | 4.04 ± 0.88 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.928 ± 0.078 | 1.840 | 1.986 | 1.00 |
| `DVC` | 65.670 ± 0.745 | 65.066 | 66.503 | 34.06 ± 1.43 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.315 ± 0.009 | 0.307 | 0.324 | 1.00 |
| `DVC` | 1.601 ± 0.020 | 1.583 | 1.622 | 5.09 ± 0.15 |
