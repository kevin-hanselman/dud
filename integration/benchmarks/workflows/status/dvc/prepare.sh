#!/bin/bash
set -euo pipefail

cd "$1"

dataset_src_dir=$2

if ! test -d .dvc; then
    rm -rf data/

    rclone copy --progress "$dataset_src_dir" ./data

    dvc init --no-scm

    dvc add data/
fi

sync
