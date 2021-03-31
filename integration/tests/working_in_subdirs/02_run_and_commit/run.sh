#!/bin/bash
set -euo pipefail

cd subdir

dud run stage.yaml

dud commit
