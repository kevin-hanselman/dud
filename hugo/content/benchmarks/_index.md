# Benchmarks

**OS**: Linux 5.12

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Software versions**:

Dud version: v0.1.0

DVC version: 2.3.0 (pip)
## Few large files

This dataset consists of four 1 GB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.018 ± 0.017 | 0.007 | 0.038 | 1.00 |
| `DVC` | 0.387 ± 0.013 | 0.377 | 0.401 | 21.50 ± 20.29 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.705 ± 0.028 | 0.673 | 0.723 | 1.00 |
| `DVC` | 3.700 ± 0.050 | 3.642 | 3.736 | 5.25 ± 0.22 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.020 ± 0.013 | 0.011 | 0.036 | 1.00 |
| `DVC` | 0.338 ± 0.009 | 0.328 | 0.344 | 16.55 ± 10.75 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.319 ± 0.127 | 1.218 | 1.461 | 1.00 |
| `DVC` | 4.953 ± 0.146 | 4.858 | 5.122 | 3.76 ± 0.38 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.493 ± 0.145 | 1.326 | 1.583 | 1.00 |
| `DVC` | 32.049 ± 0.276 | 31.761 | 32.312 | 21.46 ± 2.09 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.294 ± 0.013 | 0.283 | 0.309 | 1.00 |
| `DVC` | 1.438 ± 0.021 | 1.421 | 1.462 | 4.89 ± 0.23 |
