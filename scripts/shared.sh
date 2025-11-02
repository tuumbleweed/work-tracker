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

# Install desktop entries by name (with or without .desktop).
# Usage: install_desktop_files work-tracker report
#        install_desktop_files work-tracker.desktop report.desktop
install_desktop_files() {
  # Resolve target dir (XDG default)
  local apps_dir="${XDG_DATA_HOME:-$HOME/.local/share}/applications"
  mkdir -p "$apps_dir"

  local name src base dest any_failed=0
  for name in "$@"; do
    # Allow names with or without .desktop
    if [[ "$name" == *.desktop ]]; then
      base="$name"
    else
      base="$name.desktop"
    fi

    src="./scripts/$base"
    dest="$apps_dir/$base"

    if [[ ! -f "$src" ]]; then
      echo "Missing desktop source: $src"
      any_failed=1
      continue
    fi

    cp -f "$src" "$dest"
    chmod +x "$dest" 2>/dev/null || true

    # Mark as "trusted" on GNOME/Nautilus if gio is available
    if command -v gio >/dev/null 2>&1; then
      gio set "$dest" "metadata::trusted" true 2>/dev/null || true
    fi

    echo "Installed $base → $dest"
  done

  # Refresh desktop database if available (harmless if not)
  if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "$apps_dir" >/dev/null 2>&1 || true
  fi

  return $any_failed
}
