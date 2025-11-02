// in case you need to create an entrypoint with multiple subprograms
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"work-tracker/src/pkg/config"
	"work-tracker/src/pkg/email"
	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
	"work-tracker/src/pkg/util"
)

/*
Pick prvider and use it to send a test email to admin/specified address.
Specify test email file path (generate it with substitute-variables subprogram)
*/
func testProvider(subprogram string, flags []string) {
	util.CheckIfEnvVarsPresent([]string{
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", // amazon ses
		"MAILGUN_DOMAIN", "MAILGUN_API_KEY", // mailgun
		"SENDGRID_API_KEY", // sendgrid
	})

	// common flags
	subprogramCmd := flag.NewFlagSet(subprogram, flag.ExitOnError)
	logLevelOverride := subprogramCmd.Int("log-level", -1, "Log level. Default is whatever value is in configuration file. Keep at -1 to not override.")
	logDirOverride := subprogramCmd.String("log-dir", "", "File directory at which to save log files. Keep empty to use configuration file instead.")
	configPath := subprogramCmd.String("config", "./cfg/config.json", "Log level. Default is LOG_LEVEL env var value")

	// custom flags
	provider := subprogramCmd.String("provider", "mailgun", "Provider to use when sending emails")
	senderAddress := subprogramCmd.String("sender", "", "Sender's address")
	recipientAddress := subprogramCmd.String("recipient", "", "Recipient's address")
	subject := subprogramCmd.String("subject", "Test subject", "Subject of an email")
	emailHtmlFilePath := subprogramCmd.String("html-file", "./tmp/email.html", "Html of an email")
	emailTxtFilePath := subprogramCmd.String("plain-file", "./tmp/email.txt", "Plain text of an email")

	// parse and init config
	er.QuitIfError(subprogramCmd.Parse(flags), "Unable to subprogramCmd.Parse")
	config.InitializeConfig(*configPath, logger.LogLevel(*logLevelOverride), *logDirOverride)

	util.RequiredFlag(senderAddress, "--sender")
	util.RequiredFlag(recipientAddress, "--recipient")
	util.EnsureFlags()

	recipientAddresses := strings.Split(*recipientAddress, ",")

	// read html file
	htmlFileContentsBytes, err := os.ReadFile(*emailHtmlFilePath)
	er.QuitIfError(err, fmt.Sprintf("Unable to read file '%s'", *emailHtmlFilePath))
	htmlFileContents := string(htmlFileContentsBytes)
	logger.Log(logger.Verbose, logger.DimBlueColor, "Full Email:\n```\n%s\n```", htmlFileContents)
	// read text file
	textFileContentsBytes, err := os.ReadFile(*emailTxtFilePath)
	er.QuitIfError(err, fmt.Sprintf("Unable to read file '%s'", *emailTxtFilePath))
	textFileContents := string(textFileContentsBytes)
	logger.Log(logger.Verbose, logger.DimBlueColor, "Full Email:\n```\n%s\n```", textFileContents)

	// send email here
	sendEmails := true
	e := email.SendMessage(email.Provider(*provider), &sendEmails, *senderAddress, recipientAddresses, *subject, textFileContents, htmlFileContents, nil)
	e.QuitIf("error")
}

func main() {
	// Check if there are enough arguments
	if len(os.Args) < 2 {
		logger.Log(logger.Error, logger.RedColor, "Usage: %s", "go run src/cmd/unsubscriber/main.go subprogram_name(for exampe unsubscriber)")
		os.Exit(1)
	}
	subprogram := os.Args[1]
	flags := os.Args[2:]

	// Switch subprogram based on the first argument
	switch subprogram {
	case "test-provider":
		testProvider(subprogram, flags)
	default:
		logger.Log(logger.Error, logger.RedColor, "Unknown subprogram: %s", subprogram)
		os.Exit(1)
	}
}
