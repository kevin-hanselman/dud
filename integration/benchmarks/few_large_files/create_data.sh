#!/bin/bash
set -euo pipefail

num_files=5
file_size_mb=1000
out_dir=data

mkdir "$out_dir"

for file in $(seq 1 "$num_files"); do
    dd if=/dev/urandom of="$out_dir/$file.bin" bs=1M count="$file_size_mb"
done
