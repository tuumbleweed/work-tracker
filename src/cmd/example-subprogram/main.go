// in case you need to create an entrypoint with multiple subprograms
package main

import (
	"flag"
	"os"

	"work-tracker/src/pkg/config"
	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
	"work-tracker/src/pkg/util"
)

func example(subprogram string, flags []string) {
	util.CheckIfEnvVarsPresent([]string{})
	// common flags
	subprogramCmd := flag.NewFlagSet(subprogram, flag.ExitOnError)
	logLevelOverride := subprogramCmd.Int("log-level", -1, "Log level. Default is whatever value is in configuration file. Keep at -1 to not override.")
	logDirOverride := subprogramCmd.String("log-dir", "", "File directory at which to save log files. Keep empty to use configuration file instead.")
	configPath := subprogramCmd.String("config", "./cfg/config.json", "Path to your configuration file.")
	// program's custom flags
	// parse and init config
	er.QuitIfError(subprogramCmd.Parse(flags), "Unable to subprogramCmd.Parse")
	config.InitializeConfig(*configPath, logger.LogLevel(*logLevelOverride), *logDirOverride)

	logger.Log(
		logger.Notice, logger.BoldBlueColor, "%s example-subprogram entrypoint. Subprogram: '%s'. Config path: '%s'",
		"Running", subprogram, *configPath,
	)
}

func main() {
	// Check if there are enough arguments
	if len(os.Args) < 2 {
		logger.Log(logger.Error, logger.RedColor, "Usage: %s", "go run src/cmd/example-subprogram/main.go subprogram_name(for exampe first-example)")
		os.Exit(0)
	}
	subprogram := os.Args[1]
	flags := os.Args[2:]

	// Switch subprogram based on the first argument
	switch subprogram {
	case "first-example":
		example(subprogram, flags)
	default:
		logger.Log(logger.Error, logger.RedColor, "Unknown subprogram: %s", subprogram)
		os.Exit(0)
	}
}
