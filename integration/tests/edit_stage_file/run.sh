#!/bin/bash
set -euo pipefail

dud init

echo 'foo' > foo.txt

dud stage gen -o foo.txt > stage.yaml

dud stage add stage.yaml

dud commit

echo 'bar' > bar.txt

# TODO: We can't use stage gen alone because it will wipe out the checksums
# added to the YAML by the above commit. This is a use case for stage
# merge/update:
#
# dud stage gen -o bar.txt -o foo.txt | dud stage update stage.yaml
echo '  bar.txt:' >> stage.yaml

dud commit
