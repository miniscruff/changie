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
		Newlines: core.NewlinesConfig{
			// BeforeVersion: 1,
			// AfterVersion: 1,
			AfterChange:            1,
			BeforeChangelogVersion: 0,
			// AfterChangelogVersion: 1,
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
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
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

func TestMergeVersionsSuccessfullyWithProject(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.HeaderPath = ""
	cfg.Replacements = nil
	cfg.Projects = []core.ProjectConfig{
		{
			Label:         "A thing",
			Key:           "a",
			ChangelogPath: "a/thing/CHANGELOG.md",
			Replacements: []core.Replacement{
				{
					Path:    "a/VERSION",
					Find:    "version",
					Replace: "{{.Version}}",
				},
			},
		},
	}
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "a", "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.ChangesDir, "a", "v0.2.0.md")
	then.WriteFile(t, []byte("version\n"), "a", "VERSION")
	then.Nil(t, os.MkdirAll(filepath.Join("a", "thing"), core.CreateDirMode))

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContents := `second version
first version
`
	then.FileContents(t, changeContents, "a", "thing", "CHANGELOG.md")
	then.FileContents(t, "v0.2.0\n", "a", "VERSION")
}

func TestMergeVersionsErrorMissingProjectDir(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.HeaderPath = ""
	cfg.Replacements = nil
	cfg.Projects = []core.ProjectConfig{
		{
			Label:         "A thing",
			Key:           "a",
			ChangelogPath: "a/thing/CHANGELOG.md",
		},
	}
	then.WithTempDirConfig(t, cfg)

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestMergeVersionsWithUnreleasedChanges(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.HeaderPath = ""
	cfg.Replacements = nil
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")

	unrel := core.Change{
		Kind: "Added",
		Body: "new feature coming soon",
	}
	writeChangeFile(t, cfg, &unrel)

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Create,
		core.NewTemplateCache(),
	)
	cmd.UnreleasedHeader = "## Coming Soon"
	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContents := `## Coming Soon
### Added
* new feature coming soon
second version
first version
`
	then.FileContents(t, changeContents, "news.md")
}

func TestMergeVersionsWithUnreleasedChangesErrorsOnBadChanges(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.HeaderPath = ""
	cfg.Replacements = nil
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")

	aVer := []byte("not a valid change")
	then.WriteFile(t, aVer, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Create,
		core.NewTemplateCache(),
	)
	cmd.UnreleasedHeader = "## Coming Soon"
	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestMergeVersionsWithUnreleasedChangesErrorsOnBadChangeFormat(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.HeaderPath = ""
	cfg.Replacements = nil
	cfg.ChangeFormat = "{{...invalid format{{{"
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.ChangesDir, "v0.2.0.md")

	unrel := core.Change{
		Kind: "Added",
		Body: "new feature coming soon",
	}
	writeChangeFile(t, cfg, &unrel)

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Create,
		core.NewTemplateCache(),
	)
	cmd.UnreleasedHeader = "## Coming Soon"
	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestMergeVersionsWithUnreleasedChangesInOneProject(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.HeaderPath = ""
	cfg.Replacements = nil
	cfg.Projects = []core.ProjectConfig{
		{
			Label:         "A thing",
			Key:           "a",
			ChangelogPath: "a/thing/CHANGELOG.md",
		},
		{
			Label:         "B thing",
			Key:           "b",
			ChangelogPath: "b/thing/CHANGELOG.md",
		},
	}
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first A version\n"), cfg.ChangesDir, "a", "v0.1.0.md")
	then.WriteFile(t, []byte("second A version\n"), cfg.ChangesDir, "a", "v0.2.0.md")
	then.Nil(t, os.MkdirAll(filepath.Join("a", "thing"), core.CreateDirMode))
	then.WriteFile(t, []byte("first B version\n"), cfg.ChangesDir, "b", "v0.1.0.md")
	then.WriteFile(t, []byte("second B version\n"), cfg.ChangesDir, "b", "v0.2.0.md")
	then.Nil(t, os.MkdirAll(filepath.Join("b", "thing"), core.CreateDirMode))

	unrel := core.Change{
		Kind:    "Added",
		Body:    "new feature coming soon",
		Project: "a",
	}
	writeChangeFile(t, cfg, &unrel)

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Create,
		core.NewTemplateCache(),
	)
	cmd.UnreleasedHeader = "## Coming Soon"
	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContentsA := `## Coming Soon
### Added
* new feature coming soon
second A version
first A version
`
	then.FileContents(t, changeContentsA, "a", "thing", "CHANGELOG.md")

	changeContentsB := `second B version
first B version
`
	then.FileContents(t, changeContentsB, "b", "thing", "CHANGELOG.md")
}

func TestMergeVersionsWithHeaderAndReplacements(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg)

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
	then.WithTempDirConfig(t, cfg)

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
	then.WithTempDirConfig(t, cfg)

	changeContents := "a simple header\n"
	builder := strings.Builder{}

	then.WriteFile(t, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Create,
		core.NewTemplateCache(),
	)
	cmd.DryRun = true
	cmd.Command.SetOut(&builder)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, changeContents, builder.String())
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
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestErrorMergeUnableToReadChanges(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg)

	mockErr := errors.New("bad read dir")
	mockReadDir := func(name string) ([]os.DirEntry, error) {
		return nil, mockErr
	}

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		mockReadDir,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.Err(t, mockErr, err)
}

func TestErrorMergeBadReplacement(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.Replacements[0].Replace = "{{bad....}}"
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("a simple header\n"), cfg.ChangesDir, cfg.HeaderPath)
	then.WriteFile(t, []byte("first version\n"), cfg.ChangesDir, "v0.1.0.md")

	cmd := NewMerge(
		os.ReadFile,
		os.WriteFile,
		os.ReadDir,
		os.Create,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}
