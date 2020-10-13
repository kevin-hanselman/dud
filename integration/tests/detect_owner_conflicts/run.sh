#!/bin/bash
set -euo pipefail

duc init

(
cat <<EOF
outputs:
- path: foo.txt
EOF
) > foo.yaml

(
cat <<EOF
outputs:
- path: bar.txt
EOF
) > bar.yaml

duc add *.yaml

sed -i 's/foo/bar/' foo.yaml

if duc status; then
    echo 'Expected command to fail' 1>&2
    exit 1
fi
