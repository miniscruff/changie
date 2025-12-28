package cmd

import (
	"fmt"
	"io/fs"
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
		// RootDir:         "news",
		// Fragment.Dir:      "future",
		// HeaderPath:         "",
		// ChangelogPath:      "news.md",
		// VersionFileFormat:  "{{.Version}}.md",
		// VersionExt:         "md",
		// VersionFormat:      "## {{.Version}}",
		// KindFormat:         "### {{.Kind}}",
		// ChangeFormat:       "* {{.Body}}",
		// FragmentFileFormat: "",
		// Kinds: []core.KindConfig{
		// {Label: "added"},
		// {Label: "removed"},
		// {Label: "other"},
		// },
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
		change.Filename = filepath.Join(cfg.RootDir, cfg.Fragment.Dir, name)
	}

	then.WriteFile(t, bs, change.Filename)
}

func TestBatchCanBatch(t *testing.T) {
	cfg := batchTestConfig()
	// declared path but missing is accepted
	cfg.ReleaseNotes.Header.FilePath = "header.md"

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

	then.FileContents(t, verContents, cfg.RootDir, "v0.2.0.md")
	then.DirectoryFileCount(t, 0, cfg.RootDir, cfg.Fragment.Dir)
}

func TestBatchCanBatchWithProject(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ReleaseNotes.Version.Format = "{{.Version}}-custom-format.md"
	cfg.ReleaseNotes.Header.FilePath = "header.md"
	cfg.Project = core.ProjectConfig{
		Options: []core.ProjectOptions{
			{
				Label: "A",
				Key:   "a",
			},
		},
	}
	// declared path but missing is accepted

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

	then.FileContents(t, verContents, cfg.RootDir, "a", "v0.2.0-custom-format.md")
	then.DirectoryFileCount(t, 1, cfg.RootDir, cfg.Fragment.Dir)
}

func TestBatchProjectFailsIfUnableToFindProject(t *testing.T) {
	cfg := batchTestConfig()
	cfg.Project = core.ProjectConfig{
		Options: []core.ProjectOptions{
			{
				Label: "A",
				Key:   "a",
			},
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
	cfg.ReleaseNotes.Header.FilePath = "header.md"
	cfg.Kind.Newlines.Before = 2
	cfg.Kind.Newlines.After = 2

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

	then.FileContents(t, verContents, cfg.RootDir, "v0.2.0.md")
	then.DirectoryFileCount(t, 0, cfg.RootDir, cfg.Fragment.Dir)
}

func TestBatchErrorStandardBadWriter(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	w := then.NewErrWriter()

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.writer = w
	err := batch.WriteTemplate(
		cfg.ReleaseNotes.Version.Format,
		2, // tries to write some before lines and fails
		cfg.ReleaseNotes.Version.Newlines.After,
		"v0.2.0",
	)
	w.Raised(t, err)
}

func TestBatchComplexVersion(t *testing.T) {
	cfg := batchTestConfig()
	cfg.EnvPrefix = "TEST_CHANGIE_"
	cfg.ReleaseNotes.Footer.Format = `### contributors
{{- range (customs .Changes "Author" | uniq) }}
* [{{.}}](https://github.com/{{.}})
{{- end}}
env: {{.Env.ENV_KEY}}`
	cfg.ReleaseNotes.Version.Format = "## {{.VersionNoPrefix}}"
	cfg.ReleaseNotes.Header.Format = "{{ bodies .Changes | len }} changes this release"
	cfg.Change.Prompt.Format = "* {{.Body}} by {{.Custom.Author}}"

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

	then.WriteFileTo(t, betaChange, cfg.RootDir, "beta", "change.yaml")

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
	then.FileContents(t, verContents, cfg.RootDir, "v0.2.0-b1+hash.md")
	then.DirectoryFileCount(t, 0, cfg.RootDir, cfg.Fragment.Dir)
	then.FileNotExists(t, cfg.RootDir, "beta", "change.yaml")
}

func TestBatchDryRun(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.DryRun = true

	var builder strings.Builder

	batch.SetOut(&builder)
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
	then.DirectoryFileCount(t, 3, cfg.RootDir, cfg.Fragment.Dir)
}

func TestBatchDryRunWithKeys(t *testing.T) {
	cfg := batchTestConfig()
	cfg.Kind.Options[0] = core.KindOptions{
		Prompt: core.PromptFormat{
			Label: ":fire: Added",
			Key:   "added",
		},
	}
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.DryRun = true

	var builder strings.Builder

	batch.SetOut(&builder)
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "D"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "E"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "F"})

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	verContents := `## v0.2.0
### :fire: Added
* D
* E
### removed
* F`
	then.Equals(t, verContents, builder.String())
	then.DirectoryFileCount(t, 3, cfg.RootDir, cfg.Fragment.Dir)
}

