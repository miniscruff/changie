package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

var changeIncrementer = 0

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

// writeChangeFile will write a change file with an auto-incrementing index to prevent
// same second clobbering
func writeChangeFile(t *testing.T, cfg *core.Config, change *core.Change) {
	// set our time as an arbitrary amount from jan 1 2000 so
	// each change is 1 hour later then the last
	if change.Time.Year() == 0 {
		diff := time.Duration(changeIncrementer) * time.Hour
		change.Time = time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC).Add(diff)
	}

	bs, _ := yaml.Marshal(&change)
	name := fmt.Sprintf("change-%d.yaml", changeIncrementer)
	changeIncrementer++

	if change.Filename == "" {
		change.Filename = filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir, name)
	}

	then.WriteFile(t, bs, change.Filename)
}

func TestBatchCanBatch(t *testing.T) {
	cfg := batchTestConfig()
	// declared path but missing is accepted
	cfg.VersionHeaderPath = "header.md"

	then.WithTempDirConfig(t, cfg)

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	batch := NewBatch(time.Now, core.NewTemplateCache())
	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	verContents := `## v0.2.0
### added
* A
* B
### removed
* C`

	then.FileContents(t, verContents, cfg.ChangesDir, "v0.2.0.md")
	then.DirectoryFileCount(t, 0, cfg.ChangesDir, cfg.UnreleasedDir)
}

func TestBatchCanBatchWithProject(t *testing.T) {
	cfg := batchTestConfig()
	cfg.Projects = []core.ProjectConfig{
		{
			Label: "A",
			Key:   "a",
		},
	}
	// declared path but missing is accepted
	cfg.VersionHeaderPath = "header.md"

	then.WithTempDirConfig(t, cfg)

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A", Project: "a"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B", Project: "b"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C", Project: "a"})

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.Project = "a"
	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	verContents := `## v0.2.0
### added
* A
### removed
* C`

	then.FileContents(t, verContents, cfg.ChangesDir, "a", "v0.2.0.md")
	then.DirectoryFileCount(t, 1, cfg.ChangesDir, cfg.UnreleasedDir)
}

func TestBatchProjectFailsIfUnableToFindProject(t *testing.T) {
	cfg := batchTestConfig()
	cfg.Projects = []core.ProjectConfig{
		{
			Label: "A",
			Key:   "a",
		},
	}

	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.Project = "not found"

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.NotNil(t, err)
}

func TestBatchCanAddNewLinesBeforeAndAfterKindHeader(t *testing.T) {
	cfg := batchTestConfig()
	// declared path but missing is accepted
	cfg.VersionHeaderPath = "header.md"
	cfg.Newlines.BeforeKind = 2
	cfg.Newlines.AfterKind = 2

	then.WithTempDirConfig(t, cfg)

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	batch := NewBatch(time.Now, core.NewTemplateCache())
	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	verContents := `## v0.2.0


### added


* A
* B


### removed


* C`

	then.FileContents(t, verContents, cfg.ChangesDir, "v0.2.0.md")
	then.DirectoryFileCount(t, 0, cfg.ChangesDir, cfg.UnreleasedDir)
}

func TestBatchErrorStandardBadWriter(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)
	w := then.NewErrWriter()

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.writer = w
	err := batch.WriteTemplate(
		cfg.VersionFormat,
		2, // tries to write some before lines and fails
		cfg.Newlines.AfterVersion,
		"v0.2.0",
	)
	w.Raised(t, err)
}

func TestBatchComplexVersion(t *testing.T) {
	cfg := batchTestConfig()
	cfg.EnvPrefix = "TEST_CHANGIE_"
	cfg.HeaderFormat = "{{ bodies .Changes | len }} changes this release"
	cfg.ChangeFormat = "* {{.Body}} by {{.Custom.Author}}"
	cfg.FooterFormat = `### contributors
{{- range (customs .Changes "Author" | uniq) }}
* [{{.}}](https://github.com/{{.}})
{{- end}}
env: {{.Env.ENV_KEY}}`
	cfg.VersionFormat = "## {{.VersionNoPrefix}}"

	then.WithTempDirConfig(t, cfg)
	t.Setenv("TEST_CHANGIE_ENV_KEY", "env value")

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.IncludeDirs = []string{"beta"}
	batch.Prerelease = []string{"b1"}
	batch.Meta = []string{"hash"}

	writeChangeFile(t, cfg,
		&core.Change{
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

	then.WriteFileTo(t, betaChange, cfg.ChangesDir, "beta", "change.yaml")

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	verContents := `## 0.2.0-b1+hash
2 changes this release
### added
* D by miniscruff
* E by otherAuthor
### contributors
* [miniscruff](https://github.com/miniscruff)
* [otherAuthor](https://github.com/otherAuthor)
env: env value`
	then.FileContents(t, verContents, cfg.ChangesDir, "v0.2.0-b1+hash.md")
	then.DirectoryFileCount(t, 0, cfg.ChangesDir, cfg.UnreleasedDir)
	then.FileNotExists(t, cfg.ChangesDir, "beta", "change.yaml")
}

func TestBatchDryRun(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.DryRun = true

	var builder strings.Builder

	batch.Command.SetOut(&builder)
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "D"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "E"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "F"})

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	verContents := `## v0.2.0
### added
* D
* E
### removed
* F`
	then.Equals(t, verContents, builder.String())
	then.DirectoryFileCount(t, 3, cfg.ChangesDir, cfg.UnreleasedDir)
}

