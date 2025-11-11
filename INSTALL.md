# Installation

The steps below install Go and the libraries needed to run the GUI (Fyne) and activity tracking, then set up desktop files and example configs.

> These commands target Debian/Ubuntu. For other distros, install equivalent packages.

## Install prerequisites & tools
```bash
# install golang

sudo apt-get update

# install xprintidle to track activity
sudo apt install xprintidle

# Runtime libs (X11/Wayland + OpenGL)
sudo apt-get install -y \
  libgl1 libegl1 libgl1-mesa-dri \
  libx11-6 libxrandr2 libxcursor1 libxinerama1 libxi6 \
  libwayland-client0 libwayland-cursor0

# Build prerequisites
sudo apt-get install -y \
  golang build-essential pkg-config \
  libgl1-mesa-dev xorg-dev libwayland-dev

# run fyne demo to confirm that it works
go run fyne.io/demo@latest
# fyne cli
go install fyne.io/fyne/v2/cmd/fyne@latest
export PATH="$HOME/go/bin:$PATH"

# build and install desktop files
./scripts/install.sh

# copy config files (don't forget to edit them)
cp ./cfg/example.config.json ./cfg/config.json
cp ./cfg/example.tasks.json ./cfg/tasks.json
```

### Notes

- **Fyne demo check**: if `go run fyne.io/demo@latest` launches a demo window, your graphics/toolchain setup is good.
- **PATH**: ensure `~/go/bin` is on your `PATH` (the line above adds it for the current shell).
- **Desktop files**: `./scripts/install.sh` installs icons/`.desktop` entries so you can launch from your app menu.
- **Configs**: edit `./cfg/config.json` and `./cfg/tasks.json` after copying to match your email provider and task categories.

### Troubleshooting

- If the demo or app fails to start on Wayland, try running under XWayland or ensure the Wayland dev packages listed above are installed.
- On headless servers, report generation works, but the GUI requires a desktop session.
