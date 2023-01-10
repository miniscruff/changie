package then

import (
	"io"
	"testing"
	"time"
)

const delayTimeMS = 50

func DelayWrite(t *testing.T, writer io.Writer, data ...[]byte) {
	t.Helper()

	go func() {
		for _, bs := range data {
			time.Sleep(delayTimeMS * time.Millisecond)

			_, err := writer.Write(bs)
			Nil(t, err)
		}
	}()
}

// ErrWriter is a simple struct that will return an error when trying to Write
type ErrWriter struct {
	err error
}

func NewErrWriter(err error) *ErrWriter {
	return &ErrWriter{err: err}
}

func (w *ErrWriter) Write(data []byte) (int, error) {
	return 0, w.err
}
