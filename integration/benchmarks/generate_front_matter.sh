#!/bin/bash
set -euo pipefail

assert_installed() {
    command -v "$1" &>/dev/null || { echo >&2 "$1 not installed"; exit 1; }
}

assert_installed dud
assert_installed dvc
assert_installed rclone

cat << EOF
# Benchmarks

## System Information

**OS**: $(uname -sr | grep -Po '.+ \d+\.\d+')

**CPU**: $(grep -m1 '^model name' /proc/cpuinfo | cut -d':' -f2 | xargs)

**RAM**: $(free --giga | awk '/^Mem:/ {print $2}') GB

**Go version**: $(go env GOVERSION | sed 's;^go;;')

**Rclone version**: $(rclone version | head -n1 | xargs)

**Dud version**: $(dud version)

**DVC version**: $(dvc version | head -n1 | cut -d':' -f2 | xargs)

DVC non-default configuration:

EOF

# Add spaces to each line to treat the output as a Markdown code block.
dvc config --list --global | awk '{ print "    " $0 }'

echo
