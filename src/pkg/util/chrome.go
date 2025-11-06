package util

import (
	"os/exec"
	"path/filepath"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
)

func OpenInChrome(pathElements... string) (e *er.Error) {
	emailReportPath := filepath.Join(pathElements...)
	logger.Log(logger.Notice, logger.BoldBlueColor, "%s a file '%s' with %s", "Opening", emailReportPath, "google-chrome")

	err := exec.Command("google-chrome", "--new-window", emailReportPath).Start()
	if err != nil {
		return er.NewError(err, "Unable to open html report with google chrome", pathElements)
	}

	logger.Log(logger.Notice1, logger.BoldGreenColor, "%s a file '%s' with %s", "Opened", emailReportPath, "google-chrome")
	return nil
}
