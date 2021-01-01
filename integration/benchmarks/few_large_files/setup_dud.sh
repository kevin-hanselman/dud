#!/bin/bash
set -euo pipefail

dud init

dud stage gen -o data > data.yaml

dud stage add data.yaml
