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

duc add -r foo

duc commit
