#!/bin/bash
set -euo pipefail
output_dir=$1
seq 1 4 | parallel --bar -I{} -P1 \
    dd if=/dev/urandom of="${output_dir}/{}.bin" bs=8M count=128 status=none
