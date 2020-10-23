#!/bin/bash
set -euo pipefail

dud init

(
cat <<EOF
outputs:
- path: data
  is-dir: true
  is-recursive: true
EOF
) > data.yaml

dud stage add data.yaml
