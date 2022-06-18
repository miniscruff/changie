package cmd

import (
	"errors"
	"strings"
	"path/filepath"
	"os"

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
		stdout          *os.File
	)
	BeforeEach(func() {
		fs = NewMockFS()
		afs = afero.Afero{Fs: fs}
		stdout = os.Stdout
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

	AfterEach(func() {
		// reset our cli vars
		mergeDryRunOut = stdout
		_mergeDryRun = false
	})

	saveAndCheckConfig := func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())
	}

	It("can merge versions", func() {
		testConfig.HeaderPath = ""
		testConfig.Replacements = nil
		saveAndCheckConfig()

		onePath := filepath.Join("news", "v0.1.0.md")
		twoPath := filepath.Join("news", "v0.2.0.md")

		oneChanges := []byte("first version")
		err := afs.WriteFile(onePath, oneChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		twoChanges := []byte("second version")
		err = afs.WriteFile(twoPath, twoChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		err = mergePipeline(afs)
		Expect(err).To(BeNil())

		changeContents := `

second version

first version`

		Expect("news.md").To(HaveContents(afs, changeContents))
	})

	It("can merge versions with header and replacement", func() {
		saveAndCheckConfig()
		onePath := filepath.Join("news", "v0.1.0.md")
		twoPath := filepath.Join("news", "v0.2.0.md")

		oneChanges := []byte("first version")
		err := afs.WriteFile(onePath, oneChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		twoChanges := []byte("second version")
		err = afs.WriteFile(twoPath, twoChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		headerContents := []byte("a simple header\n")
		err = afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, core.CreateFileMode)
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
		err = afs.WriteFile("replace.json", []byte(jsonContents), core.CreateFileMode)
		Expect(err).To(BeNil())

		err = mergePipeline(afs)
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
		err := afs.WriteFile(onePath, oneChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		twoChanges := []byte("second version")
		err = afs.WriteFile(twoPath, twoChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		headerContents := []byte("a simple header\n")
		err = afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, core.CreateFileMode)
		Expect(err).To(BeNil())

		_, err = afs.Create(filepath.Join("news", "ignored.txt"))
		Expect(err).To(BeNil())

		changeContents := `a simple header


second version

first version`

		var writer strings.Builder
		mergeDryRunOut = &writer
		_mergeDryRun = true
		err = mergePipeline(afs)
		Expect(writer.String()).To(Equal(changeContents))
		Expect(err).To(BeNil())
	})

	It("skips versions if none found", func() {
		saveAndCheckConfig()

		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, core.CreateFileMode)
		Expect(err).To(BeNil())

		changeContents := "a simple header\n"

		var writer strings.Builder
		mergeDryRunOut = &writer
		_mergeDryRun = true
		err = mergePipeline(afs)
		Expect(writer.String()).To(Equal(changeContents))
		Expect(err).To(BeNil())
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
		fs.MockCreate = func(filename string) (afero.File, error) {
			var f afero.File
			return f, badError
		}

		err = mergePipeline(afs)
		Expect(err).To(Equal(badError))
	})

	It("returns error if unable to read changes", func() {
		saveAndCheckConfig()
		// no files, means bad read
		err := mergePipeline(afs)
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
		err := afs.WriteFile(onePath, oneChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		err = mergePipeline(afs)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad second append", func() {
		saveAndCheckConfig()
		badError := errors.New("bad write string")

		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, core.CreateFileMode)
		Expect(err).To(BeNil())

		onePath := filepath.Join("news", "v0.1.0.md")
		oneChanges := []byte("first version")
		err = afs.WriteFile(onePath, oneChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		// create a version file, then fail to open it the second time
		fs.MockOpen = func(filename string) (afero.File, error) {
			if filename == onePath {
				return nil, badError
			}

			return fs.MemFS.Open(filename)
		}

		err = mergePipeline(afs)
		Expect(err).To(Equal(badError))
	})

	It("returns error on bad replacement", func() {
		testConfig.Replacements[0].Replace = "{{bad....}}"
		saveAndCheckConfig()

		headerContents := []byte("a simple header\n")
		err := afs.WriteFile(filepath.Join("news", "header.rst"), headerContents, core.CreateFileMode)
		Expect(err).To(BeNil())

		onePath := filepath.Join("news", "v0.1.0.md")

		oneChanges := []byte("first version")
		err = afs.WriteFile(onePath, oneChanges, core.CreateFileMode)
		Expect(err).To(BeNil())

		err = mergePipeline(afs)
		Expect(err).NotTo(BeNil())
	})
})
