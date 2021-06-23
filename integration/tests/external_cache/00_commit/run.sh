#!/bin/bash
set -euo pipefail

cache_dir=/tmp/dud_integration_tests/.external_cache

dud init

dud config set cache "$cache_dir"

dud stage gen -o foo.txt > stage.yaml

echo 'foobar' > foo.txt

dud stage add stage.yaml

dud commit
