package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("next", func() {
	var (
		fs         afero.Fs
		afs        afero.Afero
		testConfig Config
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		afs = afero.Afero{Fs: fs}
		testConfig = Config{
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
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		res, err := nextPipeline(afs, "patch")
		Expect(err).To(BeNil())
		Expect(res).To(Equal("0.1.1\n"))
	})

	It("echos next minor version", func() {
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		res, err := nextPipeline(afs, "minor")
		Expect(err).To(BeNil())
		Expect(res).To(Equal("0.2.0\n"))
	})

	It("echos next major version", func() {
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		res, err := nextPipeline(afs, "major")
		Expect(err).To(BeNil())
		Expect(res).To(Equal("1.0.0\n"))
	})

	It("fails if bad config file", func() {
		err := afs.Remove(configPath)
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

	It("returns 0.0.1 if no versions found (patch)", func() {
		// include just the header
		_, err := afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())
		res, err := nextPipeline(afs, "patch")
		Expect(err).To(BeNil())
		Expect(res).To(Equal("0.0.1\n"))
	})

	It("fails if unable to get versions", func() {
		// no files, means bad read for get versions
		_, err := nextPipeline(afs, "major")
		Expect(err).NotTo(BeNil())
	})

})
