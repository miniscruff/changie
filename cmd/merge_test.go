package cmd

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Merge", func() {
	var (
		fs         *mockFs
		afs        afero.Afero
		testConfig Config
	)
	BeforeEach(func() {
		fs = newMockFs()
		afs = afero.Afero{Fs: fs}
		testConfig = Config{
			ChangesDir:    "news",
			UnreleasedDir: "future",
			HeaderPath:    "header.rst",
			ChangelogPath: "news.md",
			VersionExt:    "md",
			VersionFormat: "## {{.Version}}",
			KindFormat:    "### {{.Kind}}",
			ChangeFormat:  "* {{.Body}}",
			Kinds: []string{
				"added", "removed", "other",
			},
		}

		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())
	})

	It("can merge versions", func() {
		onePath := filepath.Join("news", "v0.1.0.md")
		twoPath := filepath.Join("news", "v0.2.0.md")

		oneChanges := []byte("first version")
		err := afs.WriteFile(onePath, oneChanges, os.ModePerm)
		Expect(err).To(BeNil())

		twoChanges := []byte("second version")
		err = afs.WriteFile(twoPath, twoChanges, os.ModePerm)
		Expect(err).To(BeNil())

		headerContents := []byte("a simple header\n")
		err = afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, os.ModePerm)
		Expect(err).To(BeNil())

		_, err = afs.Create(filepath.Join("news", "ignored.txt"))
		Expect(err).To(BeNil())

		err = mergePipeline(afs, afs.Create)
		Expect(err).To(BeNil())

		changeContents := `a simple header


second version

first version`

		Expect("news.md").To(HaveContents(afs, changeContents))
	})

	It("returns error on bad config", func() {
		err := afs.Remove(configPath)
		Expect(err).To(BeNil())

		err = mergePipeline(afs, afs.Create)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad changelog path", func() {
		_, err := afs.Create(filepath.Join("news", "ignored.txt"))
		Expect(err).To(BeNil())

		badError := errors.New("bad create")
		badCreate := func(filename string) (afero.File, error) {
			var f afero.File
			return f, badError
		}

		err = mergePipeline(afs, badCreate)
		Expect(err).To(Equal(badError))
	})

	It("returns error if unable to read unreleased", func() {
		// no files, means bad read
		err := mergePipeline(afs, afs.Create)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad header file", func() {
		mockError := errors.New("bad open")
		fs.mockOpen = func(filename string) (afero.File, error) {
			if filename == filepath.Join("news", "header.rst") {
				return nil, mockError
			}

			return fs.memFs.Open(filename)
		}

		// need at least one change to exist
		onePath := filepath.Join("news", "v0.1.0.md")
		oneChanges := []byte("first version")
		err := afs.WriteFile(onePath, oneChanges, os.ModePerm)
		Expect(err).To(BeNil())

		err = mergePipeline(afs, afs.Create)
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad changelog write string", func() {
		badError := errors.New("bad write string")
		mockFile := newMockFile(afs, "changelog")
		mockCreate := func(filename string) (afero.File, error) {
			if filename == "news.md" {
				return mockFile, nil
			}

			return afs.Create(filename)
		}

		mockFile.mockWriteString = func(data string) (int, error) {
			if data == "\n\n" {
				return 0, badError
			}

			return mockFile.memFile.WriteString(data)
		}

		// we need a header and at least one version
		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, os.ModePerm)
		Expect(err).To(BeNil())

		onePath := filepath.Join("news", "v0.1.0.md")

		oneChanges := []byte("first version")
		err = afs.WriteFile(onePath, oneChanges, os.ModePerm)
		Expect(err).To(BeNil())

		err = mergePipeline(afs, mockCreate)
		Expect(err).To(Equal(badError))
	})

	It("returns error on bad second append", func() {
		badError := errors.New("bad write string")

		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, os.ModePerm)
		Expect(err).To(BeNil())

		onePath := filepath.Join("news", "v0.1.0.md")
		oneChanges := []byte("first version")
		err = afs.WriteFile(onePath, oneChanges, os.ModePerm)
		Expect(err).To(BeNil())

		// create a version file, then fail to open it the second time
		fs.mockOpen = func(filename string) (afero.File, error) {
			if filename == onePath {
				return nil, badError
			}

			return fs.memFs.Open(filename)
		}

		err = mergePipeline(afs, afs.Create)
		Expect(err).To(Equal(badError))
	})
})
