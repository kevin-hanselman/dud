#!/bin/bash
set -euo pipefail

dud init

dud stage gen -o foo.txt -- sleep 10 > sleep.yaml

dud stage add sleep.yaml

dud run &
dud_run_pid=$!

# Give 'dud run' enough time to create the lock file.
sleep 0.05

if dud status; then
    echo 1>&2 'TEST FAIL: expected second concurrent dud command to fail'
    exit 1
fi

# We CANNOT send SIGINT here, as non-interactive Bash won't send it to its job. TIL.
# https://unix.stackexchange.com/a/677742/49491
kill -SIGTERM "$dud_run_pid"
wait

# Having both of these tests is a bit redundant, but better safe than sorry.
err=0
if ! test -f .dud/lock; then
    echo 1>&2 'TEST FAIL: expected interrupted dud command to leave its lock file'
    err=$((err + 1))
fi

if dud status; then
    echo 1>&2 'TEST FAIL: expected orphaned lock file to cause new dud command to fail'
    err=$((err + 1))
fi

# Clean up orphaned 'dud run' process
pkill sleep

exit "$err"
