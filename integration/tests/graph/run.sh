#!/bin/bash
set -euo pipefail

dud init

dud stage new -o foo.txt > foo.yaml

dud stage new -d foo.txt -o bar.txt > bar.yaml

dud stage add *.yaml

dud graph | dot -Tdot
