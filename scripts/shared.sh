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
