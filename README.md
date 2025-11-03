# Work Tracker

An app tracking work time and activity.
Can generate an HTML report, open it with google-chrome browser.
Can also send that report over email using one of the providers like Mailgun, Sendgrid, Amazon SES.

## Installation
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

## TODO

#### Work Tracker
- ~~Count time for each task to display in the table~~
- ~~Update table hours per each task dynamically~~
- ~~Highlight the row when running a task~~
- ~~Change the way we measure activity~~
    - ~~set it to 100% activity so long as time since any key was pressed is lower
than the activity poll window~~
- Separate UI ticks from activity ticks
- Organize UI code in a better way, currently a mess, especially button handling.

#### Reporting
- ~~Add an HTML report.~~
    - ~~HTML report is saved to file and then opened right away with chrome browser.~~
    - ~~Should be able to generate reports for 1-360 days.~~
    - ~~It should still look nice for both weekly, quarterly and yearly reports.~~
    - ~~Should contain bar charts~~
        - ~~Time by task~~
        - ~~Time*activity~~
    - ~~Move report code to it's separate package~~
    - ~~Send an email option~~
    - ~~Make sure that large reports like yearly and quarterly also look bearable~~

#### Installing
- ~~Add desktop entries~~
    - ~~Work Tracker~~
    - ~~Report~~
- ~~Add icons~~
- ~~Add install.sh script~~
- ~~Make .desktop file templates to not hard-code paths~~
