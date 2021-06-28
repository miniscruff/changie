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
