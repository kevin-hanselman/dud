#!/bin/bash
set -euo pipefail

num_files=100000
file_size_kb=1
out_dir=data

mkdir "$out_dir"

for file in $(seq 1 "$num_files"); do
    dd if=/dev/urandom of="$out_dir/$file.bin" bs=1K count="$file_size_kb" status=none
done
