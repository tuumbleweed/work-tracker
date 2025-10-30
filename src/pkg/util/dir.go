package util

import (
	"os"

	er "my-project/src/pkg/error"
	"my-project/src/pkg/logger"
)

func CreateDirIfDoesntExist(path string) (e *er.Error) {
	if path == "" {
		logger.Log(logger.Detailed, logger.DimBlueColor, "Path '%s' is empty, %s", path, "not creating")
		return nil
	}
	logger.Log(logger.Detailed, logger.DimBlueColor, "Creating '%s' dir", path)
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return er.NewErrorECOL(err, "Unable to create a directory", "directory path", path)
	}

	return nil
}
