package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/miniscruff/changie/core"
	. "github.com/miniscruff/changie/test_utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
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
			Kinds:         []string{},
		}
	})

	It("builds default skeleton", func() {
		err := initPipeline(afs.MkdirAll, afs.WriteFile, testConfig)

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
		err := initPipeline(badMkdir, afs.WriteFile, testConfig)
		Expect(err).To(Equal(mockError))
	})

	DescribeTable("error creating file",
		func(filename string) {
			mockWriteFile := func(path string, data []byte, perm os.FileMode) error {
				if path == filename {
					return mockError
				}
				return nil
			}
			err := initPipeline(afs.MkdirAll, mockWriteFile, testConfig)
			Expect(err).To(Equal(mockError))
		},
		Entry("config file", ".changie.yaml"),
		Entry("changelog", "changelog.md"),
		Entry("header", filepath.Join("chgs", "head.tpl.md")),
		Entry("git keep", filepath.Join("chgs", "unrel", ".gitkeep")),
	)
})
