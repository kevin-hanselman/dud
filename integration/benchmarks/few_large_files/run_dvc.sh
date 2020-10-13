#!/bin/bash
set -euo pipefail

dvc add data/
dvc commit data.dvc
