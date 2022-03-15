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

		err = nextPipeline(afs, &builder, "patch", nil, nil, false)
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

		err = nextPipeline(afs, &builder, "patch", []string{"b1"}, []string{"hash"}, false)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("v0.1.1-b1+hash"))
	})

	It("fails if bad config file", func() {
		err := afs.Remove(core.ConfigPaths[0])
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "major", nil, nil, false)
		Expect(err).NotTo(BeNil())
	})

	It("fails if part not supported", func() {
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/head.tpl.md")
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "notsupported", nil, nil, false)
		Expect(err).NotTo(BeNil())
	})

	It("fails if unable to get versions", func() {
		// no files, means bad read for get versions
		err := nextPipeline(afs, &builder, "major", nil, nil, false)
		Expect(err).NotTo(BeNil())
	})

	It("retrieves next with prerelease, default mode", func() {
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v1.1.1-staging.md")
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "patch", []string{"staging"}, nil, false)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("v1.1.1-staging"))
	})

	It("retrieves next patch with prerelease, force patch", func() {
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v1.1.1-staging.md")
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "patch", []string{"staging"}, nil, true)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("v1.1.2-staging"))
	})

	It("retrieves next minor with prerelease, no force patch", func() {
		// mode affects patch
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v1.1.1-staging.md")
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "minor", []string{"staging"}, nil, false)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("v1.2.0-staging"))
	})

	It("retrieves next minor with prerelease, force patch has no effect", func() {
		// mode doesn't affect minor
		_, err := afs.Create("chgs/v0.0.1.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v0.1.0.md")
		Expect(err).To(BeNil())
		_, err = afs.Create("chgs/v1.1.1-staging.md")
		Expect(err).To(BeNil())

		err = nextPipeline(afs, &builder, "minor", []string{"staging"}, nil, true)
		Expect(err).To(BeNil())
		Expect(builder.String()).To(Equal("v1.2.0-staging"))
	})
})
