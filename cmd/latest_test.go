package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Latest", func() {
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

	It("echos latest version", func() {
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		removePrefix = false
		res, err := latestPipeline(afs)
		Expect(err).To(BeNil())
		Expect(res).To(Equal("v0.1.0\n"))
	})

	It("echos latest version without prefix", func() {
		_, err := afs.Create("chgs/not-a-version.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())

		removePrefix = true
		res, err := latestPipeline(afs)
		Expect(err).To(BeNil())
		Expect(res).To(Equal("0.1.0\n"))
	})

	It("fails if bad config file", func() {
		err := afs.Remove(configPath)
		Expect(err).To(BeNil())

		_, err = latestPipeline(afs)
		Expect(err).NotTo(BeNil())
	})

	It("fails if unable to get versions", func() {
		// no files, means bad read for get versions
		_, err := latestPipeline(afs)
		Expect(err).NotTo(BeNil())
	})

	It("returns v0.0.0 if no versions found", func() {
		// include just the header
		_, err := afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())
		res, err := latestPipeline(afs)
		Expect(err).To(BeNil())
		Expect(res).To(Equal("v0.0.0\n"))
	})
})
