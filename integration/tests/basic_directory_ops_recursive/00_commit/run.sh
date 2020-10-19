#!/bin/bash

set -euo pipefail

dud init

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
  is-dir: true
  is-recursive: true
EOF
) > stage.yaml

dud add stage.yaml

dud commit
