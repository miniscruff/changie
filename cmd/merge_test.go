package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func mergeTestConfig() *core.Config {
	return &core.Config{
		// RootDir:    "news",
		// Fragment.Dir: "future",
		// HeaderPath:    "header.rst",
		// ChangelogPath: "news.md",
		// VersionExt:    "md",
		// VersionFormat: "## {{.Version}}",
		// KindFormat:    "### {{.Kind}}",
		// ChangeFormat:  "* {{.Body}}",
		// Kinds: []core.KindConfig{
		// {Label: "Added"},
		// {Label: "Removed"},
		// {Label: "Other"},
		// },
		// Newlines: core.NewlinesConfig{
		// BeforeVersion: 1,
		// AfterVersion: 1,
		// AfterChange:            1,
		// BeforeChangelogVersion: 0,
		// AfterChangelogVersion: 1,
		// },
		// Replacements: []core.Replacement{
		// {
		// Path:    "replace.json",
		// Find:    `  "version": ".*",`,
		// Replace: `  "version": "{{.VersionNoPrefix}}",`,
		// },
		// },
	}
}

func TestMergeVersionsSuccessfully(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.Changelog.Header.FilePath = ""
	cfg.Changelog.Replacements = nil
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.RootDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.RootDir, "v0.2.0.md")

	cmd := NewMerge(core.NewTemplateCache())
	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContents := `second version
first version
`
	then.FileContents(t, changeContents, "news.md")
}

func TestMergeVersionsSuccessfullyWithProject(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.ReleaseNotes.Header.FilePath = ""
	cfg.Changelog.Replacements = nil
	cfg.Project = core.ProjectConfig{
		Options: []core.ProjectOptions{
			{
				Label: "A thing",
				Key:   "a",
				Changelog: core.ChangelogConfig{
					Output: "a/thing/CHANGELOG.md",
				},
				Replacements: []core.Replacement{
					{
						Path:    "a/VERSION",
						Find:    "version",
						Replace: "{{.Version}}",
					},
				},
			},
		},
	}
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.RootDir, "a", "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.RootDir, "a", "v0.2.0.md")
	then.WriteFile(t, []byte("version\n"), "a", "VERSION")

	cmd := NewMerge(core.NewTemplateCache())

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContents := `second version
first version
`
	then.FileContents(t, changeContents, "a", "thing", "CHANGELOG.md")
	then.FileContents(t, "v0.2.0\n", "a", "VERSION")
}

func TestMergeVersionsSuccessfullyWithProjectAndNoChanges(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.ReleaseNotes.Header.FilePath = ""
	cfg.Changelog.Replacements = nil
	cfg.Project = core.ProjectConfig{
		Options: []core.ProjectOptions{
			{
				Label: "A thing",
				Key:   "a",
				Changelog: core.ChangelogConfig{
					Output: "a/thing/CHANGELOG.md",
				},
			},
		},
	}
	then.WithTempDirConfig(t, cfg)

	cmd := NewMerge(core.NewTemplateCache())

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContents := ``
	then.FileContents(t, changeContents, "a", "thing", "CHANGELOG.md")
}

