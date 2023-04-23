package wfl

import (
	"os"
)

// RandomFileNameInTempDir returns a random file name in the
// temporary directory. The file name contains an replacement template
// {{.ID}} which gets replaced by a different number for each task.
// The file is not created.
func RandomFileNameInTempDir() string {
	f, err := os.CreateTemp("", "wfl")
	if err != nil {
		// {{.ID}} is replaced by an ID generated for each
		// task so that the file name is unique for each task.
		return os.TempDir() + "/wfl-{{.ID}}.out"
	}
	defer f.Close()
	return f.Name() + "-{{.ID}}.out"
}
