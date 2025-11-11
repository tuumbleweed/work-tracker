package util

import (
	"os/exec"
	"path/filepath"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

func OpenInChrome(pathElements ...string) (e *xerr.Error) {
	emailReportPath := filepath.Join(pathElements...)
	tl.Log(tl.Notice, palette.BlueBold, "%s a file '%s' with %s", "Opening", emailReportPath, "google-chrome")

	err := exec.Command("google-chrome", "--new-window", emailReportPath).Start()
	if err != nil {
		return xerr.NewError(err, "Unable to open html report with google chrome", pathElements)
	}

	tl.Log(tl.Notice1, palette.GreenBold, "%s a file '%s' with %s", "Opened", emailReportPath, "google-chrome")
	return nil
}
