package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func TestMergeVersionsSuccessfully(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.HeaderPath = ""
	cfg.Replacements = nil
	then.WithTempDirConfig(t, cfg, cfg.ChangesDir, cfg.UnreleasedDir)

	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Open,
		os.Create,
		core.NewTemplateCache(),
	)
	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContents := `second version
first version
`
	then.FileContents(t, changeContents, "news.md")
}

func TestMergeVersionsWithHeaderAndReplacements(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg, cfg.ChangesDir)
	jsonContents := `{
  "key": "value",
  "version": "old-version",
}`

	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")
	then.WriteFile(t, []byte("ignored\n"), cfg.ChangesDir, "ignored.txt")
	then.WriteFile(t, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)
	then.WriteFile(t, []byte(jsonContents), "replace.json")

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Open,
		os.Create,
		core.NewTemplateCache(),
	)
	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContents := `a simple header
second version
first version
`
	then.FileContents(t, changeContents, cfg.ChangelogPath)

	newContents := `{
  "key": "value",
  "version": "0.2.0",
}`
	then.FileContents(t, newContents, "replace.json")
}

func TestMergeDryRun(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg, cfg.ChangesDir)

	changeContents := `a simple header
second version
first version
`
	writer := strings.Builder{}

	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")
	then.WriteFile(t, []byte("ignored\n"), cfg.ChangesDir, "ignored.txt")
	then.WriteFile(t, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Open,
		os.Create,
		core.NewTemplateCache(),
	)
	cmd.DryRun = true
	cmd.Command.SetOut(&writer)
	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, changeContents, writer.String())
}

func TestMergeSkipsVersionsIfNoneFound(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg, cfg.ChangesDir)
	changeContents := "a simple header\n"
	writer := strings.Builder{}

	then.WriteFile(t, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Open,
		os.Create,
		core.NewTemplateCache(),
	)
	cmd.DryRun = true
	cmd.Command.SetOut(&writer)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, changeContents, writer.String())
}

func TestErrorMergeBadChangelogPath(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg)

	badError := errors.New("bad create")
	mockCreate := func(filename string) (*os.File, error) {
		return nil, badError
	}

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Open,
		mockCreate,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.Err(t, badError, err)
}

func TestErrorMergeBadConfig(t *testing.T) {
	then.WithTempDir(t)

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Open,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestErrorMergeUnableToReadChanges(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg, cfg.ChangesDir)

	mockErr := errors.New("bad read dir")
	mockReadDir := func(name string) ([]os.DirEntry, error) {
		return nil, mockErr
	}

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		mockReadDir,
		os.Open,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.Err(t, mockErr, err)
}

func TestErrorMergeBadHeaderFile(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg, cfg.ChangesDir)
	mockError := errors.New("bad open")

	mockOpen := func(filename string) (*os.File, error) {
		if filename == filepath.Join(cfg.ChangesDir, cfg.HeaderPath) {
			return nil, mockError
		}

		return os.Open(filename)
	}

	// need at least one change to exist
	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		mockOpen,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.Err(t, mockError, err)
}

func TestErrorMergeBadAppend(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg, cfg.ChangesDir)
	mockError := errors.New("bad write string")

	// need at least one change to exist
	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)

	// create a version file, then fail to open it the second time
	mockOpen := func(filename string) (*os.File, error) {
		if filename == filepath.Join(cfg.ChangesDir, "v0.1.0.md") {
			return nil, mockError
		}

		return os.Open(filename)
	}

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		mockOpen,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.Err(t, mockError, err)
}

func TestErrorMergeBadReplacement(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.Replacements[0].Replace = "{{bad....}}"
	then.WithTempDirConfig(t, cfg, cfg.ChangesDir)

	then.WriteFile(t, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)
	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Open,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}
