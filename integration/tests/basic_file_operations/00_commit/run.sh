#!/bin/bash

set -euo pipefail

duc init

echo 'foo' > bar.txt

duc add bar.txt

duc commit
