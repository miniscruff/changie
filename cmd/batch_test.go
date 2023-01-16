package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

var changeIncrementer = 0

func batchCleanArgs() {
	versionHeaderPathFlag = ""
	versionFooterPathFlag = ""
	moveDirFlag = ""
	keepFragmentsFlag = false
	removePrereleasesFlag = false
	batchIncludeDirs = nil
	batchDryRunFlag = false
	batchPrereleaseFlag = nil
	batchMetaFlag = nil
	batchForce = false
}

func batchTestConfig() *core.Config {
	return &core.Config{
		ChangesDir:         "news",
		UnreleasedDir:      "future",
		HeaderPath:         "",
		ChangelogPath:      "news.md",
		VersionExt:         "md",
		VersionFormat:      "## {{.Version}}",
		KindFormat:         "### {{.Kind}}",
		ChangeFormat:       "* {{.Body}}",
		FragmentFileFormat: "",
		Kinds: []core.KindConfig{
			{Label: "added"},
			{Label: "removed"},
			{Label: "other"},
		},
	}
}

func withPipelines(afs afero.Afero) (*standardBatchPipeline, *MockBatchPipeline) {
	standard := &standardBatchPipeline{afs: afs, templateCache: core.NewTemplateCache()}
	mockPipeline := &MockBatchPipeline{standard: standard}

	return standard, mockPipeline
}

// writeChangeFile will write a change file with an auto-incrementing index to prevent
// same second clobbering
func writeChangeFile(t *testing.T, afs afero.Afero, cfg *core.Config, change core.Change) {
	// set our time as an arbitrary amount from jan 1 2000 so
	// each change is 1 hour later then the last
	if change.Time.Year() == 0 {
		change.Time = time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)
		change.Time.Add(time.Duration(changeIncrementer) * time.Hour)
	}

	bs, _ := yaml.Marshal(&change)
	name := fmt.Sprintf("change-%d.yaml", changeIncrementer)
	changeIncrementer++

	then.WriteFile(t, afs, bs, cfg.ChangesDir, cfg.UnreleasedDir, name)
}

func TestBatchCanBatch(t *testing.T) {
	cfg := batchTestConfig()
	// declared path but missing is accepted
	cfg.VersionHeaderPath = "header.md"

	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

	err := batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	verContents := `## v0.2.0
### added
* A
* B
### removed
* C`

	then.FileContents(t, afs, verContents, cfg.ChangesDir, "v0.2.0.md")

	infos, err := afs.ReadDir(filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir))
	then.Nil(t, err)
	then.Equals(t, 0, len(infos))
}

func TestBatchCanAddNewLinesBeforeAndAfterKindHeader(t *testing.T) {
	cfg := batchTestConfig()
	// declared path but missing is accepted
	cfg.VersionHeaderPath = "header.md"
	cfg.Newlines.BeforeKind = 2
	cfg.Newlines.AfterKind = 2

	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

	err := batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	verContents := `## v0.2.0


### added


* A
* B


### removed


* C`

	then.FileContents(t, afs, verContents, cfg.ChangesDir, "v0.2.0.md")

	infos, err := afs.ReadDir(filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir))
	then.Nil(t, err)
	then.Equals(t, 0, len(infos))
}

func TestBatchErrorStandardBadWriter(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)
	w := then.NewErrWriter()

	err := standard.WriteTemplate(
		w,
		cfg.VersionFormat,
		2, // tries to write some before lines and fails
		cfg.Newlines.AfterVersion,
		"v0.2.0",
	)
	w.Raised(t, err)
}

