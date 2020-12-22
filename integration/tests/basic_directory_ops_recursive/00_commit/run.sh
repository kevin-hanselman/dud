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

dud stage gen -o foo > stage.yaml

dud stage add stage.yaml

dud commit
