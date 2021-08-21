#!/bin/bash
set -euo pipefail

dud init
dud stage gen -o myfile > myfile

set +e
error=$(dud stage add myfile 2>&1 1>/dev/null)

if [ "$?" -eq 0 ]; then
    echo >&2 'expected non-zero exit code'
    exit 1
fi
set -e

echo "$error" | grep -qF 'stage references itself in outputs'
