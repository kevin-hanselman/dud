#!/bin/bash

set -euo pipefail

duc init

echo 'foo' > bar.txt

(
cat <<EOF
outputs:
- path: bar.txt
EOF
) > Ducfile

duc add Ducfile

duc commit
