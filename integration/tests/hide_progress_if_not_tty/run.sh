#!/bin/bash
set -euo pipefail

dud init

echo 'foo' > bar.txt

dud stage gen -o bar.txt > stage.yaml

dud stage add stage.yaml

dud commit 2> stderr.log

if test -s stderr.log; then
    >&2 echo 'stderr.log is not empty, got:'
    >&2 cat stderr.log
    exit 1
fi
