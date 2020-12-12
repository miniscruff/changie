package cmd

/*

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"os"
)

var _ = Describe("Init", func() {
	var (
		fs        afero.Fs
		afs       *afero.Afero
		mfs       *mockFs
		mf        *mockFile
		mockError error
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		afs = &afero.Afero{Fs: fs}
		mfs = newMockFs()
		mf = newMockFile(mfs)
		mockError = errors.New("dummy mock error")
	})

	It("builds default skeleton", func() {
		p := NewProject(fs, ProjectConfig{})
		err := p.Init()

		Expect(err).To(BeNil())
		Expect("changes/").To(BeADir(afs))
		Expect("changes/unreleased").To(BeADir(afs))
		Expect("changes/unreleased/.gitkeep").To(BeAnEmptyFile(afs))
		Expect("changes/header.tpl.md").To(HaveContents(afs, defaultHeader))
		Expect("CHANGELOG.md").To(HaveContents(afs, defaultChangelog))
	})

	testBadFs := func(fs afero.Fs) {
		p := NewProject(fs, ProjectConfig{})
		err := p.Init()
		Expect(err).To(Equal(mockError))
	}

	It("returns error on bad mkdir for unreleased", func() {
		mfs.mockMkdirAll = func(path string, mode os.FileMode) error {
			return mockError
		}
		testBadFs(mfs)
	})

	DescribeTable("error creating file",
		func(filename string) {
			mfs.mockCreate = func(path string) (afero.File, error) {
				if path == filename {
					return nil, mockError
				}
				return mfs.memFs.Create(path)
			}
			p := NewProject(mfs, ProjectConfig{})
			err := p.Init()
			Expect(err).To(Equal(mockError))
		},
		Entry("changelog", "CHANGELOG.md"),
		Entry("header", "changes/header.tpl.md"),
		Entry("git keep", "changes/unreleased/.gitkeep"),
	)

	DescribeTable("error writing files",
		func(filename string) {
			mfs.mockCreate = func(path string) (afero.File, error) {
				if path == filename {
					return mf, nil
				}
				return mfs.memFs.Create(path)
			}
			mf.mockWriteString = func(s string) (int, error) {
				return 0, mockError
			}
			p := NewProject(mfs, ProjectConfig{})
			err := p.Init()
			Expect(err).To(Equal(mockError))
		},
		Entry("changelog", "CHANGELOG.md"),
		Entry("header", "changes/header.tpl.md"),
	)
})

*/
