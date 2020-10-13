#!/bin/bash
set -euo pipefail

duc init

(
cat <<EOF
outputs:
- path: data
  is-dir: true
  is-recursive: true
EOF
) > data.yaml

duc add data.yaml
