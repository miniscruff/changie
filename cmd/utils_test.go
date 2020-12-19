package cmd

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Utils", func() {
	It("append file appends two files", func() {
		memFs := afero.NewMemMapFs()
		rootFile, err := memFs.Create("root.txt")
		Expect(err).To(BeNil())

		rootFile.WriteString("root")

		otherFile, err := memFs.Create("append.txt")
		Expect(err).To(BeNil())

		otherFile.WriteString(" append")
		otherFile.Close()

		err = appendFile(memFs, rootFile, "append.txt")
		Expect(err).To(BeNil())

		rootFile.Close()

		bytes, err := afero.ReadFile(memFs, "root.txt")
		Expect(err).To(BeNil())
		Expect(string(bytes)).To(Equal("root append"))
	})

	It("append file fails if open fails", func() {
		mockError := errors.New("bad open")
		mockFs := newMockFs()
		mockFs.mockOpen = func(filename string) (afero.File, error) {
			return nil, mockError
		}

		rootFile := newMockFile(mockFs, "root.txt")
		defer rootFile.Close()

		err := appendFile(mockFs, rootFile, "dummy.txt")
		Expect(err).To(Equal(mockError))
	})
})
