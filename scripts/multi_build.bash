#!/usr/bin/env bash

set -euo pipefail
OPTIND=1         # Reset in case getopts has been used previously in the shell.

declare -xr DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
declare -xr parent_dir="${DIR%/*}"
declare -xr gparent_dir="${parent_dir%/*}"

. "$DIR/set_globals.bash"
. "$DIR/clean_dirs.bash"
. "$DIR/ncpus.bash"

# Config
declare -ri n="${n:-3}"
declare -r ip_start="${ip_start:-127.0.0.1}"
declare -r subnet="${subnet:-16}"
declare -r ip_range="$ip_start/$subnet"
declare -r entry="${entry:-main}" # you may use main_profile here to enable profiling
# e.g.
# n=3 entry=main_profile BUILD_DIR="$PWD" ./scripts/multi.bash

# Install deps
"$DIR/docker/install_deps.bash"

declare debug=0
declare gc_flags=
while getopts "d" opt; do
    case "$opt" in
    d)  debug=1
        ;;
    esac
done

shift $((OPTIND-1))

[ "${1:-}" = "--" ] && shift

# Use -tags="netgo multi" in bgo build below to build multu dag1 version for testing
declare args="-X github.com/SamuelMarks/dag1/src/version.GitCommit=$(git rev-parse HEAD)"
if [ "$TARGET_OS" == "linux" ]; then
  args="$args -linkmode external -extldflags -static"
fi

if [ "$debug" == 0 ]; then
  args="$args -s -w"
else
  gc_flags="all=-N -l"
fi

env GOOS="$TARGET_OS" GOARCH=amd64 go build -tags="netgo multi debug" -ldflags "$args" -o dag1_"$TARGET_OS" -gcflags "$gc_flags" "$parent_dir/cmd/dag1/$entry.go" || exit 1

# Create peers.json and dag1_data_dir if needed
if [ ! -d "$DATAL_DIR/dag1_data_dir" ]; then
    "$GOPATH/bin/batch-ethkey" -dir "$BUILD_DIR/nodes" -network "$ip_start" -inc-port -n "$n" --port-start 12001
    cat "$BUILD_DIR/nodes/peers.json"> "$PEERS_DIR/peers.json"
    cat "$PEERS_DIR/peers.json"
    cp -rv "$BUILD_DIR/nodes" "$DATAL_DIR/dag1_data_dir"
    cp -v "$PEERS_DIR/peers.json" "$DATAL_DIR/dag1_data_dir/"
fi
