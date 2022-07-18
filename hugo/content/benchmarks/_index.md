# Benchmarks

## System Information

**OS**: Linux 5.18

**CPU**: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz

**RAM**: 16 GB

**Go version**: 1.18.4

**Rclone version**: rclone v1.59.0

**Dud version**: v0.4.0

**DVC version**: 2.13.0 (pip)

DVC non-default configuration:

    core.analytics=false
    core.check_update=false
    cache.type=symlink

## Few large files

This dataset consists of four 1 GB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.023 ± 0.000 | 0.022 | 0.023 | 1.00 |
| `DVC` | 0.338 ± 0.013 | 0.326 | 0.352 | 14.92 ± 0.64 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.677 ± 0.030 | 0.644 | 0.701 | 1.00 |
| `DVC` | 6.483 ± 0.012 | 6.471 | 6.495 | 9.57 ± 0.42 |
### fetch

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 11.876 ± 1.790 | 10.703 | 13.936 | 1.53 ± 0.43 |
| `DVC` | 7.764 ± 1.813 | 5.759 | 9.288 | 1.00 |
### push

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 10.755 ± 0.240 | 10.538 | 11.012 | 1.52 ± 0.30 |
| `DVC` | 7.082 ± 1.393 | 5.987 | 8.650 | 1.00 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.018 ± 0.006 | 0.011 | 0.022 | 1.00 |
| `DVC` | 0.298 ± 0.006 | 0.291 | 0.303 | 16.57 ± 5.21 |
## Many small files

This dataset consists of twenty thousand 100 KB files in a single directory.

### checkout

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.262 ± 0.430 | 1.013 | 1.759 | 1.00 |
| `DVC` | 9.021 ± 0.063 | 8.983 | 9.093 | 7.15 ± 2.44 |
### commit

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 1.668 ± 0.030 | 1.636 | 1.695 | 1.00 |
| `DVC` | 45.934 ± 5.117 | 40.035 | 49.171 | 27.54 ± 3.11 |
### fetch

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 7.605 ± 0.042 | 7.563 | 7.647 | 1.00 |
| `DVC` | 57.540 ± 2.498 | 56.029 | 60.422 | 7.57 ± 0.33 |
### push

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 9.896 ± 4.191 | 7.438 | 14.735 | 1.00 |
| `DVC` | 44.301 ± 0.245 | 44.072 | 44.558 | 4.48 ± 1.90 |
### status

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `Dud` | 0.283 ± 0.012 | 0.270 | 0.295 | 1.00 |
| `DVC` | 1.575 ± 0.015 | 1.562 | 1.592 | 5.56 ± 0.24 |
