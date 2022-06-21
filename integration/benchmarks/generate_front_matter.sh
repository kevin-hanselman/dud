#!/bin/bash
set -euo pipefail

command -v dud &>/dev/null || { echo >&2 'dud not installed'; exit 1; }
command -v dvc &>/dev/null || { echo >&2 'dvc not installed'; exit 1; }

cat << EOF
# Benchmarks

## System Information

**OS**: $(uname -sr | grep -Po '\w+ \d+\.\d+')

**CPU**: $(grep -m1 '^model name' /proc/cpuinfo | cut -d':' -f2 | xargs)

**RAM**: $(free --giga | grep '^Mem' | awk '{print $2}') GB

**Go version**: $(go env GOVERSION | sed 's;^go;;')

**Dud version**: $(dud version | xargs)

**DVC version**: $(dvc version | head -n1 | cut -d':' -f2 | xargs)

DVC non-default configuration:

EOF

# Add spaces to each line to treat the output as a Markdown code block.
dvc config --list --global | awk '{ print "    " $0 }'

echo
