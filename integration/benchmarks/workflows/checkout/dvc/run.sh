#!/bin/bash
set -euo pipefail

cd "$1"

dvc checkout data.dvc
