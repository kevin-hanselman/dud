#!/bin/bash
set -euo pipefail

dud init

dud stage gen -o foo.txt > foo.yaml

dud stage gen -i foo.txt -o bar.txt > bar.yaml

dud stage add *.yaml

dud graph | dot -Tdot
