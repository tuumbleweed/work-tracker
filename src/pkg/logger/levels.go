package logger

import "fmt"

// LogLevel is a custom type for log levels
type LogLevel int

// Defining the log levels
const (
	// 0-9 - Critical. Program execution ends here. Needs to be fixed.
	Critical LogLevel = iota
	Critical1
	Critical2
	Critical3
	Critical4
	Critical5
	Critical6
	Critical7
	Critical8
	Critical9

	// 10-19 - Error. Something went wrong but we continue program execution.
	// Usually needs to be fixed.
	Error
	Error1
	Error2
	Error3
	Error4
	Error5
	Error6
	Error7
	Error8
	Error9

	// 20-29 - Warning. For exampe: unable to make http requrest due to page not responding, program continues.
	// Usually doesn't need to be fixed.
	Warning
	Warning1
	Warning2
	Warning3
	Warning4
	Warning5
	Warning6
	Warning7
	Warning8
	Warning9

	// 30-39 - Important. Ususally once per run.
	Important
	Important1
	Important2
	Important3
	Important4
	Important5
	Important6
	Important7
	Important8
	Important9

	// 40-49 - Notice. Ususally once per batch.
	Notice
	Notice1
	Notice2
	Notice3
	Notice4
	Notice5
	Notice6
	Notice7
	Notice8
	Notice9

	// 50-59 - Info. Ususally once per record.
	Info
	Info1
	Info2
	Info3
	Info4
	Info5
	Info6
	Info7
	Info8
	Info9

	// 60-69 - Detailed. Ususally a few times per record.
	Detailed
	Detailed1
	Detailed2
	Detailed3
	Detailed4
	Detailed5
	Detailed6
	Detailed7
	Detailed8
	Detailed9

	// 70-79 - Verbose. Ususally many times per record.
	Verbose
	Verbose1
	Verbose2
	Verbose3
	Verbose4
	Verbose5
	Verbose6
	Verbose7
	Verbose8
	Verbose9

	// 80-89 - Debug. Extremely verbose.
	Debug
	Debug1
	Debug2
	Debug3
	Debug4
	Debug5
	Debug6
	Debug7
	Debug8
	Debug9

	// if override value is this (it's default value) - then don't override logging level.
	DontOverride LogLevel = -1
)

// String method to get the string representation of each log level
func (logLevel LogLevel) String() string {
	logLevels := [...]string{
		"Critical", "Critical1", "Critical2", "Critical3", "Critical4", "Critical5", "Critical6", "Critical7", "Critical8", "Critical9",
		"Error", "Error1", "Error2", "Error3", "Error4", "Error5", "Error6", "Error7", "Error8", "Error9",
		"Warning", "Warning1", "Warning2", "Warning3", "Warning4", "Warning5", "Warning6", "Warning7", "Warning8", "Warning9",
		"Important", "Important1", "Important2", "Important3", "Important4", "Important5", "Important6", "Important7", "Important8", "Important9",
		"Notice", "Notice1", "Notice2", "Notice3", "Notice4", "Notice5", "Notice6", "Notice7", "Notice8", "Notice9",
		"Info", "Info1", "Info2", "Info3", "Info4", "Info5", "Info6", "Info7", "Info8", "Info9",
		"Detailed", "Detailed1", "Detailed2", "Detailed3", "Detailed4", "Detailed5", "Detailed6", "Detailed7", "Detailed8", "Detailed9",
		"Verbose", "Verbose1", "Verbose2", "Verbose3", "Verbose4", "Verbose5", "Verbose6", "Verbose7", "Verbose8", "Verbose9",
		"Debug", "Debug1", "Debug2", "Debug3", "Debug4", "Debug5", "Debug6", "Debug7", "Debug8", "Debug9",
	}

	if int(logLevel) < 0 || int(logLevel) >= len(logLevels) {
		return fmt.Sprintf("Unknown%d", logLevel)
	}

	return logLevels[logLevel]
}
