// keep this file self-contained for now; we'll split into a package later.
package main

import (
	"flag"
	"time"

	"work-tracker/src/pkg/config"
	"work-tracker/src/pkg/logger"
	"work-tracker/src/pkg/report"
	"work-tracker/src/pkg/util"
)

func main() {
	util.CheckIfEnvVarsPresent([]string{})

	// common flags
	logLevelOverride := flag.Int("log-level", -1, "Log level. Default is whatever value is in configuration file. Keep at -1 to not override.")
	logDirOverride := flag.String("log-dir", "", "File directory at which to save log files. Keep empty to use configuration file instead.")
	configPath := flag.String("config", "./cfg/config.json", "Path to your configuration file.")

	// program's custom flags
	flagStart := flag.String("start", "", "Start date (inclusive) in DD-MM-YYYY; empty => this Monday")
	flagEnd := flag.String("end", "", "End date (inclusive) in DD-MM-YYYY; empty => this Sunday (or start if start set)")
	flagInputDir := flag.String("dir", "./out", "Directory with day JSONL files")
	flagOutputPath := flag.String("output", "./out/report.html", "Path to write the HTML report")
	flagTZ := flag.String("tz", "America/Bogota", "IANA timezone for week boundaries and display")
	flagBarRef := flag.Duration("ref", 12*time.Hour, "Reference duration for the horizontal marker line (N hours)")
	flagSmooth := flag.Float64("smooth", 0.0, "Activity smoothing in [0..1], 0=linear, 1=strong")

	// parse and init config
	flag.Parse()
	config.InitializeConfig(*configPath, logger.LogLevel(*logLevelOverride), *logDirOverride)

	logger.Log(logger.Notice, logger.BoldBlueColor, "%s report entrypoint. Config path: '%s'", "Running", *configPath)

	// Resolve TZ + date range
	_, startDate, endDate, e := report.ResolveRange(*flagTZ, *flagStart, *flagEnd)
	e.QuitIf("error")

	// Build the report
	e = report.BuildReport(*flagInputDir, startDate, endDate, *flagOutputPath, *flagBarRef, *flagSmooth)
	e.QuitIf("error")

	// Open in Chrome
	e = util.OpenInChrome(*flagOutputPath)
	e.QuitIf("error")
}
