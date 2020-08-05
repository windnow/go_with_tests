package helpers

import (
	"io/ioutil"
	"os"
	"testing"
)

// CreateTempFile ...
func CreateTempFile(t *testing.T, initialData string) (*os.File, func()) {
	t.Helper()

	tmpfile, err := ioutil.TempFile("", "db")

	if err != nil {
		t.Fatalf("could't create tmp file %v", err)
	}

	tmpfile.Write([]byte(initialData))

	return tmpfile, func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}
}
