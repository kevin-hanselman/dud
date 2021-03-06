#!/bin/bash

set -euo pipefail

dud init

dud stage gen -o base.txt 'seq 1 20 > base.txt' > base.yaml

dud stage add base.yaml

dud run base.yaml
