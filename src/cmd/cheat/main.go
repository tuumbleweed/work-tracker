package main

import (
	"flag"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"

	"work-tracker/src/pkg/cheat"
	"work-tracker/src/pkg/config"
	"work-tracker/src/pkg/util"
)

func main() {
	util.CheckIfEnvVarsPresent([]string{})
	configPath := flag.String("config", "./cfg/config.json", "Path to your configuration file.")
	workDir := flag.String("work-dir", "./out", "Directory for daily JSONL files")
	dateFlag := flag.String("date", time.Now().Format("2006-01-02"), "Day to update (YYYY-MM-DD; defaults to today in local time)")
	after := flag.Int64("after", 0, "Minutes between the last entry's finish and the new entry")
	minutes := flag.Int64("minutes", 0, "Minutes to log (required, positive integer)")
	activity := flag.Int64("activity", 100, "Active time percentage (0 through 100)")
	task := flag.String("task", "", "Task name (defaults to the last entry's task)")
	flag.Parse()
	config.InitializeConfig(*configPath)
	tl.Log(tl.Notice, palette.BlueBold, "%s cheat entrypoint. Config path: '%s'", "Running", *configPath)
	date, err := time.ParseInLocation("2006-01-02", *dateFlag, time.Local)
	if err != nil {
		xerr.NewErrorECOL(err, "invalid date: expected YYYY-MM-DD", "date", *dateFlag).QuitIf("error")
	}
	_, e := cheat.Append(*workDir, date, *after, *minutes, *activity, *task)
	e.QuitIf("error")
}
