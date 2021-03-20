#!/bin/bash
set -euo pipefail

cd "$1"

dataset_src_dir=$2

rm -rf .dvc/ data/ data.dvc

rclone copy --progress "$dataset_src_dir" ./data

dvc init --no-scm

sync
