package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"

	"github.com/miniscruff/changie/core"
	. "github.com/miniscruff/changie/testutils"
)

type MockBatchPipeline struct {
	MockGetChanges          func(config core.Config) ([]core.Change, error)
	MockWriteUnreleasedFile func(writer io.Writer, config core.Config, relativePath string) error
	MockWriteChanges        func(
		writer io.Writer,
		config core.Config,
		templateCache *core.TemplateCache,
		changes []core.Change,
	) error
	MockDeleteUnreleased func(config core.Config, otherFiles ...string) error
	standard             *standardBatchPipeline
}

func (m *MockBatchPipeline) GetChanges(config core.Config) ([]core.Change, error) {
	if m.MockGetChanges != nil {
		return m.MockGetChanges(config)
	}

	return m.standard.GetChanges(config)
}

func (m *MockBatchPipeline) WriteUnreleasedFile(
	writer io.Writer,
	config core.Config,
	relativePath string,
) error {
	if m.MockWriteUnreleasedFile != nil {
		return m.MockWriteUnreleasedFile(writer, config, relativePath)
	}

	return m.standard.WriteUnreleasedFile(writer, config, relativePath)
}

func (m *MockBatchPipeline) WriteChanges(
	writer io.Writer,
	config core.Config,
	templateCache *core.TemplateCache,
	changes []core.Change,
) error {
	if m.MockWriteChanges != nil {
		return m.MockWriteChanges(writer, config, templateCache, changes)
	}

	return m.standard.WriteChanges(writer, config, templateCache, changes)
}

func (m *MockBatchPipeline) DeleteUnreleased(config core.Config, otherFiles ...string) error {
	if m.MockDeleteUnreleased != nil {
		return m.MockDeleteUnreleased(config, otherFiles...)
	}

	return m.standard.DeleteUnreleased(config, otherFiles...)
}

