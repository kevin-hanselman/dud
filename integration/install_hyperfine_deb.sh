#!/bin/bash
set -euo pipefail

# Only considering 64-bit
arch=amd64
case "$(uname -m)" in
    *arm*|*aarch*)
        arch=arm64
        ;;
esac

deb_url=$(
    curl -sS https://api.github.com/repos/sharkdp/hyperfine/releases/latest \
    | jq -r '.assets[] | .browser_download_url' \
    | grep -v 'musl' \
    | grep "$arch\.deb$"
)

echo "using '$deb_url'"

curl -Lo hyperfine.deb "$deb_url"
dpkg -i hyperfine.deb
rm hyperfine.deb