func TestBatchRemovePrereleases(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.RemovePrereleases = true

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	then.CreateFile(t, cfg.ChangesDir, "v0.1.2-a1.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.1.2-a2.md")

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	infos, err := os.ReadDir(cfg.ChangesDir)
	then.Nil(t, err)
	then.Equals(t, cfg.UnreleasedDir, infos[0].Name())
	then.Equals(t, "v0.2.0.md", infos[1].Name())
	then.Equals(t, 2, len(infos))
}

func TestBatchKeepChangeFiles(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.KeepFragments = true

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)
	then.DirectoryFileCount(t, 3, cfg.ChangesDir, cfg.UnreleasedDir)
}

func TestBatchWithHeadersAndFooters(t *testing.T) {
	cfg := batchTestConfig()
	cfg.VersionHeaderPath = "h2.md"
	cfg.VersionFooterPath = "f2.md"
	cfg.HeaderFormat = "header format"
	cfg.FooterFormat = "footer format"
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.VersionHeaderPath = "h1.md"
	batch.VersionFooterPath = "f1.md"

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	then.WriteFile(t, []byte("first header\n"), cfg.ChangesDir, cfg.UnreleasedDir, "h1.md")
	then.WriteFile(t, []byte("second header\n"), cfg.ChangesDir, cfg.UnreleasedDir, "h2.md")
	then.WriteFile(t, []byte("first footer\n"), cfg.ChangesDir, cfg.UnreleasedDir, "f1.md")
	then.WriteFile(t, []byte("second footer\n"), cfg.ChangesDir, cfg.UnreleasedDir, "f2.md")

	err := batch.Run(batch.Command, []string{"v0.2.0"})
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
	then.FileContents(t, verContents, cfg.ChangesDir, "v0.2.0.md")
}

func TestBatchErrorBadChanges(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	aVer := []byte("not a valid change")
	then.WriteFile(t, aVer, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	err := batch.Run(batch.Command, []string{"v0.1.1"})
	then.NotNil(t, err)
}

func TestBatchOverrideIfForced(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.Force = true

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})

	err = batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	verContents := `## v0.2.0
### added
* B`
	then.FileContents(t, verContents, cfg.ChangesDir, "v0.2.0.md")
}

func TestBatchErrorIfBatchExists(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})

	err = batch.Run(batch.Command, []string{"v0.2.0"})
	then.Err(t, errVersionExists, err)
}

func TestBatchErrorBadVersion(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})

	err := batch.Run(batch.Command, []string{"---asdfasdf---"})
	then.Err(t, core.ErrBadVersionOrPart, err)
}

func TestBatchErrorBadLatestVersion(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ChangesDir = "../../opt/apt/not/real"
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())

	// fail trying to read a bad directory
	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.NotNil(t, err)
}

func TestBatchErrorBadConfig(t *testing.T) {
	then.WithTempDir(t)
	then.WriteFile(t, []byte("not a proper config"), core.ConfigPaths[0])

	batch := NewBatch(time.Now, core.NewTemplateCache())
	err := batch.Run(batch.Command, []string{"v0.1.0"})
	then.NotNil(t, err)
}

func TestBatchErrorBadVersionFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.VersionFormat = "{{bad.format}"
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())

	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	err := batch.Run(batch.Command, []string{"v0.11.0"})
	then.NotNil(t, err)
}

func TestBatchErrorBadKindFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.KindFormat = "bad {{...buh}}"
	batch := NewBatch(time.Now, core.NewTemplateCache())

	then.WithTempDirConfig(t, cfg)
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "C"})

	err := batch.Run(batch.Command, []string{"v0.2.3"})
	then.NotNil(t, err)
}

func TestBatchErrorBadComponentFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ComponentFormat = "bad {{...buh}}"
	batch := NewBatch(time.Now, core.NewTemplateCache())

	then.WithTempDirConfig(t, cfg)
	writeChangeFile(t, cfg, &core.Change{Component: "ui", Body: "C"})

	err := batch.Run(batch.Command, []string{"v0.2.3"})
	then.NotNil(t, err)
}

func TestBatchErrorBadChangeFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ChangeFormat = "bad {{...buh}}"
	batch := NewBatch(time.Now, core.NewTemplateCache())

	then.WithTempDirConfig(t, cfg)
	writeChangeFile(t, cfg, &core.Change{Component: "ui", Body: "C"})

	err := batch.Run(batch.Command, []string{"v0.2.3"})
	then.NotNil(t, err)
}

func TestBatchWriteChanges(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	var builder strings.Builder

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.config = cfg
	batch.writer = &builder

	changes := []core.Change{
		{Kind: "added", Body: "w"},
		{Kind: "added", Body: "x"},
		{Kind: "removed", Body: "y"},
		{Kind: "removed", Body: "z"},
	}

	err := batch.WriteChanges(changes)
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

	var builder strings.Builder

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.config = cfg
	batch.writer = &builder

	changes := []core.Change{
		{Body: "w", Kind: "added"},
		{Body: "x", Kind: "added"},
		{Body: "y", Kind: "removed"},
		{Body: "z", Kind: "removed"},
	}

	err := batch.WriteChanges(changes)
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

	var builder strings.Builder

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.config = cfg
	batch.writer = &builder

	changes := []core.Change{
		{Body: "w", Kind: "added", Component: "linker"},
		{Body: "x", Kind: "added", Component: "linker"},
		{Body: "y", Kind: "removed", Component: "linker"},
		{Body: "z", Kind: "removed", Component: "compiler"},
	}

	err := batch.WriteChanges(changes)
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

func TestBatchClearUnreleasedRemovesUnreleasedFilesIncludingHeader(t *testing.T) {
	then.WithTempDir(t)

	cfg := batchTestConfig()
	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.config = cfg

	alphaPath := filepath.Join(cfg.ChangesDir, "alpha")
	betaPath := filepath.Join(cfg.ChangesDir, "beta")
	batch.IncludeDirs = []string{"alpha", "beta"}

	then.CreateFile(t, cfg.ChangesDir, cfg.UnreleasedDir, ".gitkeep")
	then.CreateFile(t, cfg.ChangesDir, cfg.UnreleasedDir, "header.md")
	then.CreateFile(t, alphaPath, ".gitkeep")

	changes := []core.Change{
		{Kind: "added", Body: "A"},
		{Kind: "added", Body: "B"},
		{Kind: "added", Body: "alpha b", Filename: filepath.Join(alphaPath, "b.yaml")},
		{Kind: "added", Body: "beta c", Filename: filepath.Join(betaPath, "c.yaml")},
	}
	for i := range changes {
		writeChangeFile(t, cfg, &changes[i])
	}

	err := batch.ClearUnreleased(
		changes,
		"",
		"header.md",
		"does-not-exist.md",
	)
	then.Nil(t, err)

	then.DirectoryFileCount(t, 1, cfg.ChangesDir, cfg.UnreleasedDir)
	then.DirectoryFileCount(t, 1, alphaPath)

	then.FileNotExists(t, betaPath)
}

func TestBatchClearUnreleasedMovesFilesIncludingHeaderIfSpecified(t *testing.T) {
	then.WithTempDir(t)

	cfg := batchTestConfig()
	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.config = cfg
	batch.MoveDir = "beta"

	for _, name := range []string{".gitkeep", "header.md"} {
		then.CreateFile(t, cfg.ChangesDir, cfg.UnreleasedDir, name)
	}

	changes := []core.Change{
		{Kind: "added", Body: "A"},
		{Kind: "added", Body: "B"},
		{Kind: "added", Body: "C"},
	}
	for i := range changes {
		writeChangeFile(t, cfg, &changes[i])
	}

	err := batch.ClearUnreleased(
		changes,
		"header.md",
		"",
		"does-not-exist.md",
	)
	then.Nil(t, err)

	// should of moved the unreleased and header file to beta
	then.DirectoryFileCount(t, 4, cfg.ChangesDir, "beta")
	// .gitkeep should remain
	then.DirectoryFileCount(t, 1, cfg.ChangesDir, cfg.UnreleasedDir)
}
