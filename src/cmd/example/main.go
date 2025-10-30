// you can add any code you want here but don't commit it.
// keep it empty for future projects and for use ase a template.
package main

import (
	"flag"

	"my-project/src/pkg/config"
	"my-project/src/pkg/logger"
	"my-project/src/pkg/util"
)

func main() {
	util.CheckIfEnvVarsPresent([]string{})
	// common flags
	logLevelOverride := flag.Int("log-level", -1, "Log level. Default is whatever value is in configuration file. Keep at -1 to not override.")
	logDirOverride := flag.String("log-dir", "", "File directory at which to save log files. Keep empty to use configuration file instead.")
	configPath := flag.String("config", "./cfg/config.json", "Path to your configuration file.")
	// program's custom flags
	// parse and init config
	flag.Parse()
	config.InitializeConfig(*configPath, logger.LogLevel(*logLevelOverride), *logDirOverride)

	logger.Log(
		logger.Notice, logger.BoldBlueColor, "%s example entrypoint. Config path: '%s'",
		"Running", *configPath,
	)
}
