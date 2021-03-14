#!/bin/bash
set -euo pipefail

dud init

dud --profile stage gen -o foo.txt

if ! test -s dud.pprof; then
    echo 'profiling output does not exist or is empty' 1>&2
    exit 1
fi

# Commands that fail should still generate profiling output.
dud --profile stage add foo.yaml || true

if ! test -s dud.pprof; then
    echo 'profiling output does not exist or is empty' 1>&2
    exit 1
fi

dud --trace stage gen -o foo.txt

if ! test -s dud.trace; then
    echo 'tracing output does not exist or is empty' 1>&2
    exit 1
fi

# Commands that fail should still generate tracing output.
dud --trace stage add foo.yaml || true

if ! test -s dud.trace; then
    echo 'tracing output does not exist or is empty' 1>&2
    exit 1
fi
