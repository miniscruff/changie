package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Batch", func() {
	var (
		fs         *mockFs
		afs        afero.Afero
		mockError  error
		testConfig Config
	)
	BeforeEach(func() {
		fs = newMockFs()
		afs = afero.Afero{Fs: fs}
		mockError = errors.New("dummy mock error")
		testConfig = Config{
			ChangesDir:    "news",
			UnreleasedDir: "future",
			HeaderPath:    "header.rst",
			ChangelogPath: "news.md",
			VersionExt:    "md",
			VersionFormat: "## {{.Version}}",
			KindFormat:    "### {{.Kind}}",
			ChangeFormat:  "* {{.Body}}",
			Kinds: []string{
				"added", "removed", "other",
			},
		}

		// reset our shared header path
		versionHeaderPath = ""
	})

	It("can batch version", func() {
		// declared path but missing is accepted
		testConfig.VersionHeaderPath = "header.md"
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")
		newVerPath := filepath.Join("news", "v0.2.0.md")

		aVer := []byte("kind: added\nbody: A\n")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		bVer := []byte("kind: added\nbody: B\n")
		err = afs.WriteFile(filepath.Join(futurePath, "b.yaml"), bVer, os.ModePerm)
		Expect(err).To(BeNil())

		cVer := []byte("kind: removed\nbody: C\n")
		err = afs.WriteFile(filepath.Join(futurePath, "c.yaml"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "v0.2.0")
		Expect(err).To(BeNil())

		verContents := `## v0.2.0

### added
* A
* B

### removed
* C`

		Expect(newVerPath).To(HaveContents(afs, verContents))

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(0))
	})

	It("can batch version and bump", func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")
		newVerPath := filepath.Join("news", "v0.1.1.md")

		aVer := []byte("kind: added\nbody: A\n")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		bVer := []byte("kind: added\nbody: B\n")
		err = afs.WriteFile(filepath.Join(futurePath, "b.yaml"), bVer, os.ModePerm)
		Expect(err).To(BeNil())

		cVer := []byte("kind: removed\nbody: C\n")
		err = afs.WriteFile(filepath.Join(futurePath, "c.yaml"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		latestVer := []byte("contents do not matter")
		err = afs.WriteFile(filepath.Join("news", "v0.1.0.md"), latestVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "patch")
		Expect(err).To(BeNil())

		verContents := `## v0.1.1

### added
* A
* B

### removed
* C`

		Expect(newVerPath).To(HaveContents(afs, verContents))

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(0))
	})

	It("can batch version with header", func() {
		testConfig.VersionHeaderPath = "h.md"
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")
		newVerPath := filepath.Join("news", "v0.2.0.md")
		headerPath := filepath.Join(futurePath, testConfig.VersionHeaderPath)

		aVer := []byte("kind: added\nbody: A\n")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		bVer := []byte("kind: added\nbody: B\n")
		err = afs.WriteFile(filepath.Join(futurePath, "b.yaml"), bVer, os.ModePerm)
		Expect(err).To(BeNil())

		cVer := []byte("kind: removed\nbody: C\n")
		err = afs.WriteFile(filepath.Join(futurePath, "c.yaml"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		headerData := []byte("this is a new version that adds cool features")
		err = afs.WriteFile(headerPath, headerData, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "v0.2.0")
		Expect(err).To(BeNil())

		verContents := `## v0.2.0

this is a new version that adds cool features

### added
* A
* B

### removed
* C`

		Expect(newVerPath).To(HaveContents(afs, verContents))

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(0))
	})

	It("can batch version with header parameter", func() {
		versionHeaderPath = "head.md"
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")
		newVerPath := filepath.Join("news", "v0.2.0.md")

		aVer := []byte("kind: added\nbody: A\n")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		bVer := []byte("kind: added\nbody: B\n")
		err = afs.WriteFile(filepath.Join(futurePath, "b.yaml"), bVer, os.ModePerm)
		Expect(err).To(BeNil())

		cVer := []byte("kind: removed\nbody: C\n")
		err = afs.WriteFile(filepath.Join(futurePath, "c.yaml"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		headerData := []byte("this is a new version that adds cool features")
		err = afs.WriteFile(filepath.Join(futurePath, "head.md"), headerData, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "v0.2.0")
		Expect(err).To(BeNil())

		verContents := `## v0.2.0

this is a new version that adds cool features

### added
* A
* B

### removed
* C`

		Expect(newVerPath).To(HaveContents(afs, verContents))

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(0))
	})

	It("returns error on bad semver", func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")

		aVer := []byte("kind: added\nbody: A\n")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "not-semanticly-correct")
		Expect(err).To(Equal(errBadVersionOrPart))
	})

	It("returns error on bad config", func() {
		configData := []byte("not a proper config")
		err := afs.WriteFile(configPath, configData, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "v1.0.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad changes", func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")
		aVer := []byte("not a valid change")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "v1.0.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad batch", func() {
		testConfig.VersionFormat = "{{bad.format}"
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")
		aVer := []byte("kind: removed\nbody: C\n")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "v1.0.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad delete", func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")
		aVer := []byte("kind: removed\nbody: C\n")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		fs.mockRemove = func(name string) error {
			return mockError
		}

		err = batchPipeline(afs, "v1.0.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad header file", func() {
		versionHeaderPath = "header.md"
		badOpen := errors.New("bad open file")
		fs.mockOpen = func(name string) (afero.File, error) {
			if strings.HasSuffix(name, versionHeaderPath) {
				return nil, badOpen
			}
			return fs.memFs.Open(name)
		}

		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		futurePath := filepath.Join("news", "future")

		aVer := []byte("kind: added\nbody: A\n")
		err = afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(afs, "v0.2.0")
		Expect(err).To(Equal(badOpen))
	})

	It("can get all changes", func() {
		unrelPath := filepath.Join("news", "future")

		aVer := []byte("kind: added\nbody: A\n")
		err := afs.WriteFile(filepath.Join(unrelPath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		bVer := []byte("kind: added\nbody: B\n")
		err = afs.WriteFile(filepath.Join(unrelPath, "b.yaml"), bVer, os.ModePerm)
		Expect(err).To(BeNil())

		cVer := []byte("kind: removed\nbody: C\n")
		err = afs.WriteFile(filepath.Join(unrelPath, "c.yaml"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = afs.WriteFile(filepath.Join(unrelPath, "ignored.txt"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		changes, err := getChanges(afs, testConfig)
		Expect(err).To(BeNil())
		Expect(changes["added"][0].Kind).To(Equal("added"))
		Expect(changes["added"][0].Body).To(Equal("A"))
		Expect(changes["added"][1].Kind).To(Equal("added"))
		Expect(changes["added"][1].Body).To(Equal("B"))
		Expect(changes["removed"][0].Kind).To(Equal("removed"))
		Expect(changes["removed"][0].Body).To(Equal("C"))
	})

	It("returns err if unable to read directory", func() {
		fs.mockOpen = func(path string) (afero.File, error) {
			var f afero.File
			return f, mockError
		}

		_, err := getChanges(afs, testConfig)
		Expect(err).To(Equal(mockError))
	})

	It("returns err if bad changes file", func() {
		unrelPath := filepath.Join("news", "future")

		aVer := []byte("not a valid yaml:::::file---___")
		err := afs.WriteFile(filepath.Join(unrelPath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		_, err = getChanges(afs, testConfig)
		Expect(err).NotTo(BeNil())
	})

	It("can create new version file", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		vFile := newMockFile(fs, "v0.1.0.md")
		contents := ""

		fs.mockCreate = func(path string) (afero.File, error) {
			return vFile, nil
		}

		vFile.mockWrite = func(data []byte) (int, error) {
			contents += string(data)
			return len(data), nil
		}
		vFile.mockWriteString = func(data string) (int, error) {
			contents += data
			return len(data), nil
		}

		changesPerKind := map[string][]Change{
			"added":   {{Body: "w"}, {Body: "x"}},
			"removed": {{Body: "y"}, {Body: "z"}},
		}

		err = batchNewVersion(fs, testConfig, batchData{
			Version:        semver.MustParse("v0.1.0"),
			ChangesPerKind: changesPerKind,
		})
		Expect(err).To(BeNil())

		expected := `## v0.1.0

### added
* w
* x

### removed
* y
* z`
		Expect(contents).To(Equal(expected))
	})

	It("can create new version file with header", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		vFile := newMockFile(fs, "v0.1.0.md")
		contents := ""

		fs.mockCreate = func(path string) (afero.File, error) {
			return vFile, nil
		}

		vFile.mockWrite = func(data []byte) (int, error) {
			contents += string(data)
			return len(data), nil
		}
		vFile.mockWriteString = func(data string) (int, error) {
			contents += data
			return len(data), nil
		}

		changesPerKind := map[string][]Change{
			"added":   {{Body: "w"}, {Body: "x"}},
			"removed": {{Body: "y"}, {Body: "z"}},
		}

		err = batchNewVersion(fs, testConfig, batchData{
			Version:        semver.MustParse("v0.1.0"),
			ChangesPerKind: changesPerKind,
			Header:         "Some header we want included in our new version.\nCan also have newlines.",
		})
		Expect(err).To(BeNil())

		expected := `## v0.1.0

Some header we want included in our new version.
Can also have newlines.

### added
* w
* x

### removed
* y
* z`
		Expect(contents).To(Equal(expected))
	})

	It("returns error where when failing to create version file", func() {
		fs.mockCreate = func(path string) (afero.File, error) {
			var f afero.File
			return f, mockError
		}

		err := batchNewVersion(fs, testConfig, batchData{
			Version:        semver.MustParse("v0.1.0"),
			ChangesPerKind: map[string][]Change{},
		})
		Expect(err).To(Equal(mockError))
	})

	It("returns error when using bad version template", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		testConfig.VersionFormat = "{{juuunk...}}"

		err = batchNewVersion(fs, testConfig, batchData{
			Version:        semver.MustParse("v0.1.0"),
			ChangesPerKind: map[string][]Change{},
		})
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad kind template", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		testConfig.KindFormat = "{{randoooom../././}}"

		err = batchNewVersion(fs, testConfig, batchData{
			Version:        semver.MustParse("v0.1.0"),
			ChangesPerKind: map[string][]Change{},
		})
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad change template", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		testConfig.ChangeFormat = "{{not.valid.syntax....}}"

		err = batchNewVersion(fs, testConfig, batchData{
			Version:        semver.MustParse("v0.1.0"),
			ChangesPerKind: map[string][]Change{},
		})
		Expect(err).NotTo(BeNil())
	})

	templateTests := []struct {
		key    string
		prefix string
	}{
		{key: "version", prefix: "## "},
		{key: "kind", prefix: "### "},
		{key: "change", prefix: "* "},
	}

	for _, test := range templateTests {
		prefix := test.prefix
		It(fmt.Sprintf("returns error when failing to execute %s template", test.key), func() {
			err := afs.MkdirAll("news", 0644)
			Expect(err).To(BeNil())

			vFile := newMockFile(fs, "v0.1.0.md")

			fs.mockCreate = func(path string) (afero.File, error) {
				return vFile, nil
			}

			vFile.mockWrite = func(data []byte) (int, error) {
				if strings.HasPrefix(string(data), prefix) {
					return len(data), mockError
				}
				return vFile.memFile.Write(data)
			}

			changesPerKind := map[string][]Change{
				"added":   {{Body: "w"}, {Body: "x"}},
				"removed": {{Body: "y"}, {Body: "z"}},
			}

			err = batchNewVersion(fs, testConfig, batchData{
				Version:        semver.MustParse("v0.1.0"),
				ChangesPerKind: changesPerKind,
			})
			Expect(err).To(Equal(mockError))
		})
	}

	It("delete unreleased removes unreleased files including header", func() {
		futurePath := filepath.Join("news", "future")
		err := afs.MkdirAll(futurePath, 0644)
		Expect(err).To(BeNil())

		for _, name := range []string{"a.yaml", "b.yaml", "c.yaml", ".gitkeep", "header.md"} {
			var f afero.File
			f, err = afs.Create(filepath.Join(futurePath, name))
			Expect(err).To(BeNil())

			err = f.Close()
			Expect(err).To(BeNil())
		}

		err = deleteUnreleased(afs, testConfig, "header.md")
		Expect(err).To(BeNil())

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(1))
	})

	It("delete unreleased fails if remove fails", func() {
		futurePath := filepath.Join("news", "future")
		err := afs.MkdirAll(futurePath, 0644)
		Expect(err).To(BeNil())

		for _, name := range []string{"a.yaml"} {
			var f afero.File
			f, err = afs.Create(filepath.Join(futurePath, name))
			Expect(err).To(BeNil())

			err = f.Close()
			Expect(err).To(BeNil())
		}

		fs.mockRemove = func(name string) error {
			return mockError
		}

		err = deleteUnreleased(afs, testConfig, "")
		Expect(err).To(Equal(mockError))
	})
})
