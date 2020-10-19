#!/bin/bash

set -euo pipefail

dud init

echo 'foo' > bar.txt

(
cat <<EOF
outputs:
- path: bar.txt
EOF
) > stage.yaml

dud add stage.yaml

dud commit --copy
