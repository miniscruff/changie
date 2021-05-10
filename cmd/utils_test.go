package cmd

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
)

var _ = Describe("Utils", func() {
	It("append file appends two files", func() {
		memFs := afero.NewMemMapFs()
		afs := afero.Afero{Fs: memFs}
		rootFile, err := memFs.Create("root.txt")
		Expect(err).To(BeNil())

		_, err = rootFile.WriteString("root")
		Expect(err).To(BeNil())

		err = afs.WriteFile("append.txt", []byte(" append"), os.ModePerm)
		Expect(err).To(BeNil())

		err = appendFile(afs.Open, rootFile, "append.txt")
		Expect(err).To(BeNil())

		rootFile.Close()

		bytes, err := afero.ReadFile(memFs, "root.txt")
		Expect(err).To(BeNil())
		Expect(string(bytes)).To(Equal("root append"))
	})

	It("append file fails if open fails", func() {
		mockError := errors.New("bad open")
		mockFs := newMockFs()
		mockFs.mockOpen = func(filename string) (afero.File, error) {
			return nil, mockError
		}

		rootFile := newMockFile(mockFs, "root.txt")
		defer rootFile.Close()

		err := appendFile(mockFs.Open, rootFile, "dummy.txt")
		Expect(err).To(Equal(mockError))
	})

	It("get all versions returns all versions", func() {
		config := Config{
			ChangesDir: "changes",
			HeaderPath: "header.md",
		}
		fs := newMockFs()
		afs := afero.Afero{Fs: fs}

		headerFile, err := afs.Create(filepath.Join("changes", "header.md"))
		Expect(err).To(BeNil())
		file1, err := afs.Create(filepath.Join("changes", "v0.1.0.md"))
		Expect(err).To(BeNil())
		file2, err := afs.Create(filepath.Join("changes", "v0.2.0.md"))
		Expect(err).To(BeNil())
		file3, err := afs.Create(filepath.Join("changes", "not-sem-ver.md"))
		Expect(err).To(BeNil())

		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{
				headerFile.(*mem.File).Info(),
				file1.(*mem.File).Info(),
				file2.(*mem.File).Info(),
				file3.(*mem.File).Info(),
			}, nil
		}

		vers, err := getAllVersions(mockRead, config)
		Expect(err).To(BeNil())
		Expect(vers[0].Original()).To(Equal("v0.2.0"))
		Expect(vers[1].Original()).To(Equal("v0.1.0"))
	})

	It("get all versions returns error on bad read dir", func() {
		config := Config{}
		mockError := errors.New("bad stuff")
		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{}, mockError
		}

		vers, err := getAllVersions(mockRead, config)
		Expect(vers).To(BeEmpty())
		Expect(err).To(Equal(mockError))
	})

	createVersions := func(latestVersion string) (ReadDirer, Config) {
		config := Config{
			ChangesDir: "changes",
			HeaderPath: "header.md",
		}
		fs := newMockFs()
		afs := afero.Afero{Fs: fs}

		headerFile, err := afs.Create(filepath.Join("changes", "header.md"))
		Expect(err).To(BeNil())
		file1, err := afs.Create(filepath.Join("changes", "v0.1.0.md"))
		Expect(err).To(BeNil())
		file2, err := afs.Create(filepath.Join("changes", latestVersion+".md"))
		Expect(err).To(BeNil())
		file3, err := afs.Create(filepath.Join("changes", "not-sem-ver.md"))
		Expect(err).To(BeNil())

		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{
				headerFile.(*mem.File).Info(),
				file1.(*mem.File).Info(),
				file2.(*mem.File).Info(),
				file3.(*mem.File).Info(),
			}, nil
		}
		return mockRead, config
	}

	It("get latest version returns most recent version", func() {
		mockRead, config := createVersions("v0.2.0")

		ver, err := getLatestVersion(mockRead, config)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.2.0"))
	})

	It("get latest version returns v0.0.0 if no versions exist", func() {
		config := Config{
			ChangesDir: "changes",
			HeaderPath: "header.md",
		}
		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{}, nil
		}

		ver, err := getLatestVersion(mockRead, config)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.0.0"))
	})

	It("get latest version returns error if get all version errors", func() {
		config := Config{}
		mockError := errors.New("bad stuff")
		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{}, mockError
		}

		vers, err := getLatestVersion(mockRead, config)
		Expect(vers).To(BeNil())
		Expect(err).To(Equal(mockError))
	})

	It("get next version returns error if get all version errors", func() {
		config := Config{}
		mockError := errors.New("bad stuff")
		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{}, mockError
		}

		vers, err := getNextVersion(mockRead, config, "a")
		Expect(vers).To(BeNil())
		Expect(err).To(Equal(mockError))
	})

	It("get next version works on brand new version", func() {
		mockRead, config := createVersions("v0.2.0")

		ver, err := getNextVersion(mockRead, config, "v0.4.0")
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.4.0"))
	})

	It("get next version works on major version", func() {
		mockRead, config := createVersions("v0.2.0")

		ver, err := getNextVersion(mockRead, config, "major")
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v1.0.0"))
	})

	It("get next version works on minor version", func() {
		mockRead, config := createVersions("v0.2.2")

		ver, err := getNextVersion(mockRead, config, "minor")
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.3.0"))
	})

	It("get next version works on patch version", func() {
		mockRead, config := createVersions("v0.2.5")

		ver, err := getNextVersion(mockRead, config, "patch")
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.2.6"))
	})
})
