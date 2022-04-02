#!/bin/bash
set -euo pipefail

BENCH_CPU=$(grep -m1 '^model name' /proc/cpuinfo | cut -d':' -f2 | xargs)
BENCH_SYS_MEM_GB=$(free --giga | grep '^Mem' | awk '{print $2}')
# Strip the minor version and misc qualifiers
BENCH_OS=$(uname -sr | sed -e 's;\([0-9]\.[0-9]\+\)\.\S\+;\1;')

command -v dud &>/dev/null || { echo >&2 'dud not installed'; exit 1; }
command -v dvc &>/dev/null || { echo >&2 'dvc not installed'; exit 1; }

cat << EOF
# Benchmarks

**OS**: $BENCH_OS

**CPU**: $BENCH_CPU

**RAM**: $BENCH_SYS_MEM_GB GB

**Software versions**:

Dud version: $(dud version | xargs)

$(dvc version | head -n1 | xargs)

DVC non-default configuration:

EOF

# Add spaces to each line to treat the output as a Markdown code block.
dvc config --list --global | awk '{ print "    " $0 }'

echo
