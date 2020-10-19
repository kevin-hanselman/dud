#!/bin/bash
set -euo pipefail

dud init

echo 'foo' > foo.txt

(
cat <<EOF
outputs:
- path: foo.txt
EOF
) > stage.yaml

dud add stage.yaml

dud commit

echo 'bar' > bar.txt

echo '- path: bar.txt' >> stage.yaml

dud commit
