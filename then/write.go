package then

import (
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

const delayTimeMS = 50

// DelayWrite will wait a few milliseconds between writing
// some data in a separate goroutine, this is used when
// we are prompting the user for input and need to write responses.
func DelayWrite(t *testing.T, writer io.Writer, data ...[]byte) {
	t.Helper()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	t.Cleanup(wg.Wait)

	go func() {
		defer wg.Done()

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

func NewErrWriter() *ErrWriter {
	return &ErrWriter{
		err: errors.New("error from ErrWriter"),
	}
}

func (w *ErrWriter) Write(data []byte) (int, error) {
	return 0, w.err
}

// Raised will assert the error value is the one we would of returned if
// Write was called.
func (w *ErrWriter) Raised(t *testing.T, err error) {
	t.Helper()
	Err(t, w.err, err)
}