var _ = Describe("Batch", func() {

	var (
		fs              *MockFS
		afs             afero.Afero
		standard        *standardBatchPipeline
		mockPipeline    *MockBatchPipeline
		stdout          *os.File
		mockError       error
		testConfig      core.Config
		fileCreateIndex int
		futurePath      string
		newVerPath      string

		orderedTimes = []time.Time{
			time.Date(2019, 5, 25, 20, 45, 0, 0, time.UTC),
			time.Date(2017, 4, 25, 15, 20, 0, 0, time.UTC),
			time.Date(2015, 3, 25, 10, 5, 0, 0, time.UTC),
		}
	)

	BeforeEach(func() {
		fs = NewMockFS()
		afs = afero.Afero{Fs: fs}
		standard = &standardBatchPipeline{afs: afs}
		mockPipeline = &MockBatchPipeline{standard: standard}
		stdout = os.Stdout
		mockError = errors.New("dummy mock error")
		testConfig = core.Config{
			ChangesDir:    "news",
			UnreleasedDir: "future",
			HeaderPath:    "",
			ChangelogPath: "news.md",
			VersionExt:    "md",
			VersionFormat: "## {{.Version}}",
			KindFormat:    "### {{.Kind}}",
			ChangeFormat:  "* {{.Body}}",
			Kinds: []core.KindConfig{
				{Label: "added"},
				{Label: "removed"},
				{Label: "other"},
			},
		}

		futurePath = filepath.Join("news", "future")
		newVerPath = filepath.Join("news", "v0.2.0.md")
		fileCreateIndex = 0
	})

	AfterEach(func() {
		// reset all our flag vars
		versionHeaderPathFlag = ""
		keepFragmentsFlag = false
		batchDryRunFlag = false
		batchDryRunOut = stdout
	})

	// this mimics the change.SaveUnreleased but prevents clobbering in same
	// second saves
	writeChangeFile := func(change core.Change) {
		bs, _ := yaml.Marshal(&change)
		nameParts := make([]string, 0)

		if change.Component != "" {
			nameParts = append(nameParts, change.Component)
		}

		if change.Kind != "" {
			nameParts = append(nameParts, change.Kind)
		}

		// add an index to prevent clobbering
		nameParts = append(nameParts, fmt.Sprintf("%v", fileCreateIndex))
		fileCreateIndex++

		filePath := fmt.Sprintf(
			"%s/%s/%s.yaml",
			testConfig.ChangesDir,
			testConfig.UnreleasedDir,
			strings.Join(nameParts, "-"),
		)

		Expect(afs.WriteFile(filePath, bs, os.ModePerm)).To(Succeed())
	}

	writeFutureFile := func(header, configPath string) {
		headerData := []byte(header)
		headerPath := filepath.Join(futurePath, configPath)
		Expect(afs.WriteFile(headerPath, headerData, os.ModePerm)).To(Succeed())
	}

	It("can batch version", func() {
		// declared path but missing is accepted
		testConfig.VersionHeaderPath = "header.md"
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "added", Body: "A"})
		writeChangeFile(core.Change{Kind: "added", Body: "B"})
		writeChangeFile(core.Change{Kind: "removed", Body: "C"})

		err := batchPipeline(standard, afs, "v0.2.0")
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

	It("can batch a dry run version", func() {
		var builder strings.Builder
		batchDryRunOut = &builder
		batchDryRunFlag = true
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "added", Body: "A"})
		writeChangeFile(core.Change{Kind: "added", Body: "B"})
		writeChangeFile(core.Change{Kind: "removed", Body: "C"})

		err := batchPipeline(standard, afs, "v0.2.0")
		Expect(err).To(BeNil())

		verContents := `## v0.2.0
### added
* A
* B
### removed
* C`

		Expect(builder.String()).To(Equal(verContents))

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(3))
	})

	It("can batch version keeping change files", func() {
		// testBatchConfig.keepFragments = true
		keepFragmentsFlag = true
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "added", Body: "A"})
		writeChangeFile(core.Change{Kind: "added", Body: "B"})
		writeChangeFile(core.Change{Kind: "removed", Body: "C"})

		err := batchPipeline(standard, afs, "v0.2.0")
		Expect(err).To(BeNil())

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(3))
	})

	It("can batch version with header and footer files", func() {
		versionHeaderPathFlag = "h1.md"
		testConfig.VersionHeaderPath = "h2.md"
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "added", Body: "A"})
		writeChangeFile(core.Change{Kind: "added", Body: "B"})
		writeChangeFile(core.Change{Kind: "removed", Body: "C"})
		writeFutureFile("first header\n", "h1.md")
		writeFutureFile("second header\n", "h2.md")

		err := batchPipeline(standard, afs, "v0.2.0")
		Expect(err).To(BeNil())

		verContents := `## v0.2.0
first header

second header

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
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "added", Body: "A"})

		err := batchPipeline(standard, afs, "---asdfasdf---")
		Expect(err).To(Equal(core.ErrBadVersionOrPart))
	})

	It("returns error on bad latest version", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		// fail trying to read a bad directory
		testConfig.ChangesDir = ".../.../.../"
		err := batchPipeline(standard, afs, "v0.2.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad create", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "added", Body: "C"})

		mockFile := NewMockFile(fs, "v0.2.0.md")
		fs.MockCreate = func(name string) (afero.File, error) {
			return mockFile, mockError
		}

		err := batchPipeline(standard, afs, "v0.1.1")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad config", func() {
		configData := []byte("not a proper config")
		err := afs.WriteFile(core.ConfigPaths[0], configData, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(standard, afs, "v0.1.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad changes", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		mockPipeline.MockGetChanges = func(cfg core.Config) ([]core.Change, error) {
			return nil, mockError
		}

		aVer := []byte("not a valid change")
		err := afs.WriteFile(filepath.Join(futurePath, "a.yaml"), aVer, os.ModePerm)
		Expect(err).To(BeNil())

		err = batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad version format", func() {
		testConfig.VersionFormat = "{{bad.format}"
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "removed", Body: "C"})

		err := batchPipeline(standard, afs, "v0.11.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad version header write", func() {
		versionHeaderPathFlag = "h1.md"

		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())
		writeChangeFile(core.Change{Kind: "added", Body: "A"})

		mockPipeline.MockWriteUnreleasedFile = func(
			writer io.Writer,
			config core.Config,
			relativePath string,
		) error {
			fmt.Println(relativePath)
			if strings.HasSuffix(relativePath, versionHeaderPathFlag) {
				return mockError
			}
			return mockPipeline.standard.WriteUnreleasedFile(writer, config, relativePath)
		}

		err := batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad version header write", func() {
		testConfig.VersionHeaderPath = "h2.md"

		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())
		writeChangeFile(core.Change{Kind: "added", Body: "A"})

		mockPipeline.MockWriteUnreleasedFile = func(
			writer io.Writer,
			config core.Config,
			relativePath string,
		) error {
			if strings.HasSuffix(relativePath, testConfig.VersionHeaderPath) {
				return mockError
			}
			return mockPipeline.standard.WriteUnreleasedFile(writer, config, relativePath)
		}

		err := batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad delete", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "added", Body: "C"})

		mockPipeline.MockDeleteUnreleased = func(config core.Config, otherFiles ...string) error {
			return mockError
		}

		err := batchPipeline(mockPipeline, afs, "v0.2.3")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad write", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(core.Change{Kind: "added", Body: "C"})
		mockPipeline.MockWriteChanges = func(
			writer io.Writer,
			cfg core.Config,
			tc *core.TemplateCache,
			changes []core.Change,
		) error {
			return mockError
		}

		err := batchPipeline(mockPipeline, afs, "v0.2.3")
		Expect(err).To(Equal(mockError))
	})

	It("can get all changes", func() {
		writeChangeFile(core.Change{Kind: "removed", Body: "third", Time: orderedTimes[2]})
		writeChangeFile(core.Change{Kind: "added", Body: "first", Time: orderedTimes[0]})
		writeChangeFile(core.Change{Kind: "added", Body: "second", Time: orderedTimes[1]})

		ignoredPath := filepath.Join(futurePath, "ignored.txt")
		Expect(afs.WriteFile(ignoredPath, []byte("ignored"), os.ModePerm)).To(Succeed())

		changes, err := standard.GetChanges(testConfig)
		Expect(err).To(BeNil())
		Expect(changes[0].Body).To(Equal("first"))
		Expect(changes[1].Body).To(Equal("second"))
		Expect(changes[2].Body).To(Equal("third"))
	})

	It("can get all changes with components", func() {
		writeChangeFile(core.Change{
			Component: "compiler",
			Kind:      "removed",
			Body:      "A",
			Time:      orderedTimes[2],
		})
		writeChangeFile(core.Change{
			Component: "linker",
			Kind:      "added",
			Body:      "B",
			Time:      orderedTimes[0],
		})
		writeChangeFile(core.Change{
			Component: "linker",
			Kind:      "added",
			Body:      "C",
			Time:      orderedTimes[1],
		})

		ignoredPath := filepath.Join(futurePath, "ignored.txt")
		Expect(afs.WriteFile(ignoredPath, []byte("ignored"), os.ModePerm)).To(Succeed())

		testConfig.Components = []string{"linker", "compiler"}
		changes, err := standard.GetChanges(testConfig)
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

		_, err := standard.GetChanges(testConfig)
		Expect(err).To(Equal(mockError))
	})

	It("returns err if bad changes file", func() {
		badYaml := []byte("not a valid yaml:::::file---___")
		err := afs.WriteFile(filepath.Join(futurePath, "a.yaml"), badYaml, os.ModePerm)
		Expect(err).To(BeNil())

		_, err = standard.GetChanges(testConfig)
		Expect(err).NotTo(BeNil())
	})

	It("can write unreleased file", func() {
		var builder strings.Builder
		text := []byte("some text")
		err := afs.WriteFile(filepath.Join(futurePath, "a.md"), text, os.ModePerm)
		Expect(err).To(BeNil())

		err = standard.WriteUnreleasedFile(&builder, testConfig, "a.md")
		Expect(builder.String()).To(Equal("\nsome text"))
		Expect(err).To(BeNil())
	})

	It("skips writing unreleased if path is empty", func() {
		var builder strings.Builder
		text := []byte("some text")
		err := afs.WriteFile(filepath.Join(futurePath, "a.md"), text, os.ModePerm)
		Expect(err).To(BeNil())

		err = standard.WriteUnreleasedFile(&builder, testConfig, "")
		Expect(builder.String()).To(Equal(""))
		Expect(err).To(BeNil())
	})

	It("writing unreleased fails on bad read", func() {
		var builder strings.Builder
		badFile := NewMockFile(fs, "a.md")
		badFile.MockRead = func([]byte) (int, error) {
			return 0, mockError
		}

		fs.MockOpen = func(path string) (afero.File, error) {
			return badFile, mockError
		}

		err := standard.WriteUnreleasedFile(&builder, testConfig, "a.md")
		Expect(builder.String()).To(Equal(""))
		Expect(err).To(Equal(mockError))
	})

	It("writing unreleased fails on bad write", func() {
		badFile := NewMockFile(fs, "a.md")
		badFile.MockWrite = func([]byte) (int, error) {
			return 0, mockError
		}

		text := []byte("some text")
		err := afs.WriteFile(filepath.Join(futurePath, "a.md"), text, os.ModePerm)
		Expect(err).To(BeNil())

		err = standard.WriteUnreleasedFile(badFile, testConfig, "a.md")
		Expect(err).To(Equal(mockError))
	})

	It("skips writing if files does not exist", func() {
		var builder strings.Builder

		err := standard.WriteUnreleasedFile(&builder, testConfig, "a.md")
		Expect(builder.String()).To(Equal(""))
		Expect(err).To(BeNil())
	})

	It("can write changes", func() {
		var writer strings.Builder

		changes := []core.Change{
			{Kind: "added", Body: "w"},
			{Kind: "added", Body: "x"},
			{Kind: "removed", Body: "y"},
			{Kind: "removed", Body: "z"},
		}

		err := standard.WriteChanges(&writer, testConfig, core.NewTemplateCache(), changes)
		Expect(err).To(BeNil())

		expected := `
