package then

import (
	"os"
	"testing"
)

func WithReadWritePipe(t *testing.T) (*os.File, *os.File) {
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

func WithStdInOut(t *testing.T) (*os.File, *os.File, *os.File, *os.File) {
	inReader, inWriter := WithReadWritePipe(t)
	outReader, outWriter := WithReadWritePipe(t)

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	os.Stdin = inReader
	os.Stdout = outWriter

	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	})

	return inReader, inWriter, outReader, outWriter
}
