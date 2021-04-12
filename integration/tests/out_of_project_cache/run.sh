#!/bin/bash
set -euo pipefail

cache_dir=/tmp/dud/oop_cache

dud init

dud config set cache "$cache_dir"

dud stage gen -o foo.txt > stage.yaml

echo 'foobar' > foo.txt

dud stage add stage.yaml

dud commit --copy

diff foo.txt "$cache_dir/53/4659321d2eea6b13aea4f4c94c3b4f624622295da31506722b47a8eb9d726c" >&2

rm -rf "$cache_dir"
