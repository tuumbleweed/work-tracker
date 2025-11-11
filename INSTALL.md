# Installation

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
