# Benchmarks

## System Information

**OS**: Linux 5.18

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Go version**: 1.18.3

**Dud version**: v0.3.1-9-gfb7994b

**DVC version**: 2.11.0 (pip)

DVC non-default configuration:

    core.analytics=false
    core.check_update=false
    cache.type=symlink

## Few large files

This dataset consists of four 1 GB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.018 ± 0.008 | 0.010 | 0.024 | 1.00 |
| `DVC` | 0.349 ± 0.008 | 0.340 | 0.355 | 18.99 ± 7.95 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.665 ± 0.013 | 0.656 | 0.680 | 1.00 |
| `DVC` | 3.669 ± 0.034 | 3.645 | 3.708 | 5.52 ± 0.12 |
### fetch

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 9.103 ± 0.248 | 8.826 | 9.307 | 1.09 ± 0.09 |
| `DVC` | 8.353 ± 0.666 | 7.953 | 9.122 | 1.00 |
### push

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 10.877 ± 1.856 | 8.885 | 12.558 | 2.19 ± 0.73 |
| `DVC` | 4.970 ± 1.416 | 3.399 | 6.147 | 1.00 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.021 ± 0.002 | 0.019 | 0.023 | 1.00 |
| `DVC` | 0.303 ± 0.011 | 0.294 | 0.315 | 14.21 ± 1.27 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.969 ± 0.025 | 1.941 | 1.988 | 1.00 |
| `DVC` | 11.483 ± 8.241 | 6.644 | 20.998 | 5.83 ± 4.19 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.583 ± 0.261 | 1.423 | 1.885 | 1.00 |
| `DVC` | 46.087 ± 6.321 | 39.050 | 51.285 | 29.11 ± 6.25 |
### fetch

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 9.711 ± 1.550 | 8.581 | 11.478 | 1.00 |
| `DVC` | 55.610 ± 1.263 | 54.709 | 57.054 | 5.73 ± 0.92 |
### push

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 8.685 ± 0.295 | 8.376 | 8.963 | 1.00 |
| `DVC` | 47.422 ± 0.592 | 46.853 | 48.034 | 5.46 ± 0.20 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.290 ± 0.041 | 0.266 | 0.337 | 1.00 |
| `DVC` | 1.075 ± 0.018 | 1.062 | 1.095 | 3.70 ± 0.52 |