func TestBatchRemovePrereleases(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.RemovePrereleases = true

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	then.CreateFile(t, cfg.RootDir, "v0.1.2-a1.md")
	then.CreateFile(t, cfg.RootDir, "v0.1.2-a2.md")

	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	infos, err := os.ReadDir(cfg.RootDir)
	then.Nil(t, err)
	then.Equals(t, cfg.Fragment.Dir, infos[0].Name())
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
	then.DirectoryFileCount(t, 3, cfg.RootDir, cfg.Fragment.Dir)
}

func TestBatchWithHeadersAndFooters(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ReleaseNotes.Header.FilePath = "h2.md"
	cfg.ReleaseNotes.Footer.FilePath = "f2.md"
	cfg.ReleaseNotes.Header.Format = "header format"
	cfg.ReleaseNotes.Footer.Format = "footer format"
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.VersionHeaderPath = "h1.md"
	batch.VersionFooterPath = "f1.md"

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "B"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	then.WriteFile(t, []byte("first header\n"), cfg.RootDir, cfg.Fragment.Dir, "h1.md")
	then.WriteFile(t, []byte("second header\n"), cfg.RootDir, cfg.Fragment.Dir, "h2.md")
	then.WriteFile(t, []byte("first footer\n"), cfg.RootDir, cfg.Fragment.Dir, "f1.md")
	then.WriteFile(t, []byte("second footer\n"), cfg.RootDir, cfg.Fragment.Dir, "f2.md")

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
	then.FileContents(t, verContents, cfg.RootDir, "v0.2.0.md")
}

func TestBatchErrorBadChanges(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	aVer := []byte("not a valid change")
	then.WriteFile(t, aVer, cfg.RootDir, cfg.Fragment.Dir, "a.yaml")

	err := batch.Run(batch.Command, []string{"v0.1.1"})
	then.NotNil(t, err)
}

func TestBatchErrorNoChangesAndNotAllowed(t *testing.T) {
	cfg := batchTestConfig()
	then.WithTempDirConfig(t, cfg)

	err := os.MkdirAll(filepath.Join("news", "future"), 0777)
	then.Nil(t, err)

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.AllowNoChanges = false

	err = batch.Run(batch.Command, []string{"v0.1.1"})
	then.Err(t, errNoChangesNotAllowed, err)
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
	then.FileContents(t, verContents, cfg.RootDir, "v0.2.0.md")
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
	cfg.RootDir = "../../opt/apt/not/real"
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
	cfg.ReleaseNotes.Version.Format = "{{bad.format}"
	then.WithTempDirConfig(t, cfg)

	batch := NewBatch(time.Now, core.NewTemplateCache())

	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "C"})

	err := batch.Run(batch.Command, []string{"v0.11.0"})
	then.NotNil(t, err)
}

func TestBatchErrorBadKindFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.Kind.Prompt.Format = "bad {{...buh}}"
	batch := NewBatch(time.Now, core.NewTemplateCache())

	then.WithTempDirConfig(t, cfg)
	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "C"})

	err := batch.Run(batch.Command, []string{"v0.2.3"})
	then.NotNil(t, err)

	_, err = os.Stat(filepath.Join(cfg.RootDir, "v0.2.3.md"))
	then.Err(t, fs.ErrNotExist, err)
}

func TestBatchErrorBadComponentFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.Component.Prompt.Format = "bad {{...buh}}"
	batch := NewBatch(time.Now, core.NewTemplateCache())

	then.WithTempDirConfig(t, cfg)
	writeChangeFile(t, cfg, &core.Change{Component: "ui", Body: "C"})

	err := batch.Run(batch.Command, []string{"v0.2.3"})
	then.NotNil(t, err)
}

func TestBatchErrorBadChangeFormat(t *testing.T) {
	cfg := batchTestConfig()
	cfg.Change.Prompt.Format = "bad {{...buh}}"
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
	batch.cfg = cfg
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
	cfg.Kind.Prompt.Format = ""
	cfg.Change.Prompt.Format = "* {{.Body}} ({{.Kind}})"

	var builder strings.Builder

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.cfg = cfg
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
	cfg.Component = core.ComponentConfig{
		Prompt: core.PromptFormat{
			Format: "## {{.Component}}",
		},
		Options: []core.ComponentOptions{
			{
				Prompt: core.PromptFormat{Key: "linker"},
			},
			{
				Prompt: core.PromptFormat{Key: "compiler"},
			},
		},
	}
	cfg.Kind.Prompt.Format = "### {{.Kind}}"
	cfg.Change.Prompt.Format = "* {{.Body}}"

	var builder strings.Builder

	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.cfg = cfg
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
	batch.cfg = cfg

	alphaPath := filepath.Join(cfg.RootDir, "alpha")
	betaPath := filepath.Join(cfg.RootDir, "beta")
	batch.IncludeDirs = []string{"alpha", "beta"}

	then.CreateFile(t, cfg.RootDir, cfg.Fragment.Dir, ".gitkeep")
	then.CreateFile(t, cfg.RootDir, cfg.Fragment.Dir, "header.md")
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

	then.DirectoryFileCount(t, 1, cfg.RootDir, cfg.Fragment.Dir)
	then.DirectoryFileCount(t, 1, alphaPath)

	then.FileNotExists(t, betaPath)
}

func TestBatchClearUnreleasedMovesFilesIncludingHeaderIfSpecified(t *testing.T) {
	then.WithTempDir(t)

	cfg := batchTestConfig()
	batch := NewBatch(time.Now, core.NewTemplateCache())
	batch.cfg = cfg
	batch.MoveDir = "beta"

	for _, name := range []string{".gitkeep", "header.md"} {
		then.CreateFile(t, cfg.RootDir, cfg.Fragment.Dir, name)
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
	then.DirectoryFileCount(t, 4, cfg.RootDir, "beta")
	// .gitkeep should remain
	then.DirectoryFileCount(t, 1, cfg.RootDir, cfg.Fragment.Dir)
}

func TestBatchCanAddNewLinesAfterReleaseNotes(t *testing.T) {
	cfg := batchTestConfig()
	cfg.ReleaseNotes.Newlines.After = 2

	then.WithTempDirConfig(t, cfg)

	writeChangeFile(t, cfg, &core.Change{Kind: "added", Body: "A"})
	writeChangeFile(t, cfg, &core.Change{Kind: "removed", Body: "B"})

	batch := NewBatch(time.Now, core.NewTemplateCache())
	err := batch.Run(batch.Command, []string{"v0.2.0"})
	then.Nil(t, err)

	verContents := `## v0.2.0
### added
* A
### removed
* B

`
	then.FileContents(t, verContents, cfg.RootDir, "v0.2.0.md")
	then.DirectoryFileCount(t, 0, cfg.RootDir, cfg.Fragment.Dir)
}
