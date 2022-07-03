#!/bin/bash
set -euo pipefail

cd "$1"

dvc push data.dvc
