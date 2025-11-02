#!/bin/bash

# cd_to_project_dir changes the current working directory to the project root.
# It determines the script's directory, moves two levels up, and cd's there.
# The resulting directory is printed for verification.
# since we use {} it will run in the current shell and directory change will be kept
cd_to_project_dir() {
    # Get the directory of the script
    declare -g THIS_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[1]:-${BASH_SOURCE[0]}}")" && pwd)"
    # Go up two directories
    declare -g THIS_PROJECT_DIR="$(cd "$THIS_SCRIPT_DIR/.." && pwd)"
    # Change to the target directory
    cd "$THIS_PROJECT_DIR" || exit 1
    echo "Running in $(pwd)"
}

# Build Go binaries by name.
# Usage: build_go_binaries report tracker send-email
build_go_binaries() {
  # ensure build dir exists
  mkdir -p ./build

  local name src out
  for name in "$@"; do
    src="src/cmd/$name/main.go"
    out="./build/$name"

    if [[ ! -f "$src" ]]; then
      echo "Missing source for '$name': $src"
      return 1
    fi

    # Respect GOOS/GOARCH/GOFLAGS/CGO_ENABLED if caller sets them
    go build -o "$out" "$src"
    echo "Built $name"
  done
}
