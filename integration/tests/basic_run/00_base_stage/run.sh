#!/bin/bash

set -euo pipefail

(
cat <<EOF
command: seq 1 20 > base.txt
outputs:
- path: base.txt
EOF
) > base.yaml

dud init

dud add base.yaml

dud run base.yaml
