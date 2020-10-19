#!/bin/bash
set -euo pipefail

dud init

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

dud add *.yaml

sed -i 's/foo/bar/' foo.yaml

if dud status; then
    echo 'Expected command to fail' 1>&2
    exit 1
fi
