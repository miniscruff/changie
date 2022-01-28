package cmd

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/miniscruff/changie/core"
	. "github.com/miniscruff/changie/testutils"
)

var _ = Describe("Merge", func() {
	var (
		fs         *MockFS
		afs        afero.Afero
		testConfig core.Config
	)
	BeforeEach(func() {
		fs = NewMockFS()
		afs = afero.Afero{Fs: fs}
		testConfig = core.Config{
			ChangesDir:    "news",
			UnreleasedDir: "future",
			HeaderPath:    "header.rst",
			ChangelogPath: "news.md",
			VersionExt:    "md",
			VersionFormat: "## {{.Version}}",
			KindFormat:    "### {{.Kind}}",
			ChangeFormat:  "* {{.Body}}",
			Kinds: []core.KindConfig{
				{Label: "added"},
				{Label: "removed"},
				{Label: "other"},
			},
			Replacements: []core.Replacement{
				{
					Path:    "replace.json",
					Find:    `  "version": ".*",`,
					Replace: `  "version": "{{.VersionNoPrefix}}",`,
				},
			},
		}
	})

	saveAndCheckConfig := func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())
	}

	It("can merge versions", func() {
		saveAndCheckConfig()
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

		jsonContents := `{
  "key": "value",
  "version": "old-version",
}`
		newContents := `{
  "key": "value",
  "version": "0.2.0",
}`
		err = afs.WriteFile("replace.json", []byte(jsonContents), os.ModePerm)
		Expect(err).To(BeNil())

		out, err := mergePipeline(afs, afs.Create, false)
		Expect(out).To(BeEmpty())
		Expect(err).To(BeNil())

		changeContents := `a simple header


second version

first version`

		Expect("news.md").To(HaveContents(afs, changeContents))
		Expect("replace.json").To(HaveContents(afs, newContents))
	})

	It("returns new version with dry run", func() {
		saveAndCheckConfig()
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

		changeContents := `a simple header


second version

first version`

		out, err := mergePipeline(afs, afs.Create, true)
		Expect(out).To(Equal(changeContents))
		Expect(err).To(BeNil())
	})

	It("skips versions if none found", func() {
		saveAndCheckConfig()

		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, os.ModePerm)
		Expect(err).To(BeNil())

		out, err := mergePipeline(afs, afs.Create, false)
		Expect(out).To(BeEmpty())
		Expect(err).To(BeNil())

		changeContents := "a simple header\n"
		Expect("news.md").To(HaveContents(afs, changeContents))
	})

	It("returns error on bad config", func() {
		out, err := mergePipeline(afs, afs.Create, false)
		Expect(out).To(BeEmpty())
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad changelog path", func() {
		saveAndCheckConfig()
		_, err := afs.Create(filepath.Join("news", "ignored.txt"))
		Expect(err).To(BeNil())

		mockFile := NewMockFile(fs, "header")
		fs.MockOpen = func(filename string) (afero.File, error) {
			if filename == filepath.Join("news", "header.rst") {
				return mockFile, nil
			}

			return fs.MemFS.Open(filename)
		}

		badError := errors.New("bad create")
		badCreate := func(filename string) (afero.File, error) {
			var f afero.File
			return f, badError
		}

		out, err := mergePipeline(afs, badCreate, false)
		Expect(out).To(BeEmpty())
		Expect(err).To(Equal(badError))
	})

	It("returns error if unable to read changes", func() {
		saveAndCheckConfig()
		// no files, means bad read
		out, err := mergePipeline(afs, afs.Create, false)
		Expect(out).To(BeEmpty())
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad header file", func() {
		saveAndCheckConfig()
		mockError := errors.New("bad open")
		fs.MockOpen = func(filename string) (afero.File, error) {
			if filename == filepath.Join("news", "header.rst") {
				return nil, mockError
			}

			return fs.MemFS.Open(filename)
		}

		// need at least one change to exist
		onePath := filepath.Join("news", "v0.1.0.md")
		oneChanges := []byte("first version")
		err := afs.WriteFile(onePath, oneChanges, os.ModePerm)
		Expect(err).To(BeNil())

		out, err := mergePipeline(afs, afs.Create, false)
		Expect(out).To(BeEmpty())
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad changelog write string", func() {
		saveAndCheckConfig()
		badError := errors.New("bad write string")
		mockFile := NewMockFile(afs, "changelog")
		mockCreate := func(filename string) (afero.File, error) {
			if filename == "news.md" {
				return mockFile, nil
			}

			return afs.Create(filename)
		}

		mockFile.MockWrite = func(data []byte) (int, error) {
			return 0, badError
		}

		// we need a header and at least one version
		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, os.ModePerm)
		Expect(err).To(BeNil())

		onePath := filepath.Join("news", "v0.1.0.md")

		oneChanges := []byte("first version")
		err = afs.WriteFile(onePath, oneChanges, os.ModePerm)
		Expect(err).To(BeNil())

		out, err := mergePipeline(afs, mockCreate, false)
		Expect(out).To(BeEmpty())
		Expect(err).To(Equal(badError))
	})

	It("returns error on bad second append", func() {
		saveAndCheckConfig()
		badError := errors.New("bad write string")

		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, os.ModePerm)
		Expect(err).To(BeNil())

		onePath := filepath.Join("news", "v0.1.0.md")
		oneChanges := []byte("first version")
		err = afs.WriteFile(onePath, oneChanges, os.ModePerm)
		Expect(err).To(BeNil())

		// create a version file, then fail to open it the second time
		fs.MockOpen = func(filename string) (afero.File, error) {
			if filename == onePath {
				return nil, badError
			}

			return fs.MemFS.Open(filename)
		}

		out, err := mergePipeline(afs, afs.Create, false)
		Expect(out).To(BeEmpty())
		Expect(err).To(Equal(badError))
	})

	It("returns error on bad replacement", func() {
		testConfig.Replacements[0].Replace = "{{bad....}}"
		saveAndCheckConfig()

		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, os.ModePerm)
		Expect(err).To(BeNil())

		onePath := filepath.Join("news", "v0.1.0.md")

		oneChanges := []byte("first version")
		err = afs.WriteFile(onePath, oneChanges, os.ModePerm)
		Expect(err).To(BeNil())

		out, err := mergePipeline(afs, afs.Create, false)
		Expect(out).To(BeEmpty())
		Expect(err).NotTo(BeNil())
	})
})
