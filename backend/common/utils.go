package common

import (
	"os"
	"path"
)

func GetTempDir() string {
	return path.Join(os.TempDir(), "avtmp")
}
