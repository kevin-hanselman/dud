#!/bin/bash
set -euo pipefail

dud init

dud stage gen -o foo.txt -- sleep 10 > sleep.yaml

dud stage add sleep.yaml

# This loop is mainly to ensure this concurrent test code is working well.
for _ in $(seq 1 10); do
    dud run &

    # Give 'dud run' enough time to start
    sleep 0.02

    if dud status; then
        echo 1>&2 'expected second concurrent dud command to fail'
        exit 1
    fi

    # Kill dud run's command and wait for it stop. This should cause 'run' to
    # fail, but we still expect Dud to release the project lock file.
    killall sleep
    wait
    if test -f .dud/lock; then
        echo 1>&2 'expected first dud command to clean up its lock file'
        exit 1
    fi
done
