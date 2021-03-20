#!/bin/bash
set -euo pipefail

cd "$1"

dataset_src_dir=$2

rm -rf .dud/ data/

rclone copy --progress "$dataset_src_dir" ./data

dud init

dud stage gen -o ./data > data.yaml

dud stage add data.yaml

sync
