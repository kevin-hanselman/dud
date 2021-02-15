#!/bin/bash
set -euo pipefail
output_dir=$1
seq 1 20000 | parallel --bar -I{} -P4 \
    dd if=/dev/urandom of="${output_dir}/{}.bin" bs=100K count=1 status=none