### added
* w
* x
### removed
* y
* z`
		Expect(writer.String()).To(Equal(expected))
	})

	It("can create new version file without kind headers", func() {
		var writer strings.Builder
		changes := []core.Change{
			{Body: "w", Kind: "added"},
			{Body: "x", Kind: "added"},
			{Body: "y", Kind: "removed"},
			{Body: "z", Kind: "removed"},
		}

		testConfig.KindFormat = ""
		testConfig.ChangeFormat = "* {{.Body}} ({{.Kind}})"
		err := standard.WriteChanges(&writer, testConfig, core.NewTemplateCache(), changes)
		Expect(err).To(BeNil())

		expected := `
* w (added)
* x (added)
* y (removed)
* z (removed)`
		Expect(writer.String()).To(Equal(expected))
	})

	It("can create new version file with component headers", func() {
		var writer strings.Builder
		changes := []core.Change{
			{Body: "w", Kind: "added", Component: "linker"},
			{Body: "x", Kind: "added", Component: "linker"},
			{Body: "y", Kind: "removed", Component: "linker"},
			{Body: "z", Kind: "removed", Component: "compiler"},
		}

		testConfig.Components = []string{"linker", "compiler"}
		testConfig.ComponentFormat = "## {{.Component}}"
		testConfig.KindFormat = "### {{.Kind}}"
		testConfig.ChangeFormat = "* {{.Body}}"
		err := standard.WriteChanges(&writer, testConfig, core.NewTemplateCache(), changes)
		Expect(err).To(BeNil())

		expected := `
