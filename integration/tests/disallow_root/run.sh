#!/bin/bash
set -euo pipefail

# Make sure we can run sudo non-interactively. If sudo needs a password, it'll
# error-out and the test will fail.
sudo -n id

if sudo -n dud init; then
    echo 'Expected command to fail' 1>&2
    exit 1
fi
