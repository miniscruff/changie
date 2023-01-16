package cmd

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func mergeTestConfig() *core.Config {
	return &core.Config{
		ChangesDir:    "news",
		UnreleasedDir: "future",
		HeaderPath:    "header.rst",
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
		Replacements: []core.Replacement{
			{
				Path:    "replace.json",
				Find:    `  "version": ".*",`,
				Replace: `  "version": "{{.VersionNoPrefix}}",`,
			},
		},
	}
}

func mergeResetVars() {
	_mergeDryRun = false
}

func TestMergeVersionsSuccessfully(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.HeaderPath = ""
	cfg.Replacements = nil
	_, afs := then.WithAferoFSConfig(t, cfg)

	then.WriteFile(t, afs, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, afs, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")

	err := mergePipeline(afs)
	then.Nil(t, err)

	changeContents := `second version
first version
`
	then.FileContents(t, afs, changeContents, "news.md")
}

func TestMergeVersionsWithHeaderAndReplacements(t *testing.T) {
	cfg := mergeTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	jsonContents := `{
  "key": "value",
  "version": "old-version",
}`

	then.WriteFile(t, afs, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, afs, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")
	then.WriteFile(t, afs, []byte("ignored\n"), cfg.ChangesDir, "ignored.txt")
	then.WriteFile(t, afs, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)
	then.WriteFile(t, afs, []byte(jsonContents), "replace.json")

	err := mergePipeline(afs)
	then.Nil(t, err)

	changeContents := `a simple header
second version
first version
`
	then.FileContents(t, afs, changeContents, cfg.ChangelogPath)

	newContents := `{
  "key": "value",
  "version": "0.2.0",
}`
	then.FileContents(t, afs, newContents, "replace.json")
}

func TestMergeDryRun(t *testing.T) {
	cfg := mergeTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	changeContents := `a simple header
second version
first version
`
	writer := then.WithStdout(t)
	mergeDryRunOut = writer
	_mergeDryRun = true

	t.Cleanup(mergeResetVars)
	then.WriteFile(t, afs, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, afs, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")
	then.WriteFile(t, afs, []byte("ignored\n"), cfg.ChangesDir, "ignored.txt")
	then.WriteFile(t, afs, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)

	err := mergePipeline(afs)
	then.Nil(t, err)
	then.Equals(t, changeContents, writer.String())
}

func TestMergeSkipsVersionsIfNoneFound(t *testing.T) {
	cfg := mergeTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	changeContents := "a simple header\n"

	writer := then.WithStdout(t)
	mergeDryRunOut = writer
	_mergeDryRun = true

	then.WriteFile(t, afs, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)
	t.Cleanup(mergeResetVars)

	err := mergePipeline(afs)
	then.Nil(t, err)
	then.Equals(t, changeContents, writer.String())
}

func TestErrorMergeBadChangelogPath(t *testing.T) {
	cfg := mergeTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)

	badError := errors.New("bad create")
	fs.MockCreate = func(filename string) (afero.File, error) {
		var f afero.File
		return f, badError
	}

	err := mergePipeline(afs)
	then.Err(t, badError, err)
}

func TestErrorMergeUnableToReadChanges(t *testing.T) {
	cfg := mergeTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)

	// no files, means bad read
	err := mergePipeline(afs)
	then.NotNil(t, err)
}

func TestErrorMergeBadHeaderFile(t *testing.T) {
	cfg := mergeTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)
	mockError := errors.New("bad open")

	fs.MockOpen = func(filename string) (afero.File, error) {
		if filename == filepath.Join(cfg.ChangesDir, cfg.HeaderPath) {
			return nil, mockError
		}

		return fs.MemFS.Open(filename)
	}

	// need at least one change to exist
	then.WriteFile(t, afs, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")

	err := mergePipeline(afs)
	then.Err(t, mockError, err)
}

func TestErrorMergeBadAppend(t *testing.T) {
	cfg := mergeTestConfig()
	fs, afs := then.WithAferoFSConfig(t, cfg)
	mockError := errors.New("bad write string")

	// need at least one change to exist
	then.WriteFile(t, afs, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, afs, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)

	// create a version file, then fail to open it the second time
	fs.MockOpen = func(filename string) (afero.File, error) {
		if filename == filepath.Join(cfg.ChangesDir, "v0.1.0.md") {
			return nil, mockError
		}

		return fs.MemFS.Open(filename)
	}

	err := mergePipeline(afs)
	then.Err(t, mockError, err)
}

func TestErrorMergeBadReplacement(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.Replacements[0].Replace = "{{bad....}}"
	_, afs := then.WithAferoFSConfig(t, cfg)

	then.WriteFile(t, afs, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)
	then.WriteFile(t, afs, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")

	err := mergePipeline(afs)
	then.NotNil(t, err)
}
