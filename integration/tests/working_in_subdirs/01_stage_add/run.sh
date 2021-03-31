#!/bin/bash
set -euo pipefail

cd subdir

dud stage add stage.yaml

grep -q '^subdir/stage.yaml$' ../.dud/index
