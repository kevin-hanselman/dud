#!/bin/bash
set -euo pipefail

error() {
    echo "$*" >&2
    exit 1
}

assert_cache_dir_equals() {
    local want=$1
    local cache_dir
    cache_dir=$(dud config get cache | xargs)
    if [ "$cache_dir" != "$want" ]; then
        error "cache dir = '$cache_dir', want '$want'"
    fi
}

# Delete the user Dud config dir to prevent other tests from unintentionally
# using it.
trap 'rm -rf ~/.config/dud' EXIT

dud init
assert_cache_dir_equals '.dud/cache'

dud config path | grep -q "^$(pwd)/.dud/config.yaml$"
dud config path --user | grep -q "^$HOME/.config/dud/config.yaml$"

dud config set --user cache 'user_cache'
assert_cache_dir_equals 'user_cache'

echo 'cache: other/user_cache' > ~/.config/dud/config.yaml
assert_cache_dir_equals 'other/user_cache'

export XDG_CONFIG_HOME="$HOME/my_overridden_config_dir"
mkdir -p "${XDG_CONFIG_HOME}/dud"
echo 'cache: xdg_override' > "$XDG_CONFIG_HOME/dud/config.yaml"
assert_cache_dir_equals 'xdg_override'

rm '.dud/config.yaml'
dud config set cache '/project/override'
assert_cache_dir_equals '/project/override'
