#!/bin/bash
set -euo pipefail

export BENCH_CPU=$(grep -m1 '^model name' /proc/cpuinfo | cut -d':' -f2 | xargs)
export BENCH_SYS_MEM_GB=$(free --giga | grep '^Mem' | awk '{print $2}')
export BENCH_OS=$(uname -srmo)

cat << EOF
# Benchmarks

**OS**: $BENCH_OS

**CPU**: $BENCH_CPU

**RAM**: $BENCH_SYS_MEM_GB GB
EOF
