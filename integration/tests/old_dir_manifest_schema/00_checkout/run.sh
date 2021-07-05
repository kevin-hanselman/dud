#!/bin/bash
set -euo pipefail

dud init

mkdir -p .dud/cache/{00,1f,49,50,b9,de,ea,ec}

echo '1' > .dud/cache/50/cc1102b1c612e6962547aacdcef9a400d4416ef8dd9388e885991853c400c9
echo '2' > .dud/cache/b9/a1a3183dd350f0e896d0f4b59c87e7bda8b1ed3a1af76afc86c1cb8f7cbbde
echo '3' > .dud/cache/49/124bf4f7f37328738ac34216a60dcd5f58bb198c5c3f6719b6becafb7e7882
echo '4' > .dud/cache/00/51fb8f5c8288b80163ea72ab2f482fc402ca9944b580aa57e694eedfc3ad1c
echo '5' > .dud/cache/ec/2c76a158a4c8ef05a9bfd56c9e9fa993fef6de549c9e0a62791a0e5c592eb1
echo '6' > .dud/cache/1f/ad12e6bdb0d30895fb817b05d8fd97be199d01a4e489433629778eae97d314
echo '7' > .dud/cache/de/dc9531a3ea216ed967a15ede743b4e4d1e9181bf24204cdd6c316171daa2e8

echo '{"Path":"foo","Contents":{"1.txt":{"Checksum":"50cc1102b1c612e6962547aacdcef9a400d4416ef8dd9388e885991853c400c9","Path":"1.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false},"2.txt":{"Checksum":"b9a1a3183dd350f0e896d0f4b59c87e7bda8b1ed3a1af76afc86c1cb8f7cbbde","Path":"2.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false},"3.txt":{"Checksum":"49124bf4f7f37328738ac34216a60dcd5f58bb198c5c3f6719b6becafb7e7882","Path":"3.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false},"4.txt":{"Checksum":"0051fb8f5c8288b80163ea72ab2f482fc402ca9944b580aa57e694eedfc3ad1c","Path":"4.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false},"5.txt":{"Checksum":"ec2c76a158a4c8ef05a9bfd56c9e9fa993fef6de549c9e0a62791a0e5c592eb1","Path":"5.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false},"bar":{"Checksum":"ec0388aaaeb55fce40181409513e2c5d9eaef6e402084b4145ae9d46a18c5f4e","Path":"bar","IsDir":true,"DisableRecursion":false,"SkipCache":false}}}' \
    > .dud/cache/ea/e23573e6e1d7724622bcdc2d5cdd3250c6e9a2ddd08dcee87e7347b4174979

echo '{"Path":"bar","Contents":{"4.txt":{"Checksum":"0051fb8f5c8288b80163ea72ab2f482fc402ca9944b580aa57e694eedfc3ad1c","Path":"4.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false},"5.txt":{"Checksum":"ec2c76a158a4c8ef05a9bfd56c9e9fa993fef6de549c9e0a62791a0e5c592eb1","Path":"5.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false},"6.txt":{"Checksum":"1fad12e6bdb0d30895fb817b05d8fd97be199d01a4e489433629778eae97d314","Path":"6.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false},"7.txt":{"Checksum":"dedc9531a3ea216ed967a15ede743b4e4d1e9181bf24204cdd6c316171daa2e8","Path":"7.txt","IsDir":false,"DisableRecursion":false,"SkipCache":false}}}' \
    > .dud/cache/ec/0388aaaeb55fce40181409513e2c5d9eaef6e402084b4145ae9d46a18c5f4e

cat << 'EOF' > stage.yaml
checksum: 44d3fb997e664019a0cdd58689c22304f7b8d2336e148417d9c5a418788b9d55
working-dir: .
outputs:
  foo:
    checksum: eae23573e6e1d7724622bcdc2d5cdd3250c6e9a2ddd08dcee87e7347b4174979
    is-dir: true
EOF

dud stage add stage.yaml

dud checkout
