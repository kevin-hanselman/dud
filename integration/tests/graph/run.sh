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
dependencies:
- path: foo.txt
outputs:
- path: bar.txt
EOF
) > bar.yaml

dud add *.yaml

dud graph | dot -Tdot
