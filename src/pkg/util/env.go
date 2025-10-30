package util

import (
	"fmt"
	"os"

	"my-project/src/pkg/logger"
	er "my-project/src/pkg/error"
)

// Check if all required environment variables are present. If not all present - print warinig and quit.
func CheckIfEnvVarsPresent(listOfEnvVars []string) {
	for _, envVarName := range listOfEnvVars {
		if os.Getenv(envVarName) == "" {
			logger.Log(logger.Warning, logger.YellowColor, "Env var. '%s' is not set. %s", envVarName, "Check your environment variables")
			os.Exit(1)
		}
	}
}

/*
envOrDefault returns the value of an environment variable or the provided fallback when empty.
*/
func EnvOrDefault(name, fallback string) string {
	val := os.Getenv(name)
	if val == "" {
		return fallback
	}
	return val
}

// Set environment variable
func SetEnvVar(name, value string) {
	err := os.Setenv(name, value)
	if err != nil {
		e := er.NewError(err, "Unable to set environment variable", fmt.Sprintf("Env var: '%s', value: '%s'", name, value))
		e.QuitIf("error")
	}
}
