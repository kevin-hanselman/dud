#!/bin/bash
set -euo pipefail

cd "$1"

dvc fetch data.dvc
