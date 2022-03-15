package cmd

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/miniscruff/changie/core"
)

var _ = Describe("next", func() {
	var (
		fs         afero.Fs
		afs        afero.Afero
		testConfig core.Config
		builder    strings.Builder
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
			Kinds:         []core.KindConfig{},
		}
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		builder = strings.Builder{}

		nextPrereleaseFlag = nil
		nextMetaFlag = nil
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

		err = nextPipeline(afs, &builder, "patch", nil, nil)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("v0.1.1"))
	})

	It("echos next version with prerelease and meta", func() {
		// major and minor are not tested directly
		// as next version is tested in utils
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "patch", []string{"b1"}, []string{"hash"})
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("v0.1.1-b1+hash"))
	})

	It("fails if bad config file", func() {
		err := afs.Remove(core.ConfigPaths[0])
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "major", nil, nil)
		Expect(err).NotTo(BeNil())
	})

	It("fails if part not supported", func() {
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "notsupported", nil, nil)
		Expect(err).NotTo(BeNil())
	})

	It("fails if unable to get versions", func() {
		// no files, means bad read for get versions
		err := nextPipeline(afs, &builder, "major", nil, nil)
		Expect(err).NotTo(BeNil())
	})

	It("fails if bump other than patch, minor, major", func() {
		err := nextPipeline(afs, &builder, "v1.2.3", nil, nil)
		Expect(err).NotTo(BeNil())
	})
})
