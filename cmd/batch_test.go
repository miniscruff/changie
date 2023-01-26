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

	aVer := []byte("not a valid change")
	then.WriteFile(t, afs, aVer, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	err := batchPipeline(mockPipeline, afs, "v0.1.1")
	then.NotNil(t, err)
}

func TestBatchOverrideIfForced(t *testing.T) {
	batchForce = true

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
	err := batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
	err = batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	verContents := `## v0.2.0
### added
* B`
	then.FileContents(t, afs, verContents, cfg.ChangesDir, "v0.2.0.md")
}

func TestBatchErrorIfBatchExists(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
	err := batchPipeline(standard, afs, "v0.2.0")
	then.Nil(t, err)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
	err = batchPipeline(standard, afs, "v0.2.0")
	then.Err(t, errVersionExists, err)
}

func TestBatchErrorBadVersion(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

	err := batchPipeline(standard, afs, "---asdfasdf---")
	then.Err(t, core.ErrBadVersionOrPart, err)
}

func TestBatchErrorTryingToGetAllVersionsWithRemovingPrereleases(t *testing.T) {
	removePrereleasesFlag = true

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	cfg.VersionExt = "txt"
	fs, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

	prePath := filepath.Join(cfg.ChangesDir, "v0.1.2-a1.txt")
	then.CreateFile(t, afs, prePath)

	mockErr := errors.New("remove failed")
	fs.MockRemove = func(name string) error {
		if name == prePath {
			return mockErr
		}

		return fs.MemFS.Remove(name)
	}

	err := batchPipeline(standard, afs, "v0.2.0")
	then.Err(t, mockErr, err)
}

func TestBatchErrorBadLatestVersion(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ChangesDir = ".../.../.../"
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	// fail trying to read a bad directory
	err := batchPipeline(standard, afs, "v0.2.0")
	then.NotNil(t, err)
}

func TestBatchErrorBadCreateFile(t *testing.T) {
	cfg := batchTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "C"})

	mockErr := errors.New("bad create")
	fs.MockCreate = func(name string) (afero.File, error) {
		var f afero.File
		return f, mockErr
	}

	err := batchPipeline(standard, afs, "v0.1.1")
	then.Err(t, mockErr, err)
}

func TestBatchErrorBadConfig(t *testing.T) {
	_, afs := then.WithAferoFS()
	standard, _ := withPipelines(afs)

	configData := []byte("not a proper config")
	err := afs.WriteFile(core.ConfigPaths[0], configData, os.ModePerm)
	then.Nil(t, err)

	err = batchPipeline(standard, afs, "v0.1.0")
	then.NotNil(t, err)
}

func TestBatchErrorBadVersionFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.VersionFormat = "{{bad.format}"
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "removed", Body: "C"})

	err := batchPipeline(standard, afs, "v0.11.0")
	then.NotNil(t, err)
}

