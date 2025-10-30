// functions that logger relies on, so we can't put them in common
package logger

import (
	"fmt"
	"os"
	"runtime/debug"
)

// Cfg.LoggerFilePath == "" at the point of logger initialization, so
// it will just print without saving to log file
func CreateDirIfDoesntExist(path string) (err error, errMsg string) {
	Log(Info, BlueColor, "%s dir: '%s'", "Creating", path)
	if path == "" {
		Log(Verbose2, DimBlueColor, "%s", "Dir is an empty string, not creating")
		return nil, ""
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		Log(Verbose, DimCyanColor, "Dir '%s' doesn't exist", path)
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err, fmt.Sprintf("Unable to create dir: %s", path)
		}
	} else {
		Log(Info1, PurpleColor, "Dir '%s' already exists, not creating", path)
		return nil, ""
	}

	Log(Info1, GreenColor, "%s dir: '%s'", "Creating", path)

	return nil, ""
}



// Exit if error is encountered, otherwise do nothing
// this version does not rely on logger itself
func QuitIfErrorLoggerIndependent(err error, errMsg string) {
	if err != nil {
		fmt.Printf("\033[0;31m") // bold red
		fmt.Printf("%s\n", errMsg)
		fmt.Printf("%s\n", err.Error())
		fmt.Printf("\033[0m")
		fmt.Printf("\033[2;31m") // dim red
		debug.PrintStack()
		fmt.Printf("\033[0m")
		os.Exit(1)
	}
}