func TestBatchComplexVersion(t *testing.T) {
	batchIncludeDirs = []string{"beta"}
	batchPrereleaseFlag = []string{"b1"}
	batchMetaFlag = []string{"hash"}

	t.Cleanup(batchCleanArgs)
	t.Setenv("TEST_CHANGIE_ENV_KEY", "env value")

	cfg := batchTestConfig()
	cfg.EnvPrefix = "TEST_CHANGIE_"
	cfg.HeaderFormat = "{{ bodies .Changes | len }} changes this release"
	cfg.ChangeFormat = "* {{.Body}} by {{.Custom.Author}}"
	cfg.FooterFormat = `### contributors
{{- range (customs .Changes "Author" | uniq) }}
* [{{.}}](https://github.com/{{.}})
{{- end}}
env: {{.Env.ENV_KEY}}`

	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(
		t, afs, cfg,
		core.Change{
			Kind: "added",
			Body: "D",
			Custom: map[string]string{
				"Author": "miniscruff",
			},
		},
	)

	betaChange := core.Change{
		Kind: "added",
		Body: "E",
		Custom: map[string]string{
			"Author": "otherAuthor",
		},
		Time: time.Now(), // make sure our beta change is more recent then our other changes
	}

	betaPath := filepath.Join(cfg.ChangesDir, "beta")
	then.Nil(t, afs.MkdirAll(betaPath, 0644))

	betaChangeWriter, err := afs.Create(filepath.Join(betaPath, "change.yaml"))
	then.Nil(t, err)

	err = betaChange.Write(betaChangeWriter)
	then.Nil(t, err)

	err = batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	verContents := `## v0.2.0-b1+hash
2 changes this release
### added
* D by miniscruff
* E by otherAuthor
### contributors
* [miniscruff](https://github.com/miniscruff)
* [otherAuthor](https://github.com/otherAuthor)
env: env value`
	then.FileContents(t, afs, verContents, cfg.ChangesDir, "v0.2.0-b1+hash.md")

	_, err = afs.ReadDir(filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir))
	then.Nil(t, err)

	_, err = afs.Stat(betaPath)
	then.Err(t, os.ErrNotExist, err)
}

func TestBatchDryRun(t *testing.T) {
	var builder strings.Builder
	batchDryRunOut = &builder
	batchDryRunFlag = true

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "D"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "E"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "F"})

	err := batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	verContents := `## v0.2.0
### added
* D
* E
### removed
* F`
	then.Equals(t, verContents, builder.String())

	infos, err := afs.ReadDir(filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir))
	then.Nil(t, err)
	then.Equals(t, 3, len(infos))
}

func TestBatchRemovePrereleases(t *testing.T) {
	removePrereleasesFlag = true

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.2-a1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.2-a2.md")

	err := batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	infos, err := afs.ReadDir(cfg.ChangesDir)
	then.Nil(t, err)
	then.Equals(t, cfg.UnreleasedDir, infos[0].Name())
	then.Equals(t, "v0.2.0.md", infos[1].Name())
	then.Equals(t, 2, len(infos))
}

func TestBatchKeepChangeFiles(t *testing.T) {
	keepFragmentsFlag = true

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

	err := batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	infos, err := afs.ReadDir(filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir))
	then.Nil(t, err)
	then.Equals(t, 3, len(infos))
}

func TestBatchWithHeadersAndFooters(t *testing.T) {
	versionHeaderPathFlag = "h1.md"
	versionFooterPathFlag = "f1.md"

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	cfg.VersionHeaderPath = "h2.md"
	cfg.VersionFooterPath = "f2.md"
	cfg.HeaderFormat = "header format"
	cfg.FooterFormat = "footer format"
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

	then.WriteFile(t, afs, []byte("first header\n"), cfg.ChangesDir, cfg.UnreleasedDir, "h1.md")
	then.WriteFile(t, afs, []byte("second header\n"), cfg.ChangesDir, cfg.UnreleasedDir, "h2.md")
	then.WriteFile(t, afs, []byte("first footer\n"), cfg.ChangesDir, cfg.UnreleasedDir, "f1.md")
	then.WriteFile(t, afs, []byte("second footer\n"), cfg.ChangesDir, cfg.UnreleasedDir, "f2.md")

	err := batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	verContents := `## v0.2.0
first header

second header

header format
### added
* A
* B
### removed
* C
footer format
first footer

second footer
`
	then.FileContents(t, afs, verContents, cfg.ChangesDir, "v0.2.0.md")
}

