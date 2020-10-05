#!/bin/bash

set -euo pipefail

(
cat <<'EOF'
command: awk '{print $1 * 2}' base.txt > second.txt
outputs:
- path: second.txt
EOF
) > second.yaml

duc add second.yaml

duc run second.yaml
