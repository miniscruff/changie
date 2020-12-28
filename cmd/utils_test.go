package cmd

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Utils", func() {
	It("append file appends two files", func() {
		memFs := afero.NewMemMapFs()
		afs := afero.Afero{Fs: memFs}
		rootFile, err := memFs.Create("root.txt")
		Expect(err).To(BeNil())

		_, err = rootFile.WriteString("root")
		Expect(err).To(BeNil())

		err = afs.WriteFile("append.txt", []byte(" append"), os.ModePerm)
		Expect(err).To(BeNil())

		err = appendFile(afs.Open, rootFile, "append.txt")
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

		err := appendFile(mockFs.Open, rootFile, "dummy.txt")
		Expect(err).To(Equal(mockError))
	})
})
