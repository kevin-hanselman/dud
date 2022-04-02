#!/bin/bash
set -euo pipefail

dud init

mkdir -p subdir/subsubdir
cd subdir

fail() {
    echo >&2 "$*"
    exit 1
}

dud stage gen -w . -o foo.txt -- 'echo foobar > foo.txt' > stage.yaml

grep -Fq 'working-dir: subdir' stage.yaml || fail 'bad/missing working dir #1'

grep -Fq 'subdir/foo.txt:' stage.yaml || fail 'bad/missing output artifact'

# Excluding -w . should work too.
dud stage gen -o foo.txt -- 'echo foobar > foo.txt' | grep -Fq 'working-dir: subdir' \
    || fail 'bad/missing working dir #2'

dud stage gen -o subsubdir | grep -Fq 'is-dir: true' || fail 'bad/missing is-dir'
