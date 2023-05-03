#!/bin/bash
set -euo pipefail

dud init

for word in foo bar; do
    echo "$word" > "$word.txt"
    dud stage gen -o "$word.txt" > "$word.yaml"
    dud stage add "$word.yaml"
done

fail() {
    echo >&2 "$1"
    exit 1
}

# Print statuses for debugging in actual_output.txt.
dud status foo.yaml
if dud status foo.yaml | grep -q bar; then
    fail "'dud status foo.yaml' mentions bar"
fi

dud status bar.yaml
if dud status bar.yaml | grep -q foo; then
    fail "'dud status bar.yaml' mentions foo"
fi
