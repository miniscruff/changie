package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"

	"github.com/miniscruff/changie/shared"
	. "github.com/miniscruff/changie/testutils"
)

var _ = Describe("Utils", func() {
	It("append file appends two files", func() {
		memFs := afero.NewMemMapFs()
		afs := afero.Afero{Fs: memFs}
		rootFile, err := memFs.Create("root.txt")
		Expect(err).To(BeNil())

		_, err = rootFile.WriteString("root")
		Expect(err).To(BeNil())

		err = afs.WriteFile("append.txt", []byte(" append"), CreateFileMode)
		Expect(err).To(BeNil())

		err = AppendFile(afs.Open, rootFile, "append.txt")
		Expect(err).To(BeNil())

		rootFile.Close()

		bytes, err := afero.ReadFile(memFs, "root.txt")
		Expect(err).To(BeNil())
		Expect(string(bytes)).To(Equal("root append"))
	})

	It("append file fails if open fails", func() {
		mockError := errors.New("bad open")
		mockFs := NewMockFS()
		mockFs.MockOpen = func(filename string) (afero.File, error) {
			return nil, mockError
		}

		rootFile := NewMockFile(mockFs, "root.txt")
		defer rootFile.Close()

		err := AppendFile(mockFs.Open, rootFile, "dummy.txt")
		Expect(err).To(Equal(mockError))
	})

	It("get all versions returns all versions", func() {
		config := Config{
			ChangesDir: "changes",
			HeaderPath: "header.md",
		}
		fs := NewMockFS()
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

		vers, err := GetAllVersions(mockRead, config, false)
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

		vers, err := GetAllVersions(mockRead, config, false)
		Expect(vers).To(BeEmpty())
		Expect(err).To(Equal(mockError))
	})

	createVersions := func(latestVersion string) (shared.ReadDirer, Config) {
		config := Config{
			ChangesDir: "changes",
			HeaderPath: "header.md",
		}
		fs := NewMockFS()
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

		ver, err := GetLatestVersion(mockRead, config, false)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.2.0"))
	})

	It("get latest version with prerelease", func() {
		mockRead, config := createVersions("v0.2.0-rc1")

		vers, err := GetLatestVersion(mockRead, config, false)
		Expect(err).To(BeNil())
		Expect(vers.Original()).To(Equal("v0.2.0-rc1"))
	})

	It("get latest version with prerelease skipped", func() {
		mockRead, config := createVersions("v0.2.0-rc1")

		vers, err := GetLatestVersion(mockRead, config, true)
		Expect(err).To(BeNil())
		Expect(vers.Original()).To(Equal("v0.1.0"))
	})

	It("get latest version returns v0.0.0 if no versions exist", func() {
		config := Config{
			ChangesDir: "changes",
			HeaderPath: "header.md",
		}
		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{}, nil
		}

		ver, err := GetLatestVersion(mockRead, config, false)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.0.0"))
	})

	It("get latest version returns error if get all version errors", func() {
		config := Config{}
		mockError := errors.New("bad stuff")
		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{}, mockError
		}

		vers, err := GetLatestVersion(mockRead, config, false)
		Expect(vers).To(BeNil())
		Expect(err).To(Equal(mockError))
	})

	It("get next version returns error if get next version errors", func() {
		config := Config{}
		mockError := errors.New("bad stuff")
		mockRead := func(dirname string) ([]os.FileInfo, error) {
			return []os.FileInfo{}, mockError
		}

		vers, err := GetNextVersion(mockRead, config, "major", nil, nil)
		Expect(vers).To(BeNil())
		Expect(err).To(Equal(mockError))
	})

	It("get next version returns error if invalid version", func() {
		mockRead, config := createVersions("v0.2.0")

		vers, err := GetNextVersion(mockRead, config, "a", []string{}, []string{})
		Expect(vers).To(BeNil())
		Expect(err).To(Equal(ErrBadVersionOrPart))
	})

	It("get next version works on brand new version", func() {
		mockRead, config := createVersions("v0.2.0")

		ver, err := GetNextVersion(mockRead, config, "v0.4.0", nil, nil)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.4.0"))
	})

	It("get next version works on only major version", func() {
		mockRead, config := createVersions("v0.2.0")

		ver, err := GetNextVersion(mockRead, config, "v2", nil, nil)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v2"))
	})

	It("get next version works on major and minor versions", func() {
		mockRead, config := createVersions("v0.2.0")

		ver, err := GetNextVersion(mockRead, config, "v1.2", nil, nil)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v1.2"))
	})

	It("get next version with short semver and prerelease", func() {
		mockRead, config := createVersions("v0.2.0")

		ver, err := GetNextVersion(mockRead, config, "v1.5", []string{"rc2"}, nil)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v1.5.0-rc2"))
	})

	It("get next version works on major version", func() {
		mockRead, config := createVersions("v0.2.0")

		ver, err := GetNextVersion(mockRead, config, "major", nil, nil)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v1.0.0"))
	})

	It("get next version works on minor version", func() {
		mockRead, config := createVersions("v0.2.2")

		ver, err := GetNextVersion(mockRead, config, "minor", nil, nil)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.3.0"))
	})

	It("get next version works on patch version", func() {
		mockRead, config := createVersions("v0.2.5")

		ver, err := GetNextVersion(mockRead, config, "patch", nil, nil)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.2.6"))
	})

	It("get next version with prerelease", func() {
		mockRead, config := createVersions("v0.2.5")

		ver, err := GetNextVersion(mockRead, config, "patch", []string{"b1", "amd64"}, nil)
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.2.6-b1.amd64"))
	})

	It("get next version with meta", func() {
		mockRead, config := createVersions("v0.2.5")

		ver, err := GetNextVersion(mockRead, config, "patch", nil, []string{"githash", "amd64"})
		Expect(err).To(BeNil())
		Expect(ver.Original()).To(Equal("v0.2.6+githash.amd64"))
	})

	It("get next version error with bad prerelease", func() {
		mockRead, config := createVersions("v0.2.5")

		_, err := GetNextVersion(mockRead, config, "patch", []string{"0005"}, nil)
		Expect(err).NotTo(BeNil())
	})

	It("get next version error with bad meta", func() {
		mockRead, config := createVersions("v0.2.5")

		_, err := GetNextVersion(mockRead, config, "patch", nil, []string{"&*(&*(&"})
		Expect(err).NotTo(BeNil())
	})

	It("can find change files", func() {
		memFs := afero.NewMemMapFs()
		afs := afero.Afero{Fs: memFs}

		config := Config{
			ChangesDir:    ".chng",
			UnreleasedDir: "unrel",
		}

		_, _ = memFs.Create(".chng/unrel/a.yaml")
		_, _ = memFs.Create(".chng/unrel/b.yaml")
		_, _ = memFs.Create(".chng/alpha/c.yaml")
		_, _ = memFs.Create(".chng/alpha/d.yaml")
		_, _ = memFs.Create(".chng/beta/e.yaml")
		_, _ = memFs.Create(".chng/beta/f.yaml")
		_, _ = memFs.Create(".chng/beta/ignored.md")
		_, _ = memFs.Create(".chng/ignored/g.yaml")
		_, _ = memFs.Create(".chng/h.md") // ignored

		files, err := FindChangeFiles(config, afs.ReadDir, []string{"alpha", "beta"})
		Expect(err).To(BeNil())
		Expect(files).To(Equal([]string{
			filepath.Join(".chng", "alpha", "c.yaml"),
			filepath.Join(".chng", "alpha", "d.yaml"),
			filepath.Join(".chng", "beta", "e.yaml"),
			filepath.Join(".chng", "beta", "f.yaml"),
			filepath.Join(".chng", "unrel", "a.yaml"),
			filepath.Join(".chng", "unrel", "b.yaml"),
		}))
	})

	It("change files fails if dir is not accessible", func() {
		memFs := afero.NewMemMapFs()
		afs := afero.Afero{Fs: memFs}

		config := Config{
			ChangesDir:    ".chng",
			UnreleasedDir: "unrel",
		}

		_, _ = memFs.Create(".chng/unrel/a.yaml")

		_, err := FindChangeFiles(config, afs.ReadDir, []string{"../../invalid"})
		Expect(err).NotTo(BeNil())
	})

	It("can write newlines", func() {
		var writer strings.Builder
		err := WriteNewlines(&writer, 3)
		Expect(err).To(BeNil())
		Expect(writer.String()).To(Equal("\n\n\n"))
	})

	It("skips newlines if lines is 0", func() {
		var writer strings.Builder
		err := WriteNewlines(&writer, 0)
		Expect(err).To(BeNil())
		Expect(writer.String()).To(Equal(""))
	})

	It("returns error on bad write newlines", func() {
		errMock := errors.New("mock error")
		writer := BadWriter{Err: errMock}

		err := WriteNewlines(&writer, 3)
		Expect(err).To(Equal(errMock))
	})
})
