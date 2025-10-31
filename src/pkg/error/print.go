// everything related to printing errors to the terminal
// sometimes stop the program too
package er

import (
	"os"
	"runtime/debug"

	"work-tracker/src/pkg/logger"
)

// error (red), warning (yellow) or skip (purple)
type ErrorType string

const (
	ErrorTypeError   ErrorType = "error"
	ErrorTypeWarning ErrorType = "warning"
	ErrorTypeSkip    ErrorType = "skip"
)

/*
Check if error is nil. If not - print it with either red, yellow or purple color.

If stopCode parameter is not 0 - stop the program after.
*/
func (e *Error) PrintErrorWithOptions(
	logLevel logger.LogLevel, errorType ErrorType, stopCode int,
	printContext, printDebugStack bool,
) {
	if e == nil {
		return
	}

	var colors map[string]string
	switch errorType {
	case "error":
		colors = map[string]string{
			"msg":     logger.BoldRedBackgroundColor,
			"err":     logger.BoldBrightRedColor,
			"where":   logger.BoldRedColor,
			"debug":   logger.DimBrightRedColor,
			"context": logger.DimRedColor,
		}
	case "warning":
		colors = map[string]string{
			"msg":     logger.BoldYellowBackgroundColor,
			"err":     logger.BoldBrightYellowColor,
			"where":   logger.BoldYellowColor,
			"debug":   logger.DimBrightYellowColor,
			"context": logger.DimYellowColor,
		}
	case "skip":
		colors = map[string]string{
			"msg":     logger.BoldPurpleBackgroundColor,
			"err":     logger.BoldBrightPurpleColor,
			"where":   logger.BoldPurpleColor,
			"debug":   logger.DimBrightPurpleColor,
			"context": logger.DimPurpleColor,
		}
	default:
		colors = map[string]string{
			"msg":     logger.BoldRedBackgroundColor,
			"err":     logger.BoldRedColor,
			"where":   logger.BrightRedColor,
			"debug":   logger.DimBrightRedColor,
			"context": logger.DimRedColor,
		}
	}

	logger.Log(logLevel, colors["msg"], "Msg: '%s'", e.Msg)
	logger.Log(logLevel+1, colors["err"], "Err: '%s'", e.ErrStr)
	logger.Log(logLevel+2, colors["where"], "Where: '%s'", e.Where)
	if printDebugStack {
		logger.Log(logLevel+3, colors["debug"], "Debug stack:\n```\n%s\n```", string(debug.Stack()))
	}
	if printContext {
		logger.Log(logLevel+4, colors["context"], "Context:\n```\n%s\n```", e.Context)
	}

	if stopCode > 0 {
		os.Exit(stopCode)
	}
}

// there we keep all the wrappers around PrintErrorWithOptions
// for everything else - use PrintErrorWithOptions directly

// print error without context
func (e *Error) Print(errorType ErrorType, logLevel logger.LogLevel, stopCode int) {
	e.PrintErrorWithOptions(logLevel, errorType, stopCode, false, false)
}

// print error with context
func (e *Error) PrintWithContext(errorType ErrorType, logLevel logger.LogLevel, stopCode int) {
	e.PrintErrorWithOptions(logLevel, errorType, stopCode, true, true)
}

// same as PrintWithContext but with less parameters and more clear name
func (e *Error) QuitIf(errorType ErrorType) {
	e.PrintErrorWithOptions(logger.Critical1, errorType, 1, true, true)
}
