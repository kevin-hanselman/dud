# Benchmarks

## System Information

**OS**: Linux 5.18

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Go version**: 1.18.3

**Rclone version**: rclone v1.58.1

**Dud version**: v0.3.1-12-g3f98bdd

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
| `Dud` | 11.792 ± 0.519 | 11.224 | 12.244 | 1.56 ± 0.16 |
| `DVC` | 7.562 ± 0.703 | 6.865 | 8.270 | 1.00 |
### push

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 11.577 ± 0.336 | 11.340 | 11.961 | 1.00 |
| `DVC` | 12.794 ± 6.961 | 8.587 | 20.829 | 1.11 ± 0.60 |
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
| `Dud` | 7.227 ± 0.226 | 6.968 | 7.382 | 1.00 |
| `DVC` | 56.176 ± 2.097 | 53.845 | 57.910 | 7.77 ± 0.38 |
### push

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 5.644 ± 0.051 | 5.586 | 5.683 | 1.00 |
| `DVC` | 44.879 ± 0.184 | 44.745 | 45.089 | 7.95 ± 0.08 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.290 ± 0.041 | 0.266 | 0.337 | 1.00 |
| `DVC` | 1.075 ± 0.018 | 1.062 | 1.095 | 3.70 ± 0.52 |
