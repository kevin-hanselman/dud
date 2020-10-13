#!/bin/bash
set -euo pipefail

dvc init --no-scm
dvc config cache.type symlink

dvc add data/
dvc commit data.dvc
