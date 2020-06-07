#!/usr/bin/env bash
set -euo pipefail

mkdir -p duc/files && cd duc
duc init
cp -r /datasets/$1 files/

export TIMEFORMAT=$'%3lR'
addTime=$((time $(duc add -r files/ > /dev/null 2>&1)) 2>&1)
commitTime=$((time $(duc commit > /dev/null 2>&1)) 2>&1)
rm -r files/
checkoutTime=$((time $(duc checkout > /dev/null 2>&1)) 2>&1)

diff /datasets/$1 files/$1
printf "Add: %s\nCommit: %s\nCheckout: %s\n" $addTime $commitTime $checkoutTime > /datasets/$1_duc.txt
