// here we keep our logger's configuration struct
// it's populated during the initialization process
package logger

import "os"

type Config struct {
	// those are parameters that are set using a config file

	// log level to print to stderr, don't print any message with log level below this one
	// for example by setting log level to Info5 (int value 55) every message
	// with log level value bigger than 55, for example Verbose3 (int value 73)
	// will not be printed to stderr
	// lines are always saved to file if file is enabled (to look at log lines in detail)
	LogLevel LogLevel `json:"log_level"`
	// specify a log directory if you want to duplicate all logs into a file by default
	// in your programs provide a way to override this with --log-dir flag
	LogDir string `json:"log_dir"`
	// environment variable name that will allow us to identify program instance
	// for example HOSTNAME can be used inside container to get container id
	// if this variable is set and os.Getenv(ContainerIdVar) is not empty then
	// LodDir = path.Join(LogDir, os.Getenv(ContainerIdVar))
	ContainerIdVarName string `json:"container_id_var_name"` // switch to NONE to not put log files to a separate directory
	// if we want to print goroutine id with each log message
	UseTid *bool `json:"use_tid" default:"true"`
	// time format to use
	TimeFormat string `json:"time_format"`
	// log file format, don't use ':' or '/;'
	LogFileFormat string `json:"log_file_format"`
	// time color (2024/Dec/24 12:39:48)
	LogTimeColor string `json:"log_time_color"`

	// those are parameters that are supplied during logger initialization

	// by default we write to stdout and stderr, but if during logger initialization
	// an option is specified - write all lines regardless of log level to this file
	// this way we can read it later with log reader provided by this package
	LoggerFile     *os.File `json:"logger_file,omitempty"`
	LoggerFilePath string   `json:"logger_file_path,omitempty"`
	FileIsOpen     bool     `json:"log_file_is_open" default:"false"`
}

func DefaultValueConfig() Config {
	useTid := false
	return Config{
		LogLevel:           99, // default Maximum
		ContainerIdVarName: "HOSTNAME",
		UseTid:             &useTid,
		FileIsOpen:         false,
		TimeFormat:         "2006/Jan/02 15:04:05",
		LogFileFormat:      "02_Jan_2006_15_04_05.jsonl",
		LogTimeColor:       DimWhiteColor,
	}
}

// create config with default values before logger gets initialized
var Cfg Config = DefaultValueConfig()
var DefaultCfg Config = DefaultValueConfig() // Default values to replace some values with during logger initialization
