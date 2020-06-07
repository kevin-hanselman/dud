#!/usr/bin/env bash
set -euo pipefail

# cifar 10
curl -sSL https://www.cs.toronto.edu/~kriz/cifar-10-binary.tar.gz | tar -xzC datasets/

# cifar 100
curl -sSL https://www.cs.toronto.edu/~kriz/cifar-100-binary.tar.gz | tar -xzC datasets/
