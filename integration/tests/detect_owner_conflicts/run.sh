#!/bin/bash
set -euo pipefail

dud init

dud stage gen -o foo.txt > foo.yaml

dud stage gen -o bar.txt > bar.yaml

dud stage add *.yaml

sed -i 's/foo/bar/' foo.yaml

if dud status; then
    echo 'Expected command to fail' 1>&2
    exit 1
fi
