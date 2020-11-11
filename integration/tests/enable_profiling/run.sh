#!/bin/bash
set -euo pipefail

dud init

dud --profile status

if ! test -s dud.pprof; then
    echo 'profiling output does not exist or is empty' 1>&2
    exit 1
fi

dud --trace status

if ! test -s dud.trace; then
    echo 'tracing output does not exist or is empty' 1>&2
    exit 1
fi
