#!/bin/bash
set -euo pipefail

cd "$1"

dataset_src_dir=$2

if ! test -d '.dud'; then
    rm -rf data/

    rclone copy --progress "$dataset_src_dir" ./data

    dud init

    dud config set remote ./fake_remote

    dud stage gen -o ./data > data.yaml

    dud stage add data.yaml

    dud commit

    dud push
fi

rm -rf ./.dud/cache/*

sync
