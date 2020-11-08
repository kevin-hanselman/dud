#!/bin/bash
set -euo pipefail

dud init

dud --profile status

if [ $(wc -l dud.prof) -eq 0 ]; then
    echo 'dud.prof is empty' 1>&2
    exit 1
fi
