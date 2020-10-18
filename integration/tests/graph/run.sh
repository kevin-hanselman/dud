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
dependencies:
- path: foo.txt
outputs:
- path: bar.txt
EOF
) > bar.yaml

duc add *.yaml

duc graph | dot -Tdot
