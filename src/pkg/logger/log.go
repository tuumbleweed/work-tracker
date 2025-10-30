package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

/*
A wrapper around LogBool to avoid passing those boolean values each time.
We usually need to print time, new line and usually don't need slower regex formatting.
When you need to not print time and/or use regex just use LogBool instead.
*/
func Log(msgLogLevel LogLevel, color, format string, args ...any) (colorlessMsg string) {
	return LogBool(true, false, true, false, false, msgLogLevel, color, format, args...)
}

func LogRewriteLine(msgLogLevel LogLevel, color, format string, args ...any) (colorlessMsg string) {
	return LogBool(true, false, false, false, true, msgLogLevel, color, format, args...)
}

func LogNewLineAtTheStart(msgLogLevel LogLevel, color, format string, args ...any) (colorlessMsg string) {
	return LogBool(true, true, true, false, false, msgLogLevel, color, format, args...)
}

/*
Print message with white color, and arguments within that message with 'color' color.
Olny %s can be used to insert arguments. (else you gonna get messages like '%!d(string=5353)')
Return colorless msg.

New lines should be manually added. Output is printed to stderr.

printTime, useRegex need to be controlled each time we print message hence we pass them as function parameters.
*/
func LogBool(printTime, printNewlineAtTheStart, printNewlineAtTheEnd, useRegex, rewriteString bool, msgLogLevel LogLevel, color, format string, args ...any) (colorlessMsg string) {
	if printNewlineAtTheEnd {
		format += "\n"
	}

	// quit early if you don't need to both print and save file at the same time
	if (Cfg.LogLevel < msgLogLevel) && !Cfg.FileIsOpen {
		// return colorlessMsg
		return fmt.Sprintf(format, args...)
	}

	colorlessMsg = fmt.Sprintf(format, args...)
	var msgWithColors string
	if useRegex {
		msgWithColors = getFormattedStringWithRegex(color, format, args...)
	} else {
		msgWithColors = getFormattedStringWithArgIteration(color, format, args...)
	}
	var tid int
	var logLevelAndTidString string
	if *Cfg.UseTid { // it's not nil because we initialized the config
		tid = getTid()
		logLevelAndTidString = fmt.Sprintf("[%s][%s] ", fmt.Sprintf(color, msgLogLevel.String()), fmt.Sprintf(color, tid))
	} else {
		logLevelAndTidString = fmt.Sprintf("[%s] ", fmt.Sprintf(color, msgLogLevel.String()))
	}

	// we only print to stderr if Cfg.LogLevel allows for this message to pass
	// for example by setting log level to Info5 (int value 55) every message
	// with log level value bigger than 55, for example Verbose3 (int value 73)
	// will not be printed to stderr
	if Cfg.LogLevel >= msgLogLevel {
		if printTime {
			timeString := Cfg.LogTimeColor
			if rewriteString {
				timeString = "\r" + timeString
			}
			if printNewlineAtTheStart {
				timeString = "\n" + timeString
			}
			fmt.Fprint(os.Stderr, fmt.Sprintf(timeString, time.Now().Format(Cfg.TimeFormat))+" ", logLevelAndTidString, msgWithColors)
		} else {
			if rewriteString {
				logLevelAndTidString = "\r" + logLevelAndTidString
			}
			if printNewlineAtTheStart {
				logLevelAndTidString = "\n" + logLevelAndTidString
			}
			fmt.Fprint(os.Stderr, logLevelAndTidString, msgWithColors)
		}
	}

	if Cfg.FileIsOpen {
		QuitIfErrorLoggerIndependent(WriteLogLine(tid, msgLogLevel, color, msgWithColors))
	}

	return colorlessMsg
}

/*
A wrapper around LogMonoColorBool to avoid passing those boolean values each time.
We usually need to print time, new line and usually don't need slower regex formatting.
When you need to not print time and/or use regex just use LogMonoColorBool instead.
*/
func LogMonoColor(msgLogLevel LogLevel, color, format string, args ...any) (colorlessMsg string) {
	return LogMonoColorBool(true, true, msgLogLevel, color, format, args...)
}

// Print message with a single color
func LogMonoColorBool(printTime, printNewline bool, msgLogLevel LogLevel, color, format string, args ...any) (colorlessMsg string) {
	if printNewline {
		format += "\n"
	}

	// quit early if you don't need to both print and save file at the same time
	if (Cfg.LogLevel < msgLogLevel) && !Cfg.FileIsOpen {
		// return colorlessMsg
		return fmt.Sprintf(format, args...)
	}

	colorlessMsg = fmt.Sprintf(format, args...)
	monoColorMsg := fmt.Sprintf(color, colorlessMsg)
	// logLevelAndTidString := fmt.Sprintf("[%s][%s] ", fmt.Sprintf(color, msgLogLevel.String()), fmt.Sprintf(color, tid))
	var tid int
	var logLevelAndTidString string
	if *Cfg.UseTid { // it's not nil because we initialized the config
		tid = getTid()
		logLevelAndTidString = fmt.Sprintf("[%s][%s] ", fmt.Sprintf(color, msgLogLevel.String()), fmt.Sprintf(color, tid))
	} else {
		logLevelAndTidString = fmt.Sprintf("[%s] ", fmt.Sprintf(color, msgLogLevel.String()))
	}

	// we only print to stderr if Cfg.LogLevel allows for this message to pass
	// for example by setting log level to Info5 (int value 55) every message
	// with log level value bigger than 55, for example Verbose3 (int value 73)
	// will not be printed to stderr
	if Cfg.LogLevel >= msgLogLevel {
		if printTime {
			fmt.Fprint(os.Stderr, fmt.Sprintf(Cfg.LogTimeColor, time.Now().Format(Cfg.TimeFormat))+" ", logLevelAndTidString, monoColorMsg)
		} else {
			fmt.Fprint(os.Stderr, logLevelAndTidString, monoColorMsg)
		}
	}

	if Cfg.FileIsOpen {
		QuitIfErrorLoggerIndependent(WriteLogLine(tid, msgLogLevel, color, monoColorMsg))
	}

	return colorlessMsg
}

// write a line to Cfg.LoggerFile
func WriteLogLine(tid int, msgLogLevel LogLevel, color, msg string) (err error, errMsg string) {
	logLine := LogLine{
		Time:  time.Now(),
		TId:   tid,
		Level: msgLogLevel,
		Color: color,
		Msg:   msg,
	}
	bytes, err := json.Marshal(logLine)
	if err != nil {
		return err, fmt.Sprintf("Unable to marshal out: %v", logLine)
	}
	_, err = Cfg.LoggerFile.WriteString(string(bytes) + "\n")
	if err != nil {
		return err, fmt.Sprintf("Unable to write into file: '%s'", Cfg.LoggerFilePath)
	}

	return nil, ""
}
