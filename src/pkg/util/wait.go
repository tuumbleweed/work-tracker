package util

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"my-project/src/pkg/logger"
)

// Wait for timeout amount of time
func WaitFor(timeout time.Duration) {
	logger.Log(logger.Debug, logger.DimCyanColor, "%s for %v", "Waiting", timeout.Round(time.Millisecond))
	time.Sleep(timeout)
}

// Calculate a time.Duration between min and max, use WaitFor function with this value
func WaitBetween(min, max time.Duration) {
	randomDurationNanoseconds := rand.Int63n(max.Nanoseconds()+1-min.Nanoseconds()) + min.Nanoseconds()
	WaitFor(time.Duration(randomDurationNanoseconds))
}

// Wait for set amount of seconds
func WaitForSeconds(timeoutSeconds float64) {
	logger.Log(logger.Debug, logger.DimCyanColor, "%s for %s seconds", "Waiting", fmt.Sprintf("%.3f", timeoutSeconds))
	timeoutNanoseconds := int(timeoutSeconds * math.Pow(10, 9))
	time.Sleep(time.Duration(timeoutNanoseconds))
}

// Calculate a time.Duration between min and max, use WaitForSeconds function with this value
func WaitBetweenSeconds(min, max float64) {
	timeout := min + rand.Float64()*(max-min)
	WaitForSeconds(timeout)
}
