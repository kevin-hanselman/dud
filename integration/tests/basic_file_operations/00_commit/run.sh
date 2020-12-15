#!/bin/bash

set -euo pipefail

dud init

echo 'foo' > bar.txt

dud stage gen -o bar.txt > stage.yaml

dud stage add stage.yaml

dud commit
