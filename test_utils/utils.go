package test_utils

import (
	"io"
	"time"

	. "github.com/onsi/gomega"
)

func DelayWrite(writer io.Writer, data []byte) {
	time.Sleep(50 * time.Millisecond)

	_, err := writer.Write(data)
	Expect(err).To(BeNil())
}
