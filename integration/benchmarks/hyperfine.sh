#!/bin/bash
set -euo pipefail

workspace_dir=$1
dataset_dir=$2
workflow_dir=$3
base_output_dir=$4

dataset_name=$(basename "$dataset_dir")
workflow_name=$(basename "$workflow_dir")

output_dir="${base_output_dir}/${dataset_name}/${workflow_name}"

mkdir -p "$output_dir"

echo -e "### $workflow_name\n" > "${output_dir}/header.md"

hyperfine \
    --max-runs 3 \
    --time-unit second \
    --show-output \
    --command-name 'Dud' \
    --command-name 'DVC' \
    --prepare "$workflow_dir/dud/prepare.sh $workspace_dir $dataset_dir" \
    --prepare "$workflow_dir/dvc/prepare.sh $workspace_dir $dataset_dir" \
    "$workflow_dir/dud/run.sh $workspace_dir" \
    "$workflow_dir/dvc/run.sh $workspace_dir" \
    --export-markdown "${output_dir}/table.md"
