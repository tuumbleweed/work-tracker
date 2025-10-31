package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
)

// since we call this only when initializing config at the start of the program
// it's ok to quit on error
func GetCallerProgramNamePanicWrapper(skip int) (callerProgramName string) {
	callerProgramName, err, errMsg := GetCallerProgramName(skip)
	er.QuitIfError(err, errMsg)
	logger.Log(logger.Info, logger.CyanColor, "%s name: '%s'", "Caller program", callerProgramName)

	return callerProgramName
}

// returns $PWD/current_main.go_dir
func GetCallerProgramName(skip int) (callerProgramName string, err error, errMsg string) {
	callerFileDirBase, err, errMsg := getCurrentFileDirectory(skip)
	if err != nil {
		return "", err, errMsg
	}
	pwdDir, err := os.Getwd()
	if err != nil {
		return "", err, "Unable to os.Getwd()"
	}
	pwdDirBase := filepath.Base(pwdDir)
	callerProgramName = fmt.Sprintf("%s/%s", pwdDirBase, callerFileDirBase)

	return callerProgramName, nil, ""
}

func getCurrentFileDirectory(skip int) (callerFileDirBase string, err error, errMsg string) {
	// Get the caller's file path
	_, callerFilePath, _, ok := runtime.Caller(skip)
	if !ok {
		return "", fmt.Errorf("Unable to get the current file"), "Terminating the program"
	}
	logger.Log(logger.Verbose, logger.DimCyanColor, "%s path: '%s'", "Caller file", callerFilePath)

	// Get the directory from the file path
	callerFileDir := filepath.Dir(callerFilePath)
	return filepath.Base(callerFileDir), nil, ""
}

func GetPackageName() string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc).Name()
	parts := strings.Split(fn, "/")
	last := parts[len(parts)-1]
	if i := strings.Index(last, "."); i != -1 {
		return last[:i]
	}
	return last
}
