#!/bin/bash
set -euo pipefail

# Make sure we can run sudo non-interactively. If sudo needs a password, it'll
# error-out and the test will fail.
sudo -n id

# It seems that running sudo on Ubuntu doesn't let you use the PATH of the
# user. Without explicitly telling sudo/root where to find dud, we get the
# error "sudo: dud: command not found".
DUD=$(which dud)

# Use process substitution here instead of a pipe. grep -q will exit
# immediately after the first match, causing the command feeding grep to error
# out with SIGPIPE.
# See: https://stackoverflow.com/a/19120674/857893
grep -Fq WARNING <(sudo -n "$DUD")
