package wfl

import (
	"os"
)

func RandomFileNameInTempDir() string {
	f, err := os.CreateTemp("", "wfl")
	if err != nil {
		return os.TempDir() + "/wfl.tmp"
	}
	defer f.Close()
	return f.Name()
}
