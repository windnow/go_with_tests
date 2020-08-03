package helpers

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func CreateTempFile(t *testing.T, initialData string) (io.ReadWriteSeeker, func()) {
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
