// this package exists to hold Cfg global var
// can also change default values here for this project
package config

import (
	"os"
	"strings"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
	"work-tracker/src/pkg/util"
)

type Config struct {
	// parts present in configuration file (some of the parameters are generated during initilization process)
	Logger *logger.Config `json:"logger"`

	// those parametrs are initialized during InitializeConfig()
	CallerProgramName string `json:"caller_program_name,omitempty"`
}

var LocalConfig Config

func GetDefaultConfig() Config {
	return Config{}
}

// this function will quit when encountering error
// make sure to run common.LoadConfig first here since it's the earliest point we overriding log level from
func InitializeConfig(configPath string, logLevelOverride logger.LogLevel, logDirOverride string) {
	callerProgramName := strings.TrimPrefix(util.GetCallerProgramNamePanicWrapper(4), "work-tracker/") + os.Getenv("SERVICE_NAME_SUFFIX")
	logger.Log(logger.Important, logger.BoldBlueColor, "%s, config path: '%s', caller: '%s'", "Initializing", configPath, callerProgramName)
	util.SetEnvVar("SERVICE_NAME", callerProgramName)
	if configPath == "" {
		logger.Log(logger.Notice, logger.BoldCyanColor, "%s path is '', using %s", "Config", "default config")
		LocalConfig = GetDefaultConfig()
	} else if util.FileExists(configPath) {
		util.LoadConfig(configPath, &LocalConfig, logLevelOverride).QuitIf("error")
		// if certain part of the config is present in this project config file - override default config
		er.QuitIfError(logger.InitializeLogger(LocalConfig.Logger, logLevelOverride, logDirOverride))
	} else {
		logger.Log(logger.Notice, logger.BoldYellowColor, "%s path is '%s' but file %s, %s", "Config", configPath, "does not exist", "exiting...")
		os.Exit(1)
	}
	LocalConfig.CallerProgramName = util.GetCallerProgramNamePanicWrapper(4)
	logger.Log(logger.Important1, logger.BoldGreenColor, "%s, config path: '%s', caller: '%s'", "Initialized", configPath, callerProgramName)
}
