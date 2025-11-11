// in case you need to create an entrypoint with multiple subprograms
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"

	"work-tracker/src/pkg/config"
	"work-tracker/src/pkg/email"
	"work-tracker/src/pkg/report"
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
	configPath := subprogramCmd.String("config", "./cfg/config.json", "Log level. Default is LOG_LEVEL env var value")

	// custom flags
	provider := subprogramCmd.String("provider", "mailgun", "Provider to use when sending emails")
	senderAddress := subprogramCmd.String("sender", "", "Sender's address")
	recipientAddress := subprogramCmd.String("recipient", "", "Recipient's address")
	subject := subprogramCmd.String("subject", "Test subject", "Subject of an email")
	emailHtmlFilePath := subprogramCmd.String("html-file", "./tmp/email.html", "Html of an email")
	emailTxtFilePath := subprogramCmd.String("plain-file", "./tmp/email.txt", "Plain text of an email")

	// parse and init config
	xerr.QuitIfError(subprogramCmd.Parse(flags), "Unable to subprogramCmd.Parse")
	config.InitializeConfig(*configPath)

	util.RequiredFlag(senderAddress, "--sender")
	util.RequiredFlag(recipientAddress, "--recipient")
	util.EnsureFlags()

	recipientAddresses := strings.Split(*recipientAddress, ",")

	// read html file
	htmlFileContentsBytes, err := os.ReadFile(*emailHtmlFilePath)
	xerr.QuitIfError(err, fmt.Sprintf("Unable to read file '%s'", *emailHtmlFilePath))
	htmlFileContents := string(htmlFileContentsBytes)
	tl.Log(tl.Verbose, palette.BlueDim, "Full Email:\n```\n%s\n```", htmlFileContents)
	// read text file
	textFileContentsBytes, err := os.ReadFile(*emailTxtFilePath)
	xerr.QuitIfError(err, fmt.Sprintf("Unable to read file '%s'", *emailTxtFilePath))
	textFileContents := string(textFileContentsBytes)
	tl.Log(tl.Verbose, palette.BlueDim, "Full Email:\n```\n%s\n```", textFileContents)

	// send email here
	sendEmails := true
	e := email.SendMessage(email.Provider(*provider), &sendEmails, *senderAddress, recipientAddresses, *subject, textFileContents, htmlFileContents, nil)
	e.QuitIf("error")
}

/*
A shorter version of testProvidxerr.
Exists to send a report from ./out/report.html.

Currently no text version of the report.
*/
func sendReport(subprogram string, flags []string) {
	// only one provider is needed here, you can set others to dummy values
	util.CheckIfEnvVarsPresent([]string{
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", // amazon ses
		"MAILGUN_DOMAIN", "MAILGUN_API_KEY", // mailgun
		"SENDGRID_API_KEY", // sendgrid
	})

	// common flags
	subprogramCmd := flag.NewFlagSet(subprogram, flag.ExitOnError)
	configPath := subprogramCmd.String("config", "./cfg/config.json", "Log level. Default is LOG_LEVEL env var value")

	// custom flags
	provider := subprogramCmd.String("provider", "mailgun", "Provider to use when sending emails")
	senderAddress := subprogramCmd.String("sender", "", "Sender's address")
	recipientAddress := subprogramCmd.String("recipient", "", "Recipient's address")
	emailHtmlFilePath := subprogramCmd.String("html-file", "./out/report.html", "Html of an email")

	// parse and init config
	xerr.QuitIfError(subprogramCmd.Parse(flags), "Unable to subprogramCmd.Parse")
	config.InitializeConfig(*configPath)

	util.RequiredFlag(senderAddress, "--sender")
	util.RequiredFlag(recipientAddress, "--recipient")
	util.EnsureFlags()

	recipientAddresses := strings.Split(*recipientAddress, ",")

	// read html file
	htmlFileContentsBytes, err := os.ReadFile(*emailHtmlFilePath)
	xerr.QuitIfError(err, fmt.Sprintf("Unable to read file '%s'", *emailHtmlFilePath))
	htmlFileContents := string(htmlFileContentsBytes)
	tl.Log(tl.Verbose, palette.BlueDim, "Full Email:\n```\n%s\n```", htmlFileContents)
	// read text file

	reportTitle, e := report.ReadHTMLTitleFromBytes(htmlFileContentsBytes)
	e.QuitIf("error")
	subject := fmt.Sprintf(
		"Work Tracker\u00A0\u00A0\u00A0·\u00A0\u00A0\u00A0%s\u00A0\u00A0\u00A0·\u00A0\u00A0\u00A0%s",
		reportTitle, time.Now().Format("2006-01-02 (Mon) 15:04:05"),
	)

	// send email here
	sendEmails := true
	e = email.SendMessage(email.Provider(*provider), &sendEmails, *senderAddress, recipientAddresses, subject, "", htmlFileContents, nil)
	e.QuitIf("error")
}

func main() {
	// Check if there are enough arguments
	if len(os.Args) < 2 {
		tl.Log(tl.Error, palette.Red, "Usage: %s", "go run src/cmd/unsubscriber/main.go subprogram_name(for exampe unsubscriber)")
		os.Exit(1)
	}
	subprogram := os.Args[1]
	flags := os.Args[2:]

	// Switch subprogram based on the first argument
	switch subprogram {
	case "test-provider":
		testProvider(subprogram, flags)
	case "report":
		sendReport(subprogram, flags)
	default:
		tl.Log(tl.Error, palette.Red, "Unknown subprogram: %s", subprogram)
		os.Exit(1)
	}
}
