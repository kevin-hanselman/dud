#!/bin/bash
set -euo pipefail

cd "$1"

dvc status data.dvc
