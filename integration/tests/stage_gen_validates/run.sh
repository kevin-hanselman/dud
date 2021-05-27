#!/bin/bash
set -euo pipefail

dud init

if dud stage gen foo.yaml; then
    echo 1>&2 'expected failure due to no deps and no outs'
    exit 1
fi
