package util

import (
	"os"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

func CreateDirIfDoesntExist(path string) (e *xerr.Error) {
	if path == "" {
		tl.Log(tl.Detailed, palette.BlueDim, "Path '%s' is empty, %s", path, "not creating")
		return nil
	}
	tl.Log(tl.Detailed, palette.BlueDim, "Creating '%s' dir", path)
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return xerr.NewErrorECOL(err, "Unable to create a directory", "directory path", path)
	}

	return nil
}
