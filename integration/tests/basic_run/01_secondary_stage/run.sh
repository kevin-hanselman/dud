#!/bin/bash

set -euo pipefail

# Escaping this awk command is tricky in Bash so I'm sticking with a here doc.
# Normally I wouldn't mix-and-match here docs with `dud stage gen` because it
# hurts readability, but this is an integration test, so let's kick the tires.
(
cat <<'EOF'
command: awk '{print $1 * 2}' base.txt > second.txt
EOF
) > second.yaml

dud stage gen -d base.txt -o second.txt >> second.yaml

dud stage add second.yaml

dud run second.yaml
