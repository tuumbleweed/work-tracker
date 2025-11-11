// you can add any code you want here but don't commit it.
// keep it empty for future projects and for use ase a template.
package main

import (
	"flag"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"

	"work-tracker/src/pkg/config"
	"work-tracker/src/pkg/util"
)

func main() {
	util.CheckIfEnvVarsPresent([]string{})
	// common flags
	configPath := flag.String("config", "./cfg/config.json", "Path to your configuration file.")
	// program's custom flags
	// parse and init config
	flag.Parse()
	config.InitializeConfig(*configPath)

	tl.Log(
		tl.Notice, palette.BlueBold, "%s example entrypoint. Config path: '%s'",
		"Running", *configPath,
	)
}
