package util

import (
	"encoding/json"
	"os"

	er "my-project/src/pkg/error"
	"my-project/src/pkg/logger"
)

/*
Load a config file located at filePath
Pass a pointer to whatever config variable you need.
Make sure that file matches the config type.

It's ok to use logger before loading it's config. It still can print but won't save log line to a file

logLevelOverride is used to print all messages (by default log level is 50 so only first message is printed)
pass -1 if you don't want to change logging level from it's default value
need this since config has no effect yet during LoadConfig.

We keep it in util package in order to reuse it in case of projects with multiple config packages
*/
func LoadConfig(filePath string, config any, logLevelOverride logger.LogLevel) (e *er.Error) {
	logger.Log(logger.Notice, logger.BlueColor, "%s config '%s'", "Loading", filePath)
	// override log level here first in order for common.LoadConfig to be able to print Info1 and above
	// within LoadConfig function
	if logLevelOverride != logger.DontOverride {
		logger.Log(logger.Cfg.LogLevel, logger.CyanColor, "%s was switched from '%s' to '%s'", "Log level", logger.Cfg.LogLevel, logLevelOverride)
		logger.Cfg.LogLevel = logLevelOverride
	}

	byteValue, err := os.ReadFile(filePath)
	if err != nil {
		return er.NewErrorECOL(err, "Unable to read a file", "file path", filePath)
	}

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return er.NewErrorECOL(err, "Unable to json.Unmarshl file contents to config", "file contents", string(filePath))
	}

	logger.Log(logger.Notice1, logger.GreenColor, "%s config '%s'", "Loaded", filePath)
	return nil
}
