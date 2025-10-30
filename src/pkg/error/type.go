// our error type
package er

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"

	"gorm.io/gorm"
)

type Error struct {
	Err     error  `json:"-" gorm:"-"`            // when converting to JSON or saving to the database make sure to use .Err.Error() to populate ErrStr
	Msg     string `json:"msg"`                   // additional error message, explaining error
	Where   string `json:"where"`                 // place where error was created with NewError /abs/file/path:<line-number>
	ErrStr  string `json:"err" gorm:"column:err"` // this is the part that is sent via json and goes into the databse
	Context string `json:"context,omitempty"`     // large variable contents
}

// NewError creates a new *Error from an error, message, and arbitrary context.
// The context value is stringified intelligently based on its type.
func NewError(err error, msg string, context any) (e *Error) {
	var contextString string
	if context != nil {
		contextString = StringifyContext(context)
	}

	e = &Error{
		Err:     err,
		Msg:     msg,
		Where:   getCallerLocation(2),
		Context: contextString, // user-provided context
	}
	if err != nil {
		e.ErrStr = err.Error() // populated automatically by BeforeSave / MarshalJSON, but set it anyway here
	}

	return e
}

/*
New Error - Explain Context

A wrapper around NewError.
Allows us to avoid explainig context with fmt.Sprintf("<context explaination> '%s'", contextVar).

Instead just specify contextLabel parameter.
*/
func NewErrorEC(err error, msg, contextLabel string, context any, multilineContext bool) (e *Error) {
	var contextString string
	if context != nil {
		contextString = StringifyContext(context)
	}
	if multilineContext {
		contextString = fmt.Sprintf("%s:\n'''\n%s\n'''", contextLabel, contextString)
	} else {
		contextString = fmt.Sprintf("%s: '%s'", contextLabel, contextString)
	}

	return NewError(err, msg, contextString)
}

// A wrapper around NewErrorEC, single-line context.
func NewErrorECOL(err error, msg, contextLabel string, context any) (e *Error) {
	return NewErrorEC(err, msg, contextLabel, context, false)
}

// A wrapper around NewErrorEC, muilti-line context.
func NewErrorECML(err error, msg, contextLabel string, context any) (e *Error) {
	return NewErrorEC(err, msg, contextLabel, context, true)
}

// BeforeSave ensures ErrStr is kept in sync with Err before writing to DB.
func (e *Error) BeforeSave(tx *gorm.DB) (err error) {
	if e.Err != nil {
		e.ErrStr = e.Err.Error()
	} else {
		e.ErrStr = ""
	}
	return nil
}

// AfterFind restores Err from ErrStr after reading from DB (generic error only).
func (e *Error) AfterFind(tx *gorm.DB) (err error) {
	if e.ErrStr != "" {
		e.Err = errors.New(e.ErrStr) // note: original error type is lost
	}
	return nil
}

// MarshalJSON ensures ErrStr is always derived from Err when encoding to JSON.
func (e Error) MarshalJSON() ([]byte, error) {
	// json.Marshal would detect that Error implements MarshalJSON and call it again,
	// and again, forever, until a stack overflow.
	// so we use alias type to prevent that, Alias does not have it's methods, but has it's fields.
	type Alias Error
	if e.Err != nil {
		e.ErrStr = e.Err.Error()
	} else {
		e.ErrStr = ""
	}
	return json.Marshal((Alias)(e))
}

/*
Get the cal ler info.
skip controlls how many levels up we want to go
*/
func getCallerLocation(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	where := "UNKNOWN"
	if ok {
		where = fmt.Sprintf("%s:%d", file, line)
	}

	return where
}
