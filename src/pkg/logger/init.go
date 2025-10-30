// this is where we initialize logger and it's global variables
package logger

import (
	"fmt"
	"os"
	"path"
)

// Pass logger.Config to this function.
// logger.Config should be part of every configuration file
// call this funcion when initializing your configuration file(s)
func InitializeLogger(localConfig *Config, logLevelOverride LogLevel, logDirOverride string) (err error, errMsg string) {
	// If not provided - just use defaultConfig
	if localConfig == nil {
		Log(Notice, BoldPurpleColor, "%s config is %s, keeping %s", "logger", "not provided", "default logger config")
		return nil, ""
	}

	Log(Notice, BlueColor, "%s with loaded config", "Initializing logger")
	Cfg = *localConfig

	if Cfg.LogTimeColor == "" {
		Cfg.LogTimeColor = DefaultCfg.LogTimeColor
	}

	if Cfg.UseTid == nil {
		Cfg.UseTid = DefaultCfg.UseTid
	}

	// appy default values to some parameters
	if Cfg.TimeFormat == "" {
		oldTimeFormat := Cfg.TimeFormat
		Cfg.TimeFormat = DefaultCfg.TimeFormat
		Log(Cfg.LogLevel, CyanColor, "%s was switched from '%s' to '%s'", "Time format", oldTimeFormat, Cfg.TimeFormat)
	}
	if Cfg.LogFileFormat == "" {
		Log(Cfg.LogLevel, CyanColor, "%s was switched from '%s' to '%s'", "Log file format", Cfg.LogFileFormat, DefaultCfg.LogFileFormat)
		Cfg.LogFileFormat = DefaultCfg.LogFileFormat
	}
	if Cfg.ContainerIdVarName == "" {
		Log(Cfg.LogLevel, CyanColor, "%s was switched from '%s' to '%s'", "ContainerIdVarName", Cfg.ContainerIdVarName, DefaultCfg.ContainerIdVarName)
		Cfg.ContainerIdVarName = DefaultCfg.ContainerIdVarName
	}

	Log(Cfg.LogLevel, CyanColor, "Config file's %s: '%s'", "log level", Cfg.LogLevel)
	// if logLevelOverride is not at it's default value. Overriding again after loading the file.
	if logLevelOverride != DontOverride {
		Log(Cfg.LogLevel, CyanColor, "Log level was switched from '%s' to '%s' (after loading config file)", Cfg.LogLevel, logLevelOverride)
		Cfg.LogLevel = logLevelOverride
	}

	Log(Notice, CyanColor, "%s is: '%s'", "Log directory", Cfg.LogDir)
	// if either log dir in config file or logFilePathOverride (--log-file-path flag) is set
	if Cfg.LogDir != "" || logDirOverride != "" {
		// override log dir before adding container directory on top
		if logDirOverride != "" {
			Log(Notice, CyanColor, "%s was switched from '%s' to '%s'", "Log directory", Cfg.LogDir, logDirOverride)
			Cfg.LogDir = logDirOverride
		}
		// if ContainerIdVarName is supplied - set Cfg.LogDir to Cfg.LogDir/container-id
		if Cfg.ContainerIdVarName != "NONE" {
			containerIdVar := os.Getenv(Cfg.ContainerIdVarName)
			// if env var with this name is not supplied - return error
			if containerIdVar == "" {
				err = fmt.Errorf("Env variable '%s' is not supplied!", Cfg.ContainerIdVarName)
				errMsg = fmt.Sprintf("You need to supply the value if Cfg.ContainerIdVarName (currently '%s') is not ''", Cfg.ContainerIdVarName)
				return err, errMsg
			}
			Cfg.LogDir = path.Join(Cfg.LogDir, containerIdVar)
		}
		// this function will change Cfg.LoggerFilePath and Cfg.LoggerFile
		err, errMsg := OpenLoggerFile(Cfg.LogDir)
		if err != nil {
			return err, errMsg
		}
	}

	Log(Notice1, GreenColor, "%s with loaded config", "Initialized logger")
	return nil, ""
}
