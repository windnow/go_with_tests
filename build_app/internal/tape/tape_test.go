package tape

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/windnow/edusrv/internal/helpers"
)

func TestTape_Write(t *testing.T) {
	file, clean := CreateTempFile(t, "12345")
	defer clean()

	tape := &Tape{file.(*os.File)}

	tape.Write([]byte("abc"))
	file.Seek(0, 0)
	newFileContent, _ := ioutil.ReadAll(file)

	got := string(newFileContent)
	want := "abc"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
