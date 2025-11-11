package util

import (
	"os"
	"strings"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
)

var RequiredFlags = map[*string]string{}

// RequiredFlag(senderPtr, "--sender")
func RequiredFlag(flagPointer *string, cliName string) {
	RequiredFlags[flagPointer] = cliName
}

// Ensure logs every missing required flag and exits(1) if any were missing.
func EnsureFlags() {
	missing := false
	for flagPointer, cliName := range RequiredFlags {
		if flagPointer == nil || strings.TrimSpace(*flagPointer) == "" {
			tl.Log(tl.Warning, palette.YellowBold, "%s parameter is %s", cliName, "required")
			missing = true
		}
	}
	if missing {
		os.Exit(1)
	}
}
