package logger

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type LogLine struct {
	Time  time.Time `json:"t"`
	TId   int       `json:"tid,omitempty"`
	Level LogLevel  `json:"l"`
	Color string    `json:"c"`
	Msg   string    `json:"msg"`
}

func replaceVerbs(format, color string) string {
	// Regex to find all verbs (format specifiers)
	re := regexp.MustCompile(`%[0-9 #\-%.]{0,4}[a-zA-Z]{1,1}`)

	// Replace all matches with the custom format
	return re.ReplaceAllStringFunc(format, func(match string) string {
		// meaning "\033[1;33m%v\033[0m" %v is replaced by custom say %d
		return fmt.Sprintf(color, match)
	})
}

// alternative version of log using regex, works with verbs such as '% 9d' or '% -9d' but is slower
func getFormattedStringWithRegex(color, format string, args ...any) (msg string) {
	format = replaceVerbs(format, color)
	msg = fmt.Sprintf(format, args...)

	return msg
}

// default version, can't handle verbs like '% 9d' or '% -9d' but is faster
// can handle "% 6s | %-40s | %40s" or % 6v | %-40v | %40v" though
func getFormattedStringWithArgIteration(color, format string, args ...any) (msg string) {
	var colorfulStringArgs []any
	for _, arg := range args {
		colorfulStringArgs = append(colorfulStringArgs, fmt.Sprintf(color, arg))
	}
	msg = fmt.Sprintf(format, colorfulStringArgs...)

	return msg
}

// get goroutine id
// gotta have this, it will slow things down but we need to track
// in which goroutine thing happens
func getTid() (tid int) {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	tid, err := strconv.Atoi(idField)
	if err != nil {
		log.Printf("Cannot get goroutine id, Err: %s", err.Error())
		return -1
	}
	return tid
}
