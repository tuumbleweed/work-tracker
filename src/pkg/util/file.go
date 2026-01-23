package util

import (
	"os"
	"path/filepath"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

/*
EnsureDirExists ensures dirPath exists and is a directory.

Behavior:
- Cleans dirPath (removes "./", resolves "..").
- If dirPath already exists and is a directory: returns nil.
- If dirPath does not exist: creates it and any missing parents (mkdir -p).
- If dirPath exists but is not a directory: returns *xerr.Error.
- If the OS refuses access (permissions, etc.): returns *xerr.Error.

Notes:
- This function does not log on success to avoid spam.
- It logs intent at Debug to help trace filesystem setup when debugging.
*/
func EnsureDirExists(dirPath string, perm os.FileMode) (e *xerr.Error) {
	dirPath = filepath.Clean(dirPath)

	// Intent at Debug (dim blue).
	tl.Log(tl.Debug, palette.BlueDim, "Ensure dir exists: '%s'", dirPath)

	info, err := os.Stat(dirPath)
	if err == nil {
		if !info.IsDir() {
			err = &os.PathError{Op: "stat", Path: dirPath, Err: os.ErrInvalid}
			e = xerr.NewErrorECOL(err, "path exists but is not a directory", "dir", dirPath)
			return e
		}
		return nil
	}

	if os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, perm)
		if err != nil {
			e = xerr.NewErrorECOL(err, "create directory", "dir", dirPath)
			return e
		}
		return nil
	}

	e = xerr.NewErrorECOL(err, "stat directory", "dir", dirPath)
	return e
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}
