package then

import (
	"os"
	"testing"
)

func WithStdIn(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal("error creating pipe:", err)
	}

	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatal("failure to close reader:", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal("failure to close writer:", err)
		}
	})

	return reader, writer
}