## linker
### added
* w
* x
### removed
* y
## compiler
### removed
* z`
		Expect(writer.String()).To(Equal(expected))
	})

	It("returns error when using bad kind template", func() {
		var writer strings.Builder
		testConfig.KindFormat = "{{randoooom../././}}"
		changes := []core.Change{
			{Body: "x", Kind: "added"},
		}

		err := standard.WriteChanges(&writer, testConfig, core.NewTemplateCache(), changes)
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad component template", func() {
		var writer strings.Builder
		testConfig.ComponentFormat = "{{deja vu}}"
		changes := []core.Change{
			{Component: "x", Kind: "added"},
		}

		err := standard.WriteChanges(&writer, testConfig, core.NewTemplateCache(), changes)
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad change template", func() {
		var writer strings.Builder
		testConfig.ChangeFormat = "{{not.valid.syntax....}}"
		changes := []core.Change{
			{Body: "x", Kind: "added"},
		}

		err := standard.WriteChanges(&writer, testConfig, core.NewTemplateCache(), changes)
		Expect(err).NotTo(BeNil())
	})

	It("delete unreleased removes unreleased files including header", func() {
		err := afs.MkdirAll(futurePath, 0644)
		Expect(err).To(BeNil())

		for _, name := range []string{"a.yaml", "b.yaml", "c.yaml", ".gitkeep", "header.md"} {
			var f afero.File
			f, err = afs.Create(filepath.Join(futurePath, name))
			Expect(err).To(BeNil())

			Expect(f.Close()).To(Succeed())
		}

		err = standard.DeleteUnreleased(testConfig, "header.md", "", "does-not-exist.md")
		Expect(err).To(BeNil())

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(1))
	})

	It("delete unreleased fails if remove fails", func() {
		var err error

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

		err = standard.DeleteUnreleased(testConfig)
		Expect(err).To(Equal(mockError))
	})
})
