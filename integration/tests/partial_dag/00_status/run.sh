#!/bin/bash
set -euo pipefail

dud init

for word in bish bash bosh; do
    echo "$word" > "$word.txt"
done

dud stage gen -o bish.txt > bish.yaml
dud stage add bish.yaml

dud stage gen -i bish.txt -o bash.txt -- 'echo fake bash command' > bash.yaml
dud stage add bash.yaml

dud stage gen -i bash.txt -o bosh.txt -- 'echo fake bosh command' > bosh.yaml
dud stage add bosh.yaml

fail() {
    echo >&2 "$1"
    exit 1
}

# Print statuses for debugging in actual_output.txt.
dud status bish.yaml
if dud status bish.yaml | grep -q bash; then
    fail "'dud status bish.yaml' mentions bash"
fi

dud status bash.yaml
if dud status bash.yaml | grep -q bosh; then
    fail "'dud status bash.yaml' mentions bosh"
fi
dud status bash.yaml | grep -q bish \
    || fail "'dud status bash.yaml' does not mention bish"
