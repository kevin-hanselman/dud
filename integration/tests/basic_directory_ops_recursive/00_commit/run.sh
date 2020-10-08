#!/bin/bash

set -euo pipefail

duc init

mkdir -p foo/bar

for n in $(seq 1 5); do
    echo "$n" > "foo/$n.txt"
done

for n in $(seq 4 7); do
    echo "$n" > "foo/bar/$n.txt"
done

(
cat <<EOF
outputs:
- path: foo
  isdir: true
  isrecursive: true
EOF
) > Ducfile

duc add Ducfile

duc commit
