package cmd

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
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
		testConfig.Save(afs.WriteFile)
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
		Expect(err).NotTo(BeNil())
	})

	It("fails if unable to read dir", func() {
		// no files, means bad read
		err := mergePipeline(afs, afs.Create)
		Expect(err).NotTo(BeNil())
	})
})
