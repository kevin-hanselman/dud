#!/bin/bash
set -euo pipefail

duc init

echo 'foo' > foo.txt

(
cat <<EOF
outputs:
- path: foo.txt
EOF
) > stage.yaml

duc add stage.yaml

duc commit

echo 'bar' > bar.txt

echo '- path: bar.txt' >> stage.yaml

duc commit