func TestBatchErrorBadVersionHeaderWrite(t *testing.T) {
	versionHeaderPathFlag = "h1.md"

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

	mockErr := errors.New("bad version write")
	mockPipeline.MockWriteFile = func(
		writer io.Writer,
		config core.Config,
		beforeNewlines int,
		afterNewlines int,
		relativePath string,
		templateData interface{},
	) error {
		if strings.HasSuffix(relativePath, versionHeaderPathFlag) {
			return mockErr
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
	then.Err(t, mockErr, err)
}

func TestBatchErrorBadVersionHeaderWriteWithConfig(t *testing.T) {
	cfg := batchTestConfig()
	cfg.VersionHeaderPath = "h2.md"
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

	mockErr := errors.New("bad version write")
	mockPipeline.MockWriteFile = func(
		writer io.Writer,
		config core.Config,
		beforeNewlines int,
		afterNewlines int,
		relativePath string,
		templateData interface{},
	) error {
		if strings.HasSuffix(relativePath, cfg.VersionHeaderPath) {
			return mockErr
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
	then.Err(t, mockErr, err)
}

func TestBatchErrorBadVersionFooterWrite(t *testing.T) {
	versionFooterPathFlag = "f1.md"

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

	mockErr := errors.New("bad version write")
	mockPipeline.MockWriteFile = func(
		writer io.Writer,
		config core.Config,
		beforeNewlines int,
		afterNewlines int,
		relativePath string,
		templateData interface{},
	) error {
		if strings.HasSuffix(relativePath, versionFooterPathFlag) {
			return mockErr
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
	then.Err(t, mockErr, err)
}

func TestBatchErrorBadVersionFooterWriteWithConfig(t *testing.T) {
	cfg := batchTestConfig()
	cfg.VersionFooterPath = "f2.md"
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

	mockErr := errors.New("bad version write")
	mockPipeline.MockWriteFile = func(
		writer io.Writer,
		config core.Config,
		beforeNewlines int,
		afterNewlines int,
		relativePath string,
		templateData interface{},
	) error {
		if strings.HasSuffix(relativePath, cfg.VersionFooterPath) {
			return mockErr
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
	then.Err(t, mockErr, err)
}

func TestBatchErrorBadHeaderFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.HeaderFormat = "bad format {{...{{"
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

	err := batchPipeline(mockPipeline, afs, "v0.1.1")
	then.NotNil(t, err)
}

func TestBatchErrorBadFooterFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.FooterFormat = "bad format {{...{{"
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "A"})

	err := batchPipeline(mockPipeline, afs, "v0.1.1")
	then.NotNil(t, err)
}

func TestBatchErrorBadDelete(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "C"})

	mockErr := errors.New("bad clear unreleased")
	mockPipeline.MockClearUnreleased = func(
		config core.Config,
		moveDir string,
		includeDirs []string,
		otherFiles ...string,
	) error {
		return mockErr
	}

	err := batchPipeline(mockPipeline, afs, "v0.2.3")
	then.Err(t, mockErr, err)
}

func TestBatchErrorBadWrite(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	_, mockPipeline := withPipelines(afs)

	writeChangeFile(t, afs, cfg, core.Change{Kind: "added", Body: "C"})

	mockErr := errors.New("bad clear unreleased")
	mockPipeline.MockWriteChanges = func(
		writer io.Writer,
		cfg core.Config,
		changes []core.Change,
	) error {
		return mockErr
	}

	err := batchPipeline(mockPipeline, afs, "v0.2.3")
	then.Err(t, mockErr, err)
}

func TestBatchSkipWritingUnreleasedIfPathIsEmpty(t *testing.T) {
	var builder strings.Builder

	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	text := []byte("some text")
	then.WriteFile(t, afs, text, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	err := standard.WriteFile(&builder, *cfg, 0, 0, "", nil)
	then.Equals(t, "", builder.String())
	then.Nil(t, err)
}

func TestBatchErrorOnWriteUnreleasedIfBadRead(t *testing.T) {
	var builder strings.Builder

	cfg := batchTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	mockErr := errors.New("bad mock open")
	fs.MockOpen = func(path string) (afero.File, error) {
		var f afero.File
		return f, mockErr
	}

	err := standard.WriteFile(&builder, *cfg, 0, 0, "a.md", nil)
	then.Equals(t, "", builder.String())
	then.Err(t, mockErr, err)
}

func TestBatchErrorWritingUnreleasedIfBadWrite(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	text := []byte("some text")
	then.WriteFile(t, afs, text, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	w := then.NewErrWriter()
	err := standard.WriteFile(w, *cfg, 0, 0, "a.yaml", nil)
	w.Raised(t, err)
}

func TestBatchSkipWriteIfFileDoesNotExist(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	var builder strings.Builder

	err := standard.WriteFile(&builder, *cfg, 0, 0, "a.md", nil)
	then.Equals(t, "", builder.String())
	then.Nil(t, err)
}

func TestBatchWriteChanges(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)
	changes := []core.Change{
		{Kind: "added", Body: "w"},
		{Kind: "added", Body: "x"},
		{Kind: "removed", Body: "y"},
		{Kind: "removed", Body: "z"},
	}

	var builder strings.Builder
	err := standard.WriteChanges(&builder, *cfg, changes)
	then.Nil(t, err)

	expected := `
### added
* w
* x
### removed
* y
* z`
	then.Equals(t, expected, builder.String())
}

func TestBatchCreateVersionsWithoutKindHeaders(t *testing.T) {
	cfg := batchTestConfig()
	cfg.KindFormat = ""
	cfg.ChangeFormat = "* {{.Body}} ({{.Kind}})"

	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)
	changes := []core.Change{
		{Body: "w", Kind: "added"},
		{Body: "x", Kind: "added"},
		{Body: "y", Kind: "removed"},
		{Body: "z", Kind: "removed"},
	}

	var builder strings.Builder
	err := standard.WriteChanges(&builder, *cfg, changes)
	then.Nil(t, err)

	expected := `
* w (added)
* x (added)
* y (removed)
* z (removed)`
	then.Equals(t, expected, builder.String())
}

func TestBatchVersionFileWithComponentHeaders(t *testing.T) {
	cfg := batchTestConfig()
	cfg.Components = []string{"linker", "compiler"}
	cfg.ComponentFormat = "## {{.Component}}"
	cfg.KindFormat = "### {{.Kind}}"
	cfg.ChangeFormat = "* {{.Body}}"

	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)
	changes := []core.Change{
		{Body: "w", Kind: "added", Component: "linker"},
		{Body: "x", Kind: "added", Component: "linker"},
		{Body: "y", Kind: "removed", Component: "linker"},
		{Body: "z", Kind: "removed", Component: "compiler"},
	}

	var builder strings.Builder
	err := standard.WriteChanges(&builder, *cfg, changes)
	then.Nil(t, err)

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
	then.Equals(t, expected, builder.String())
}

func TestBatchErrorBadKindTemplate(t *testing.T) {
	cfg := batchTestConfig()
	cfg.KindFormat = "{{randoooom../././}}"

	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)
	changes := []core.Change{
		{Body: "x", Kind: "added"},
	}

	var builder strings.Builder
	err := standard.WriteChanges(&builder, *cfg, changes)
	then.NotNil(t, err)
}

