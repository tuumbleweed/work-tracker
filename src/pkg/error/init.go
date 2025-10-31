// Everything that works with raw err error
// allows us to wrap NewError calls in one line
package er

import "work-tracker/src/pkg/logger"

/*
If err is nil - do nothing.
If it's not nil - use NewError and then PrintError.

Without context.
*/
func QuitIfError(err error, msg string) {
	if err == nil {
		return
	}
	NewError(err, msg, "").Print("error", logger.Critical, 1)
}

/*
If err is nil - do nothing.
If it's not nil - use NewError and then PrintError.

Without context.
*/
func QuitIfErrorWithContext(err error, msg, context string) {
	if err == nil {
		return
	}
	NewError(err, msg, context).PrintWithContext("error", logger.Critical, 1)
}

// same as QuitIfError but errorType is a parameter
func QuitIf(errorType ErrorType, err error, msg string) {
	if err == nil {
		return
	}
	NewError(err, msg, "").Print(errorType, logger.Critical, 1)
}

// same as QuitIf but don't stop the program
func Print(errorType ErrorType, err error, msg string) {
	if err == nil {
		return
	}
	NewError(err, msg, "").Print(errorType, logger.Critical, 0)
}
