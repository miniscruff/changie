package testutils

import (
	"io"
	"time"

	"github.com/onsi/gomega"
)

const delayTimeMS = 50

func DelayWrite(writer io.Writer, data []byte) {
	time.Sleep(delayTimeMS * time.Millisecond)

	_, err := writer.Write(data)
	gomega.Expect(err).To(gomega.BeNil())
}

// BadWriter is a simple struct that will return an error when trying to Write
type BadWriter struct {
	Err error
}

func (bw *BadWriter) Write(data []byte) (int, error) {
	return 0, bw.Err
}