func TestBatchErrorBadComponentTemplate(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ComponentFormat = "{{deja vu}}"

	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)
	changes := []core.Change{
		{Component: "x", Kind: "added"},
	}

	var builder strings.Builder
	err := standard.WriteChanges(&builder, *cfg, changes)
	then.NotNil(t, err)
}

func TestBatchErrorBadChangeTemplate(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ChangeFormat = "{{not.valid.syntax....}}"

	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)
	changes := []core.Change{
		{Body: "x", Kind: "added"},
	}

	var builder strings.Builder
	err := standard.WriteChanges(&builder, *cfg, changes)
	then.NotNil(t, err)
}

func TestBatchClearUnreleasedREmovesUnreleasedFilesIncludingHeader(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	alphaPath := filepath.Join(cfg.ChangesDir, "alpha")
	betaPath := filepath.Join(cfg.ChangesDir, "beta")

	then.CreateFile(t, afs, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")
	then.CreateFile(t, afs, cfg.ChangesDir, cfg.UnreleasedDir, ".gitkeep")
	then.CreateFile(t, afs, cfg.ChangesDir, cfg.UnreleasedDir, "header.md")
	then.CreateFile(t, afs, alphaPath, "b.yaml")
	then.CreateFile(t, afs, alphaPath, ".gitkeep")
	then.CreateFile(t, afs, betaPath, "c.yaml")

	err := standard.ClearUnreleased(
		*cfg,
		"",
		[]string{"alpha", "beta"},
		"header.md",
		"",
		"does-not-exist.md",
	)
	then.Nil(t, err)

	infos, err := afs.ReadDir(filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir))
	then.Nil(t, err)
	then.Equals(t, 1, len(infos))

	infos, err = afs.ReadDir(alphaPath)
	then.Nil(t, err)
	then.Equals(t, 1, len(infos))

	_, err = afs.Stat(betaPath)
	then.Err(t, os.ErrNotExist, err)
}

func TestBatchClearUnreleasedMovesFilesIncludingHeaderIfSpecified(t *testing.T) {
	moveDirFlag = "beta"

	t.Cleanup(batchCleanArgs)

	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	for _, name := range []string{"a.yaml", "b.yaml", "c.yaml", ".gitkeep", "header.md"} {
		then.CreateFile(t, afs, cfg.ChangesDir, cfg.UnreleasedDir, name)
	}

	err := standard.ClearUnreleased(
		*cfg,
		moveDirFlag,
		nil,
		"header.md",
		"",
		"does-not-exist.md",
	)
	then.Nil(t, err)

	// should of moved the unreleased and header file to beta
	infos, err := afs.ReadDir(filepath.Join(cfg.ChangesDir, "beta"))
	then.Nil(t, err)
	then.Equals(t, 4, len(infos))
}

func TestBatchErrorClearUnreleasedIfMoveFails(t *testing.T) {
	cfg := batchTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	then.CreateFile(t, afs, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	mockErr := errors.New("bad mock remove")
	fs.MockRemove = func(name string) error {
		return mockErr
	}

	err := standard.ClearUnreleased(*cfg, "", nil)
	then.Err(t, mockErr, err)
}

func TestBatchErrorUnreleasedIfRemoveAllFails(t *testing.T) {
	cfg := batchTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	alphaPath := filepath.Join(cfg.ChangesDir, "alpha")

	then.CreateFile(t, afs, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")
	then.CreateFile(t, afs, alphaPath, "b.yaml")

	mockErr := errors.New("bad mock remove all")
	fs.MockRemoveAll = func(path string) error {
		return mockErr
	}

	err := standard.ClearUnreleased(
		*cfg,
		"",
		[]string{"alpha"},
		"header.md",
		"",
		"does-not-exist.md",
	)
	then.Err(t, mockErr, err)
}

func TestBatchUnreleasedFailsIfMoveFails(t *testing.T) {
	cfg := batchTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	then.CreateFile(t, afs, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	mockErr := errors.New("bad mock rename")
	fs.MockRename = func(before, after string) error {
		return mockErr
	}

	err := standard.ClearUnreleased(*cfg, "beta", nil)
	then.Err(t, mockErr, err)
}

func TestBatchUnreleasedFailsIfMakeDirFails(t *testing.T) {
	cfg := batchTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	then.CreateFile(t, afs, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	mockErr := errors.New("bad mock mkdir all")
	fs.MockMkdirAll = func(p string, mode os.FileMode) error {
		return mockErr
	}

	err := standard.ClearUnreleased(*cfg, "beta", nil)
	then.Err(t, mockErr, err)
}

func TestBatchUnreleasedFailsIfUnableToFindFiles(t *testing.T) {
	cfg := batchTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	standard, _ := withPipelines(afs)

	err := standard.ClearUnreleased(*cfg, "beta", []string{"../../bad"})
	then.NotNil(t, err)
}

type MockBatchPipeline struct {
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
