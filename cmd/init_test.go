package cmd

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/miniscruff/changie/core"
	. "github.com/miniscruff/changie/testutils"
)

var _ = Describe("Init", func() {
	var (
		fs         afero.Fs
		afs        afero.Afero
		mockError  error
		testConfig core.Config
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		afs = afero.Afero{Fs: fs}
		mockError = errors.New("dummy mock error")
		testConfig = core.Config{
			ChangesDir:    "chgs",
			UnreleasedDir: "unrel",
			HeaderPath:    "head.tpl.md",
			ChangelogPath: "changelog.md",
			VersionExt:    "md",
			VersionFormat: "",
			KindFormat:    "",
			ChangeFormat:  "",
			Kinds:         []core.KindConfig{},
		}
		initForce = false
	})

	It("builds default skeleton", func() {
		err := initPipeline(afs.MkdirAll, afs.WriteFile, afs.Exists, testConfig)

		Expect(err).To(BeNil())
		Expect("chgs/").To(BeADir(afs))
		Expect("chgs/unrel").To(BeADir(afs))
		Expect("chgs/unrel/.gitkeep").To(BeAnEmptyFile(afs))
		Expect("chgs/head.tpl.md").To(HaveContents(afs, defaultHeader))
		Expect("changelog.md").To(HaveContents(afs, defaultChangelog))
	})

	It("builds default skeleton if file exists but forced", func() {
		initForce = true
		fileExists := func(path string) (bool, error) {
			return true, nil
		}

		err := initPipeline(afs.MkdirAll, afs.WriteFile, fileExists, testConfig)

		Expect(err).To(BeNil())
		Expect("chgs/").To(BeADir(afs))
		Expect("chgs/unrel").To(BeADir(afs))
		Expect("chgs/unrel/.gitkeep").To(BeAnEmptyFile(afs))
		Expect("chgs/head.tpl.md").To(HaveContents(afs, defaultHeader))
		Expect("changelog.md").To(HaveContents(afs, defaultChangelog))
	})

	It("returns error on bad mkdir for unreleased", func() {
		badMkdir := func(path string, mode os.FileMode) error {
			return mockError
		}
		err := initPipeline(badMkdir, afs.WriteFile, afs.Exists, testConfig)
		Expect(err).To(Equal(mockError))
	})

	It("returns error if file exisits", func() {
		badFileExists := func(path string) (bool, error) {
			return true, nil
		}

		err := initPipeline(afs.MkdirAll, afs.WriteFile, badFileExists, testConfig)
		Expect(err).To(Equal(errConfigExists))
	})

	It("returns error if unable to check if file exists", func() {
		badFileExists := func(path string) (bool, error) {
			return false, errors.New("doesn't matter")
		}

		err := initPipeline(afs.MkdirAll, afs.WriteFile, badFileExists, testConfig)
		Expect(err).To(Equal(errConfigExists))
	})

	DescribeTable("error creating file",
		func(filename string) {
			mockWriteFile := func(path string, data []byte, perm os.FileMode) error {
				if path == filename {
					return mockError
				}
				return nil
			}
			err := initPipeline(afs.MkdirAll, mockWriteFile, afs.Exists, testConfig)
			Expect(err).To(Equal(mockError))
		},
		Entry("config file", ".changie.yaml"),
		Entry("changelog", "changelog.md"),
		Entry("header", filepath.Join("chgs", "head.tpl.md")),
		Entry("git keep", filepath.Join("chgs", "unrel", ".gitkeep")),
	)
})
