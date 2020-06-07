#!/usr/bin/env bash
set -euo pipefail

mkdir -p dvc/files && cd dvc
git init; dvc init
dvc config cache.type symlink
cp -r /datasets/$1 files/

export TIMEFORMAT=$'%3lR'
addTime=$((time $(dvc add --no-commit files/ > /dev/null 2>&1)) 2>&1)
commitTime=$((time $(dvc commit files.dvc > /dev/null 2>&1)) 2>&1)
rm -r files/
checkoutTime=$((time $(dvc checkout files.dvc > /dev/null 2>&1)) 2>&1)

diff /datasets/$1 files/$1
printf "Add: %s\nCommit: %s\nCheckout: %s\n" $addTime $commitTime $checkoutTime > /datasets/$1_dvc.txt