func TestMergeVersionsWithUnreleasedChanges(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.ReleaseNotes.Header.FilePath = ""
	cfg.Changelog.Replacements = nil
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.RootDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.RootDir, "v0.2.0.md")

	unrel := core.Change{
		Kind: "Added",
		Body: "new feature coming soon",
	}
	writeChangeFile(t, cfg, &unrel)

	cmd := NewMerge(core.NewTemplateCache())
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
	cfg.ReleaseNotes.Header.FilePath = ""
	cfg.Changelog.Replacements = nil
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.RootDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.RootDir, "v0.2.0.md")

	aVer := []byte("not a valid change")
	then.WriteFile(t, aVer, cfg.RootDir, cfg.Fragment.Dir, "a.yaml")

	cmd := NewMerge(core.NewTemplateCache())
	cmd.UnreleasedHeader = "## Coming Soon"
	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestMergeVersionsWithUnreleasedChangesErrorsOnBadChangeFormat(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.ReleaseNotes.Header.FilePath = ""
	cfg.Changelog.Replacements = nil
	cfg.Change.Prompt.Format = "{{...invalid format{{{"
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first version\n"), cfg.RootDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.RootDir, "v0.2.0.md")

	unrel := core.Change{
		Kind: "Added",
		Body: "new feature coming soon",
	}
	writeChangeFile(t, cfg, &unrel)

	cmd := NewMerge(core.NewTemplateCache())
	cmd.UnreleasedHeader = "## Coming Soon"
	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestMergeVersionsWithUnreleasedChangesInOneProject(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.ReleaseNotes.Header.FilePath = ""
	cfg.Changelog.Replacements = nil
	cfg.Project = core.ProjectConfig{
		Options: []core.ProjectOptions{
			{
				Label: "A thing",
				Key:   "a",
				Changelog: core.ChangelogConfig{
					Output: "a/thing/CHANGELOG.md",
				},
			},
			{
				Label: "B thing",
				Key:   "b",
				Changelog: core.ChangelogConfig{
					Output: "b/thing/CHANGELOG.md",
				},
			},
		},
	}
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("first A version\n"), cfg.RootDir, "a", "v0.1.0.md")
	then.WriteFile(t, []byte("second A version\n"), cfg.RootDir, "a", "v0.2.0.md")
	then.Nil(t, os.MkdirAll(filepath.Join("a", "thing"), core.CreateDirMode))
	then.WriteFile(t, []byte("first B version\n"), cfg.RootDir, "b", "v0.1.0.md")
	then.WriteFile(t, []byte("second B version\n"), cfg.RootDir, "b", "v0.2.0.md")
	then.Nil(t, os.MkdirAll(filepath.Join("b", "thing"), core.CreateDirMode))

	unrel := core.Change{
		Kind:    "Added",
		Body:    "new feature coming soon",
		Project: "a",
	}
	writeChangeFile(t, cfg, &unrel)

	cmd := NewMerge(core.NewTemplateCache())
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

	then.WriteFile(t, []byte("first version\n"), cfg.RootDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.RootDir, "v0.2.0.md")
	then.WriteFile(t, []byte("ignored\n"), cfg.RootDir, "ignored.txt")
	then.WriteFile(t, []byte("a simple header\n"), cfg.RootDir, cfg.ReleaseNotes.Header.FilePath)
	then.WriteFile(t, []byte(jsonContents), "replace.json")

	cmd := NewMerge(core.NewTemplateCache())
	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	changeContents := `a simple header
second version
first version
`
	then.FileContents(t, changeContents, cfg.Changelog.Output)

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

	then.WriteFile(t, []byte("first version\n"), cfg.RootDir, "v0.1.0.md")
	then.WriteFile(t, []byte("second version\n"), cfg.RootDir, "v0.2.0.md")
	then.WriteFile(t, []byte("ignored\n"), cfg.RootDir, "ignored.txt")
	then.WriteFile(t, []byte("a simple header\n"), cfg.RootDir, cfg.ReleaseNotes.Header.FilePath)

	cmd := NewMerge(core.NewTemplateCache())
	cmd.DryRun = true
	cmd.SetOut(&writer)
	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, changeContents, writer.String())
}

func TestMergeSkipsVersionsIfNoneFound(t *testing.T) {
	cfg := mergeTestConfig()
	then.WithTempDirConfig(t, cfg)

	changeContents := "a simple header\n"
	builder := strings.Builder{}

	then.WriteFile(t, []byte("a simple header\n"), cfg.RootDir, cfg.ReleaseNotes.Header.FilePath)

	cmd := NewMerge(core.NewTemplateCache())
	cmd.DryRun = true
	cmd.SetOut(&builder)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, changeContents, builder.String())
}

func TestErrorMergeBadConfig(t *testing.T) {
	then.WithTempDir(t)

	cmd := NewMerge(core.NewTemplateCache())
	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestErrorMergeBadReplacement(t *testing.T) {
	cfg := mergeTestConfig()
	cfg.Changelog.Replacements[0].Replace = "{{bad....}}"
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("a simple header\n"), cfg.RootDir, cfg.ReleaseNotes.Header.FilePath)
	then.WriteFile(t, []byte("first version\n"), cfg.RootDir, "v0.1.0.md")

	cmd := NewMerge(core.NewTemplateCache())
	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}
