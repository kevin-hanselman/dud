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
    foo:
        is-dir: true
        disable-recursion: true
EOF
) > stage.yaml

dud stage add stage.yaml

dud commit
