#!/bin/bash
set -euo pipefail

dud init

dud stage gen -o foo.bin > foo.yaml

dud stage add foo.yaml

grep -qF 'foo.yaml' .dud/index

dud stage rm foo.yaml

if test -s .dud/index; then
    echo >&2 'expected the index to be empty'
    exit 1
fi
