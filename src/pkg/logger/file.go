package logger

import (
	"fmt"
	"os"
	"path"
	"time"
)

/*
This will initiate json logging file with name like Cfg.LogFileFormat and contents like those:
{"t": timestamp, "tid": yourthreadid, "l": loglevel, "msg": "log message with \033[0;32mcolor\033[0m"}
{"t": timestamp, "tid": yourthreadid, "l": loglevel, "msg": "another log message with \033[0;32mcolor\033[0m"}

This function is only called if an option to save to logger file is specified when initializing logger.
After this function is ran - you can check Cfg.FileIsOpen == true to see if file is open.

This function will change Cfg.LoggerFilePath and Cfg.LoggerFile
*/
func OpenLoggerFile(logDir string) (err error, errMsg string) {
	err, errMsg = CreateDirIfDoesntExist(logDir)
	if err != nil {
		return err, errMsg
	}

	Cfg.LoggerFilePath = path.Join(logDir, time.Now().Format(Cfg.LogFileFormat))

	Log(Notice, BoldBlueColor, "%s log file '%s'", "Creating", Cfg.LoggerFilePath)
	Cfg.LoggerFile, err = os.OpenFile(Cfg.LoggerFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err, fmt.Sprintf("Unable to open file: '%s'", Cfg.LoggerFilePath)
	}
	Cfg.FileIsOpen = true

	Log(Notice1, BoldGreenColor, "%s log file '%v'", "Created", Cfg.LoggerFilePath)
	return nil, ""
}

// this will close logger.LoggerFile
func CloseLogFile() {
	Cfg.LoggerFile.Close()
}
