// Read log file and print it to stderr.
// Has an option to choose maximum log level.
// This way we can see messages that we missed during the program run.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"my-project/src/pkg/logger"
)

func main() {
	logFile := flag.String("file", "", "File path to read.")
	logLevel := flag.Int("level", -1, "Log level. Only print messages with log level <= this.")
	startTimeStr := flag.String("start", "0000/Jan/01 00:00:00", "Start time in --time-format format. Keep empty to read from the beginning of the file.")
	endTimeStr := flag.String("end", "9999/Dec/31 23:59:59", "End time in --time-format format. Keep empty to read to the end of the file.")
	timeFormat := flag.String("time-format", logger.Cfg.TimeFormat, "Time format to use for --start and --end. Default is the same as default logger package time format.")
	tail := flag.Int("tail", -1, "Number of lines to show with --tail.")
	flag.Parse()

	if *logFile == "" {
		fmt.Println("Need to specify --file")
		os.Exit(1)
	}
	if *logLevel == -1 {
		fmt.Println("Need to set --level")
		os.Exit(1)
	}

	// Parse the provided start and end times
	startTime, err := time.Parse(*timeFormat, *startTimeStr)
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		return
	}

	endTime, err := time.Parse(*timeFormat, *endTimeStr)
	if err != nil {
		fmt.Println("Error parsing end time:", err)
		return
	}

	// Read file with combined logic
	readErr, errMsg := readLogFile(*logFile, logger.LogLevel(*logLevel), startTime, endTime, *tail)
	logger.QuitIfErrorLoggerIndependent(readErr, errMsg)
}

/*
Read --file.
If --tail is set, collect last N lines.
For each line check if it's between startTime and endTime,
and if its logging level is below or equal to --level.
If conditions are satisfied - print this message using fmt
including all other parts of LogLine.
*/
func readLogFile(logFile string, logLevel logger.LogLevel, startTime, endTime time.Time, tailCount int) (err error, errMsg string) {
	// Open the file
	file, err := os.Open(logFile)
	if err != nil {
		return err, "Unable to open file"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var buffer [][]byte

	// First read all lines
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...) // copy
		buffer = append(buffer, line)
	}
	if err := scanner.Err(); err != nil {
		return err, "Scanner error while reading file"
	}

	// Trim to last N lines if --tail is set
	if tailCount >= 0 && len(buffer) > tailCount {
		buffer = buffer[len(buffer)-tailCount:]
	}

	// Process and print each line
	for _, line := range buffer {
		err, errMsg = processLogLine(line, logLevel, startTime, endTime)
		if err != nil {
			return err, errMsg
		}
	}

	return nil, ""
}

func processLogLine(logLineBytes []byte, logLevel logger.LogLevel, startTime, endTime time.Time) (err error, errMsg string) {
	var logLine logger.LogLine
	// Unmarshal the JSON into the struct
	err = json.Unmarshal(logLineBytes, &logLine)
	if err != nil {
		return err, fmt.Sprintf("Unable to json.Unmarshal line: '%s'", string(logLineBytes))
	}

	// first check time
	if !(AfterOrEqual(logLine.Time, startTime) && BeforeOrEqual(logLine.Time, endTime)) {
		// skip the line if it's not within our time range
		return nil, ""
	}
	// then check log level
	if logLine.Level > logLevel {
		// skip the line if log level is above specified
		return nil, ""
	}

	// now print it
	fmt.Printf(
		"%s [%s][%s] %s",
		fmt.Sprintf(logger.Cfg.LogTimeColor, logLine.Time.Format(logger.Cfg.TimeFormat)),
		fmt.Sprintf(logLine.Color, logLine.Level.String()),
		fmt.Sprintf(logLine.Color, logLine.TId),
		logLine.Msg,
	)

	return nil, ""
}

func AfterOrEqual(t, u time.Time) bool {
	return t.After(u) || t.Equal(u)
}

func BeforeOrEqual(t, u time.Time) bool {
	return t.Before(u) || t.Equal(u)
}
