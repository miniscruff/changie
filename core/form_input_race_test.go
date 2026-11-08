//go:build race

package core

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/miniscruff/changie/then"
)

func TestFormInput(t *testing.T) {
	t.Run("pipe is read as a plain reader", func(t *testing.T) {
		reader, writer, err := os.Pipe()
		then.Nil(t, err)

		t.Cleanup(func() {
			_ = reader.Close()
			_ = writer.Close()
		})

		_, isFile := formInput(reader).(*os.File)
		then.False(t, isFile)
	})

	t.Run("other readers are unchanged", func(t *testing.T) {
		reader := strings.NewReader("input")
		then.Equals[io.Reader](t, reader, formInput(reader))
	})
}