func TestBatchErrorBadChanges(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	mockError := errors.New("bad get changes")
	mockPipeline.MockGetChanges = func(cfg core.Config, search []string) ([]core.Change, error) {
		return nil, mockError
	}

	aVer := []byte("not a valid change")
	then.WriteFile(t, afs, aVer, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	err := batchPipeline(mockPipeline, afs, "v0.1.1")
	then.Err(t, mockError, err)
}

/*
var _ = Describe("Batch", func() {
	var (
		stdout          *os.File
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
		stdout = os.Stdout

		futurePath = filepath.Join("news", "future")
		newVerPath = filepath.Join("news", "v0.2.0.md")
		fileCreateIndex = 0
	})

	writeFutureFile := func(header, configPath string) {
		headerData := []byte(header)
		headerPath := filepath.Join(futurePath, configPath)
		Expect(afs.WriteFile(headerPath, headerData, os.ModePerm)).To(Succeed())
	}

	It("can override version if batch is forced", func() {
		batchForce = true
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
		err := batchPipeline(standard, afs, "v0.2.0")
		Expect(err).To(BeNil())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
		err = batchPipeline(standard, afs, "v0.2.0")
		Expect(err).To(BeNil())

		verContents := `## v0.2.0
### added
* B`
		Expect(newVerPath).To(HaveContents(afs, verContents))
	})

	It("returns error if batch already exists", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
		err := batchPipeline(standard, afs, "v0.2.0")
		Expect(err).To(BeNil())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
		err = batchPipeline(standard, afs, "v0.2.0")
		Expect(err).To(MatchError(errVersionExists))
	})

	It("returns error on bad semver", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

		err := batchPipeline(standard, afs, "---asdfasdf---")
		Expect(err).To(Equal(core.ErrBadVersionOrPart))
	})

	It("returns error trying to get all versions with removing prereleases", func() {
		removePrereleasesFlag = true
		testConfig.VersionExt = "txt"
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
		writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

		prePath := filepath.Join("news", "v0.1.2-a1.txt")
		Expect(afs.WriteFile(prePath, []byte{}, 0644)).To(Succeed())

		fs.MockRemove = func(name string) error {
			if name == prePath {
				return mockError
			}

			return fs.MemFS.Remove(name)
		}

		err := batchPipeline(standard, afs, "v0.2.0")
		Expect(err).NotTo(BeNil())
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

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "C"})

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

	It("returns error on bad version format", func() {
		testConfig.VersionFormat = "{{bad.format}"
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

		err := batchPipeline(standard, afs, "v0.11.0")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad version header write", func() {
		versionHeaderPathFlag = "h1.md"

		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

		mockPipeline.MockWriteFile = func(
			writer io.Writer,
			config core.Config,
			beforeNewlines int,
			afterNewlines int,
			relativePath string,
			templateData interface{},
		) error {
			if strings.HasSuffix(relativePath, versionHeaderPathFlag) {
				return mockError
			}
			return mockPipeline.standard.WriteFile(
				writer,
				config,
				beforeNewlines,
				afterNewlines,
				relativePath,
				templateData,
			)
		}

		err := batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad version header write", func() {
		testConfig.VersionHeaderPath = "h2.md"

		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

		mockPipeline.MockWriteFile = func(
			writer io.Writer,
			config core.Config,
			beforeNewlines int,
			afterNewlines int,
			relativePath string,
			templateData interface{},
		) error {
			if strings.HasSuffix(relativePath, testConfig.VersionHeaderPath) {
				return mockError
			}
			return mockPipeline.standard.WriteFile(
				writer,
				config,
				beforeNewlines,
				afterNewlines,
				relativePath,
				templateData,
			)
		}

		err := batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad version footer write", func() {
		versionFooterPathFlag = "f1.md"

		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

		mockPipeline.MockWriteFile = func(
			writer io.Writer,
			config core.Config,
			beforeNewlines int,
			afterNewlines int,
			relativePath string,
			templateData interface{},
		) error {
			if strings.HasSuffix(relativePath, versionFooterPathFlag) {
				return mockError
			}
			return mockPipeline.standard.WriteFile(
				writer,
				config,
				beforeNewlines,
				afterNewlines,
				relativePath,
				templateData,
			)
		}

		err := batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad version header write", func() {
		testConfig.VersionFooterPath = "f2.md"

		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

		mockPipeline.MockWriteFile = func(
			writer io.Writer,
			config core.Config,
			beforeNewlines int,
			afterNewlines int,
			relativePath string,
			templateData interface{},
		) error {
			if strings.HasSuffix(relativePath, testConfig.VersionFooterPath) {
				return mockError
			}
			return mockPipeline.standard.WriteFile(
				writer,
				config,
				beforeNewlines,
				afterNewlines,
				relativePath,
				templateData,
			)
		}

		err := batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad header format", func() {
		testConfig.HeaderFormat = "bad format {{...{{"

		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

		err := batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad footer format", func() {
		testConfig.FooterFormat = "bad format {{...{{"

		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

		err := batchPipeline(mockPipeline, afs, "v0.1.1")
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad delete", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "C"})

		mockPipeline.MockClearUnreleased = func(
			config core.Config,
			moveDir string,
			includeDirs []string,
			otherFiles ...string,
		) error {
			return mockError
		}

		err := batchPipeline(mockPipeline, afs, "v0.2.3")
		Expect(err).To(Equal(mockError))
	})

	It("returns error on bad write", func() {
		Expect(testConfig.Save(afs.WriteFile)).To(Succeed())

		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "C"})
		mockPipeline.MockWriteChanges = func(
			writer io.Writer,
			cfg core.Config,
			changes []core.Change,
		) error {
			return mockError
		}

		err := batchPipeline(mockPipeline, afs, "v0.2.3")
		Expect(err).To(Equal(mockError))
	})

	It("can get all changes", func() {
		writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "third", Time: orderedTimes[2]})
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "first", Time: orderedTimes[0]})
		writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "second", Time: orderedTimes[1]})

		ignoredPath := filepath.Join(futurePath, "ignored.txt")
		Expect(afs.WriteFile(ignoredPath, []byte("ignored"), os.ModePerm)).To(Succeed())

		changes, err := standard.GetChanges(testConfig, nil)
		Expect(err).To(BeNil())
		Expect(changes[0].Body).To(Equal("second"))
		Expect(changes[1].Body).To(Equal("first"))
		Expect(changes[2].Body).To(Equal("third"))
	})

	It("returns err if unable to read directory", func() {
		fs.MockOpen = func(path string) (afero.File, error) {
			var f afero.File
			return f, mockError
		}

		_, err := standard.GetChanges(testConfig, nil)
		Expect(err).To(Equal(mockError))
	})

	It("returns err if bad changes file", func() {
		badYaml := []byte("not a valid yaml:::::file---___")
		err := afs.WriteFile(filepath.Join(futurePath, "a.yaml"), badYaml, os.ModePerm)
		Expect(err).To(BeNil())

		_, err = standard.GetChanges(testConfig, nil)
		Expect(err).NotTo(BeNil())
	})

	It("can write unreleased file", func() {
		var builder strings.Builder
		text := []byte("some text")
		err := afs.WriteFile(filepath.Join(futurePath, "a.md"), text, os.ModePerm)
		Expect(err).To(BeNil())

		err = standard.WriteFile(&builder, testConfig, 1, 0, "a.md", nil)
		Expect(builder.String()).To(Equal("\nsome text"))
		Expect(err).To(BeNil())
	})

	It("skips writing unreleased if path is empty", func() {
		var builder strings.Builder
		text := []byte("some text")
		err := afs.WriteFile(filepath.Join(futurePath, "a.md"), text, os.ModePerm)
		Expect(err).To(BeNil())

		err = standard.WriteFile(&builder, testConfig, 0, 0, "", nil)
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

		err := standard.WriteFile(&builder, testConfig, 0, 0, "a.md", nil)
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

		err = standard.WriteFile(badFile, testConfig, 0, 0, "a.md", nil)
		Expect(err).To(Equal(mockError))
	})

	It("skips writing if files does not exist", func() {
		var builder strings.Builder

		err := standard.WriteFile(&builder, testConfig, 0, 0, "a.md", nil)
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

		err := standard.WriteChanges(&writer, testConfig, changes)
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
		err := standard.WriteChanges(&writer, testConfig, changes)
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
		err := standard.WriteChanges(&writer, testConfig, changes)
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

		err := standard.WriteChanges(&writer, testConfig, changes)
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad component template", func() {
		var writer strings.Builder
		testConfig.ComponentFormat = "{{deja vu}}"
		changes := []core.Change{
			{Component: "x", Kind: "added"},
		}

		err := standard.WriteChanges(&writer, testConfig, changes)
		Expect(err).NotTo(BeNil())
	})

	It("returns error when using bad change template", func() {
		var writer strings.Builder
		testConfig.ChangeFormat = "{{not.valid.syntax....}}"
		changes := []core.Change{
			{Body: "x", Kind: "added"},
		}

		err := standard.WriteChanges(&writer, testConfig, changes)
		Expect(err).NotTo(BeNil())
	})

	It("clear unreleased removes unreleased files including header", func() {
		var err error

		alphaPath := filepath.Join("news", "alpha")
		betaPath := filepath.Join("news", "beta")

		Expect(afs.MkdirAll(futurePath, core.CreateDirMode)).To(Succeed())
		Expect(afs.MkdirAll(alphaPath, core.CreateDirMode)).To(Succeed())
		Expect(afs.MkdirAll(betaPath, core.CreateDirMode)).To(Succeed())

		_, _ = afs.Create(filepath.Join(futurePath, "a.yaml"))
		_, _ = afs.Create(filepath.Join(futurePath, ".gitkeep"))
		_, _ = afs.Create(filepath.Join(futurePath, "header.md"))
		_, _ = afs.Create(filepath.Join(alphaPath, "b.yaml"))
		_, _ = afs.Create(filepath.Join(alphaPath, ".gitkeep"))
		_, _ = afs.Create(filepath.Join(betaPath, "c.yaml"))

		err = standard.ClearUnreleased(
			testConfig,
			"",
			[]string{"alpha", "beta"},
			"header.md",
			"",
			"does-not-exist.md",
		)
		Expect(err).To(BeNil())

		infos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(1))

		infos, err = afs.ReadDir(alphaPath)
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(1))

		_, err = afs.Stat(betaPath)
		Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
	})

	It("clear unreleased moves unreleased files including header if specified", func() {
		moveDirFlag = "beta"

		err := afs.MkdirAll(futurePath, 0644)
		Expect(err).To(BeNil())

		for _, name := range []string{"a.yaml", "b.yaml", "c.yaml", ".gitkeep", "header.md"} {
			var f afero.File
			f, err = afs.Create(filepath.Join(futurePath, name))
			Expect(err).To(BeNil())

			Expect(f.Close()).To(Succeed())
		}

		err = standard.ClearUnreleased(
			testConfig,
			moveDirFlag,
			nil,
			"header.md",
			"",
			"does-not-exist.md",
		)
		Expect(err).To(BeNil())

		// should of moved the unreleased and header file to beta
		infos, err := afs.ReadDir(filepath.Join("news", "beta"))
		Expect(err).To(BeNil())
		Expect(len(infos)).To(Equal(4))
	})

	It("clear unreleased fails if remove fails", func() {
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

		err = standard.ClearUnreleased(testConfig, "", nil)
		Expect(err).To(Equal(mockError))
	})

	It("clear unreleased fails if remove all fails", func() {
		var err error

		alphaPath := filepath.Join("news", "alpha")

		Expect(afs.MkdirAll(futurePath, core.CreateDirMode)).To(Succeed())
		Expect(afs.MkdirAll(alphaPath, core.CreateDirMode)).To(Succeed())

		_, _ = afs.Create(filepath.Join(futurePath, "a.yaml"))
		_, _ = afs.Create(filepath.Join(alphaPath, "b.yaml"))

		fs.MockRemoveAll = func(path string) error {
			return mockError
		}

		err = standard.ClearUnreleased(
			testConfig,
			"",
			[]string{"alpha"},
			"header.md",
			"",
			"does-not-exist.md",
		)
		Expect(err).To(Equal(mockError))
	})

	It("clear unreleased fails if move fails", func() {
		var err error

		for _, name := range []string{"a.yaml"} {
			var f afero.File
			f, err = afs.Create(filepath.Join(futurePath, name))
			Expect(err).To(BeNil())

			err = f.Close()
			Expect(err).To(BeNil())
		}

		fs.MockRename = func(before, after string) error {
			return mockError
		}

		err = standard.ClearUnreleased(testConfig, "beta", nil)
		Expect(err).To(Equal(mockError))
	})

	It("clear unreleased fails if unable to make move dir", func() {
		var err error

		for _, name := range []string{"a.yaml"} {
			var f afero.File
			f, err = afs.Create(filepath.Join(futurePath, name))
			Expect(err).To(BeNil())

			err = f.Close()
			Expect(err).To(BeNil())
		}

		fs.MockMkdirAll = func(p string, mode os.FileMode) error {
			return mockError
		}

		err = standard.ClearUnreleased(testConfig, "beta", nil)
		Expect(err).To(Equal(mockError))
	})

	It("clear unreleased fails if unable to find files", func() {
		err := standard.ClearUnreleased(testConfig, "beta", []string{"../../bad"})
		Expect(err).NotTo(BeNil())
	})
})
*/

type MockBatchPipeline struct {
	MockGetChanges    func(config core.Config, searchPaths []string) ([]core.Change, error)
	MockWriteTemplate func(
		writer io.Writer,
		template string,
		beforeNewlines int,
		afterNewlines int,
		templateData interface{},
	) error
	MockWriteFile func(
		writer io.Writer,
		config core.Config,
		beforeNewlines int,
		afterNewlines int,
		relativePath string,
		templateData interface{},
	) error
	MockWriteChanges func(
		writer io.Writer,
		config core.Config,
		changes []core.Change,
	) error
	MockClearUnreleased func(
		config core.Config,
		moveDir string,
		includeDirs []string,
		otherFiles ...string,
	) error
	standard *standardBatchPipeline
}

func (m *MockBatchPipeline) GetChanges(
	config core.Config,
	searchPaths []string,
) ([]core.Change, error) {
	if m.MockGetChanges != nil {
		return m.MockGetChanges(config, searchPaths)
	}

	return m.standard.GetChanges(config, searchPaths)
}

func (m *MockBatchPipeline) WriteTemplate(
	writer io.Writer,
	template string,
	beforeNewlines int,
	afterNewlines int,
	templateData interface{},
) error {
	if m.MockWriteTemplate != nil {
		return m.MockWriteTemplate(writer, template, beforeNewlines, afterNewlines, templateData)
	}

	return m.standard.WriteTemplate(writer, template, beforeNewlines, afterNewlines, templateData)
}

func (m *MockBatchPipeline) WriteFile(
	writer io.Writer,
	config core.Config,
	beforeNewlines int,
	afterNewlines int,
	relativePath string,
	templateData interface{},
) error {
	if m.MockWriteFile != nil {
		return m.MockWriteFile(writer, config, beforeNewlines, afterNewlines, relativePath, templateData)
	}

	return m.standard.WriteFile(
		writer,
		config,
		beforeNewlines,
		afterNewlines,
		relativePath,
		templateData,
	)
}

func (m *MockBatchPipeline) WriteChanges(
	writer io.Writer,
	config core.Config,
	changes []core.Change,
) error {
	if m.MockWriteChanges != nil {
		return m.MockWriteChanges(writer, config, changes)
	}

	return m.standard.WriteChanges(writer, config, changes)
}

func (m *MockBatchPipeline) ClearUnreleased(
	config core.Config,
	moveDir string,
	includeDirs []string,
	otherFiles ...string,
) error {
	if m.MockClearUnreleased != nil {
		return m.MockClearUnreleased(config, moveDir, includeDirs, otherFiles...)
	}

	return m.standard.ClearUnreleased(config, moveDir, includeDirs, otherFiles...)
}
