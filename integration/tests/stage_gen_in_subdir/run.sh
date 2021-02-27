#!/bin/bash
set -euo pipefail

dud init

mkdir subdir
cd subdir

dud stage gen -w . -o foo.txt > stage.yaml

grep -Fq 'working-dir: subdir' stage.yaml

grep -Fq 'subdir/foo.txt:' stage.yaml
