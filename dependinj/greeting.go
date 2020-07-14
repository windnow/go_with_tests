package dependinj

import (
	"fmt"
	"io"
)

// Greet send greeting
func Greet(writer io.Writer, name string) {
	fmt.Fprintf(writer, "Hello, %s", name)
}
