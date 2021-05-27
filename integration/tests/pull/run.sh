#!/bin/bash
set -euo pipefail

error() {
    echo 1>&2 "$1"
    exit 1
}

dud init

touch foo

dud stage gen -o foo > foo.yaml

dud stage add foo.yaml

dud commit

mkdir fake_remote

dud config set remote fake_remote

dud push

rm foo

dud pull

test -h foo || error "expected symlink, got regular file"

rm foo

# Fetch copy flag still works
dud pull -c

test -h foo && error "expected regular file, got symlink"

rm foo

# Passing a stage works
dud pull foo.yaml
