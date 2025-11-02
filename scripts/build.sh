#!/bin/bash

set -eo pipefail # Exit immediately if any command returns a non-zero status

# source the shared functions
source "$(dirname "${BASH_SOURCE[0]}")/shared.sh"

cd_to_project_dir
build_go_binaries report tracker send-email
