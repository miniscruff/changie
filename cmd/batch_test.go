package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/miniscruff/changie/core"
	. "github.com/miniscruff/changie/testutils"
)

var _ = Describe("Batch", func() {

	var (
		fs         *MockFS
		afs        afero.Afero
		mockError  error
		testConfig core.Config

		orderedTimes = []time.Time{
			time.Date(2019, 5, 25, 20, 45, 0, 0, time.UTC),
			time.Date(2017, 4, 25, 15, 20, 0, 0, time.UTC),
			time.Date(2015, 3, 25, 10, 5, 0, 0, time.UTC),
		}

		formatTime = func(t time.Time) string {
			return t.Format(time.RFC3339Nano)
		}
	)

	BeforeEach(func() {
		fs = NewMockFS()
		afs = afero.Afero{Fs: fs}
		mockError = errors.New("dummy mock error")
		testConfig = core.Config{
			ChangesDir:    "news",
			UnreleasedDir: "future",
			HeaderPath:    "header.rst",
			ChangelogPath: "news.md",
			VersionExt:    "md",
			VersionFormat: "## {{.Version}}",
			KindFormat:    "\n### {{.Kind}}",
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
		testConfig.VersionFormat = testConfig.VersionFormat + "\n"
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
		testConfig.VersionFormat = testConfig.VersionFormat + "\n"
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
		Expect(err).To(Equal(core.ErrBadVersionOrPart))
	})

	It("returns error on bad config", func() {
		configData := []byte("not a proper config")
		err := afs.WriteFile(core.ConfigPath, configData, os.ModePerm)
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

		fs.MockRemove = func(name string) error {
			return mockError
		}

		err = batchPipeline(afs, "v1.0.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad header file", func() {
		versionHeaderPath = "header.md"
		badOpen := errors.New("bad open file")
		fs.MockOpen = func(name string) (afero.File, error) {
			if strings.HasSuffix(name, versionHeaderPath) {
				return nil, badOpen
			}
			return fs.MemFS.Open(name)
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

		aVer := []byte(fmt.Sprintf(
			"kind: removed\nbody: third\ntime: %s",
			formatTime(orderedTimes[2]),
		))
		err := afs.WriteFile(filepath.Join(unrelPath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		bVer := []byte(fmt.Sprintf(
			"kind: added\nbody: first\ntime: %s",
			formatTime(orderedTimes[0]),
		))
		err = afs.WriteFile(filepath.Join(unrelPath, "b.yaml"), bVer, os.ModePerm)
		Expect(err).To(BeNil())

		cVer := []byte(fmt.Sprintf(
			"kind: added\nbody: second\ntime: %s",
			formatTime(orderedTimes[1]),
		))
		err = afs.WriteFile(filepath.Join(unrelPath, "c.yaml"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = afs.WriteFile(filepath.Join(unrelPath, "ignored.txt"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		changes, err := getChanges(afs, testConfig)
		Expect(err).To(BeNil())
		Expect(changes[0].Body).To(Equal("first"))
		Expect(changes[1].Body).To(Equal("second"))
		Expect(changes[2].Body).To(Equal("third"))
	})

	It("can get all changes with components", func() {
		unrelPath := filepath.Join("news", "future")

		aVer := []byte(fmt.Sprintf(
			"component: compiler\nkind: removed\nbody: A\ntime: %s",
			formatTime(orderedTimes[2]),
		))
		err := afs.WriteFile(filepath.Join(unrelPath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		bVer := []byte(fmt.Sprintf(
			"component: linker\nkind: added\nbody: B\ntime: %s",
			formatTime(orderedTimes[0]),
		))
		err = afs.WriteFile(filepath.Join(unrelPath, "b.yaml"), bVer, os.ModePerm)
		Expect(err).To(BeNil())

		cVer := []byte(fmt.Sprintf(
			"component: linker\nkind: added\nbody: C\ntime: %s",
			formatTime(orderedTimes[1]),
		))
		err = afs.WriteFile(filepath.Join(unrelPath, "c.yaml"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = afs.WriteFile(filepath.Join(unrelPath, "ignored.txt"), cVer, os.ModePerm)
		Expect(err).To(BeNil())

		testConfig.Components = []string{"linker", "compiler"}
		changes, err := getChanges(afs, testConfig)
		Expect(err).To(BeNil())
		Expect(changes[0].Body).To(Equal("B"))
		Expect(changes[1].Body).To(Equal("C"))
		Expect(changes[2].Body).To(Equal("A"))
	})

	It("returns err if unable to read directory", func() {
		fs.MockOpen = func(path string) (afero.File, error) {
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

		vFile := NewMockFile(fs, "v0.1.0.md")
		contents := ""

		fs.MockCreate = func(path string) (afero.File, error) {
			return vFile, nil
		}

		vFile.MockWrite = func(data []byte) (int, error) {
			contents += string(data)
			return len(data), nil
		}
		vFile.MockWriteString = func(data string) (int, error) {
			contents += data
			return len(data), nil
		}

		changes := []core.Change{
			{Kind: "added", Body: "w"},
			{Kind: "added", Body: "x"},
			{Kind: "removed", Body: "y"},
			{Kind: "removed", Body: "z"},
		}

		err = batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: changes,
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

	It("can create new version file without kind headers", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		vFile := NewMockFile(fs, "v0.1.0.md")
		contents := ""

		fs.MockCreate = func(path string) (afero.File, error) {
			return vFile, nil
		}

		vFile.MockWrite = func(data []byte) (int, error) {
			contents += string(data)
			return len(data), nil
		}
		vFile.MockWriteString = func(data string) (int, error) {
			contents += data
			return len(data), nil
		}

		changes := []core.Change{
			{Body: "w", Kind: "added"},
			{Body: "x", Kind: "added"},
			{Body: "y", Kind: "removed"},
			{Body: "z", Kind: "removed"},
		}

		testConfig.KindFormat = ""
		testConfig.ChangeFormat = "* {{.Body}} ({{.Kind}})"
		err = batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: changes,
		})
		Expect(err).To(BeNil())

		expected := `## v0.1.0
* w (added)
* x (added)
* y (removed)
* z (removed)`
		Expect(contents).To(Equal(expected))
	})

	It("can create new version file with component headers", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		vFile := NewMockFile(fs, "v0.1.0.md")
		contents := ""

		fs.MockCreate = func(path string) (afero.File, error) {
			return vFile, nil
		}

		vFile.MockWrite = func(data []byte) (int, error) {
			contents += string(data)
			return len(data), nil
		}
		vFile.MockWriteString = func(data string) (int, error) {
			contents += data
			return len(data), nil
		}

		changes := []core.Change{
			{Body: "w", Kind: "added", Component: "linker"},
			{Body: "x", Kind: "added", Component: "linker"},
			{Body: "y", Kind: "removed", Component: "linker"},
			{Body: "z", Kind: "removed", Component: "compiler"},
		}

		testConfig.Components = []string{"linker", "compiler"}
		testConfig.ComponentFormat = "\n## {{.Component}}"
		testConfig.KindFormat = "### {{.Kind}}"
		testConfig.ChangeFormat = "* {{.Body}}"
		err = batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: changes,
		})
		Expect(err).To(BeNil())

		expected := `## v0.1.0

## linker
### added
* w
* x
### removed
* y

## compiler
### removed
* z`
		Expect(contents).To(Equal(expected))
	})

	It("can create new version file with header", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		vFile := NewMockFile(fs, "v0.1.0.md")
		contents := ""

		fs.MockCreate = func(path string) (afero.File, error) {
			return vFile, nil
		}

		vFile.MockWrite = func(data []byte) (int, error) {
			contents += string(data)
			return len(data), nil
		}
		vFile.MockWriteString = func(data string) (int, error) {
			contents += data
			return len(data), nil
		}

		changes := []core.Change{
			{Body: "w", Kind: "added"},
			{Body: "x", Kind: "added"},
			{Body: "y", Kind: "removed"},
			{Body: "z", Kind: "removed"},
		}

		testConfig.VersionFormat = testConfig.VersionFormat + "\n"
		err = batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: changes,
			Header:  "Some header we want included in our new version.\nCan also have newlines.",
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
		fs.MockCreate = func(path string) (afero.File, error) {
			var f afero.File
			return f, mockError
		}

		err := batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: []core.Change{},
		})
		Expect(err).To(Equal(mockError))
	})

	It("returns error when using bad version template", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		testConfig.VersionFormat = "{{juuunk...}}"

		err = batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: []core.Change{},
		})
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad kind template", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		testConfig.KindFormat = "{{randoooom../././}}"

		err = batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: []core.Change{},
		})
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad component template", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		testConfig.ComponentFormat = "{{deja vu}}"

		err = batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: []core.Change{},
		})
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad change template", func() {
		err := afs.MkdirAll("news", 0644)
		Expect(err).To(BeNil())

		testConfig.ChangeFormat = "{{not.valid.syntax....}}"

		err = batchNewVersion(fs, testConfig, batchData{
			Version: semver.MustParse("v0.1.0"),
			Changes: []core.Change{},
		})
		Expect(err).NotTo(BeNil())
	})

	templateTests := []struct {
		key    string
		prefix string
	}{
		{key: "version", prefix: "## "},
		{key: "component", prefix: "\n## "},
		{key: "kind", prefix: "\n### "},
		{key: "change", prefix: "* "},
	}

	for _, test := range templateTests {
		prefix := test.prefix
		It(fmt.Sprintf("returns error when failing to execute %s template", test.key), func() {
			err := afs.MkdirAll("news", 0644)
			Expect(err).To(BeNil())

			vFile := NewMockFile(fs, "v0.1.0.md")

			fs.MockCreate = func(path string) (afero.File, error) {
				return vFile, nil
			}

			vFile.MockWrite = func(data []byte) (int, error) {
				if strings.HasPrefix(string(data), prefix) {
					return len(data), mockError
				}
				return vFile.MemFile.Write(data)
			}

			changes := []core.Change{
				{Component: "A", Kind: "added", Body: "w"},
				{Component: "A", Kind: "added", Body: "x"},
				{Component: "A", Kind: "removed", Body: "y"},
				{Component: "B", Kind: "removed", Body: "z"},
			}

			testConfig.ComponentFormat = "\n## {{.Component}}"
			testConfig.Components = []string{"A", "B"}
			err = batchNewVersion(fs, testConfig, batchData{
				Version: semver.MustParse("v0.1.0"),
				Changes: changes,
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

		fs.MockRemove = func(name string) error {
			return mockError
		}

		err = deleteUnreleased(afs, testConfig, "")
		Expect(err).To(Equal(mockError))
	})
})
