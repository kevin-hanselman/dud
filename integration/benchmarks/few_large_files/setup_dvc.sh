#!/bin/bash
set -euo pipefail

dvc init --no-scm
dvc config cache.type symlink
