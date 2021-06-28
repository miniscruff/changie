package core

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	. "github.com/miniscruff/changie/testutils"
)

var _ = Describe("Replacement", func() {
	var (
		fs  afero.Fs
		afs afero.Afero
	)
	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		afs = afero.Afero{Fs: fs}
	})

	It("can find and replace contents in a file", func() {
		startData := `first line
second line
third line
ignore me`
		endData := `first line
replaced here
third line
ignore me`

		err := afs.WriteFile("file.txt", []byte(startData), os.ModeTemporary)
		Expect(err).To(BeNil())

		rep := Replacement{
			Path:    "file.txt",
			Find:    "second line",
			Replace: "replaced here",
		}
		err = rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{})
		Expect(err).To(BeNil())
		Expect("file.txt").To(HaveContents(afs, endData))
	})

	It("can find and replace with template", func() {
		startData := `{
  "name": "demo-project",
  "version": "1.0.0",
}`
		endData := `{
  "name": "demo-project",
  "version": "1.1.0",
}`

		err := afs.WriteFile("file.txt", []byte(startData), os.ModeTemporary)
		Expect(err).To(BeNil())

		rep := Replacement{
			Path:    "file.txt",
			Find:    `  "version": ".*",`,
			Replace: `  "version": "{{.VersionNoPrefix}}",`,
		}
		err = rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{
			VersionNoPrefix: "1.1.0",
		})
		Expect(err).To(BeNil())
		Expect("file.txt").To(HaveContents(afs, endData))
	})

	It("returns error on bad file read", func() {
		rep := Replacement{
			Path: "does not exist",
		}
		err := rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{})
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad template parsing", func() {
		rep := Replacement{
			Replace: "{{.bad..}}",
		}
		err := rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{})
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad template execute", func() {
		rep := Replacement{
			Replace: "{{.bad..}}",
		}
		err := rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{})
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad write file", func() {
		startData := "some data"

		err := afs.WriteFile("file.txt", []byte(startData), os.ModeTemporary)
		Expect(err).To(BeNil())

		mockError := errors.New("bad write")
		badWrite := func(path string, data []byte, mode os.FileMode) error {
			return mockError
		}

		rep := Replacement{
			Path:    "file.txt",
			Find:    "some",
			Replace: "none",
		}
		err = rep.Execute(afs.ReadFile, badWrite, ReplaceData{})
		Expect(err).To(Equal(mockError))
	})
})
