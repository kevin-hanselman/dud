#!/bin/bash
set -euo pipefail

dud init

echo 'foo' > foo.txt

dud stage new -o foo.txt > stage.yaml

dud stage add stage.yaml

dud commit

echo 'bar' > bar.txt

dud stage new -o foo.txt -o bar.txt > stage.yaml

dud commit
