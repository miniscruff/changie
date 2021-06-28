package cmd

import (
	"github.com/miniscruff/changie/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("next", func() {
	var (
		fs         afero.Fs
		afs        afero.Afero
		testConfig core.Config
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		afs = afero.Afero{Fs: fs}
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
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())
	})

	It("echos next patch version", func() {
		// major and minor are not tested directly
		// as next version is tested in utils
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		res, err := nextPipeline(afs, "patch")
		Expect(err).To(BeNil())
		Expect(res).To(Equal("v0.1.1\n"))
	})

	It("fails if bad config file", func() {
		err := afs.Remove(core.ConfigPath)
		Expect(err).To(BeNil())

		_, err = nextPipeline(afs, "major")
		Expect(err).NotTo(BeNil())
	})

	It("fails if part not supported", func() {
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		_, err = nextPipeline(afs, "notsupported")
		Expect(err).NotTo(BeNil())
	})

	It("fails if unable to get versions", func() {
		// no files, means bad read for get versions
		_, err := nextPipeline(afs, "major")
		Expect(err).NotTo(BeNil())
	})
})
