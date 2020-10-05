#!/bin/bash

set -euo pipefail

(
cat <<EOF
command: seq 1 20 > base.txt
outputs:
- path: base.txt
EOF
) > base.yaml

duc init

duc add base.yaml

duc run base.yaml
