#!/bin/bash
set -euo pipefail

dud init

echo 'bar' > foo.txt

dud stage gen -o foo.txt > foo.yaml

dud stage add foo.yaml

dud commit

mkdir fake_remote

dud config set remote fake_remote

dud push

rm foo.txt
rm -rf .dud/cache/*

dud pull
