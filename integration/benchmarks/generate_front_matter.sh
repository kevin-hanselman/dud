#!/bin/bash
set -euo pipefail

BENCH_CPU=$(grep -m1 '^model name' /proc/cpuinfo | cut -d':' -f2 | xargs)
BENCH_SYS_MEM_GB=$(free --giga | grep '^Mem' | awk '{print $2}')
# Strip the minor version and misc qualifiers
BENCH_OS=$(uname -srm | sed -e 's;\([0-9]\.[0-9]\+\)\.\S\+;\1;')

cat << EOF
# Benchmarks

**OS**: $BENCH_OS

**CPU**: $BENCH_CPU

**RAM**: $BENCH_SYS_MEM_GB GB

**Dud version**: $(dud --version | xargs)

**DVC version**: $(dvc version | head -n1 | xargs)
EOF
