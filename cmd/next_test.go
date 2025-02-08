package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func nextTestConfig() *core.Config {
	return &core.Config{
		ChangesDir:    "chgs",
		UnreleasedDir: "unrel",
		HeaderPath:    "head.tpl.md",
		ChangelogPath: "changelog.md",
		VersionExt:    "md",
		VersionFormat: "",
		KindFormat:    "",
		ChangeFormat:  "",
		Kinds:         []core.KindConfig{},
	}
}

func TestNextVersionWithPatch(t *testing.T) {
	cfg := nextTestConfig()
	then.WithTempDirConfig(t, cfg)

	next := NewNext()
	builder := strings.Builder{}

	next.SetOut(&builder)

	// major and minor are not tested directly
	// as next version is tested in utils
	then.CreateFile(t, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "head.tpl.md")

	err := next.Run(next.Command, []string{"patch"})
	then.Nil(t, err)
	then.Equals(t, "v0.1.1", builder.String())
}

func TestNextVersionWithProject(t *testing.T) {
	cfg := nextTestConfig()
	cfg.ProjectsVersionSeparator = "|"
	cfg.Projects = []core.ProjectConfig{
		{
			Label: "W things",
			Key:   "w",
		},
	}
	then.WithTempDirConfig(t, cfg)

	builder := strings.Builder{}
	next := NewNext()
	next.Project = "w"

	next.SetOut(&builder)

	// major and minor are not tested directly
	// as next version is tested in utils
	then.CreateFile(t, cfg.ChangesDir, "w", "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "w", "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "w", "head.tpl.md")

	err := next.Run(next.Command, []string{"patch"})
	then.Nil(t, err)
	then.Equals(t, "w|v0.1.1", builder.String())
}

func TestNextVersionWithProjectBadProject(t *testing.T) {
	cfg := nextTestConfig()
	cfg.ProjectsVersionSeparator = "|"
	cfg.Projects = []core.ProjectConfig{
		{
			Label: "W things",
			Key:   "w",
		},
	}
	then.WithTempDirConfig(t, cfg)

	builder := strings.Builder{}
	next := NewNext()
	next.Project = "missing_proj"

	next.SetOut(&builder)

	// major and minor are not tested directly
	// as next version is tested in utils
	then.CreateFile(t, cfg.ChangesDir, "w", "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "w", "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "w", "head.tpl.md")

	err := next.Run(next.Command, []string{"patch"})
	then.NotNil(t, err)
}

func TestNextVersionWithAuto(t *testing.T) {
	cfg := nextTestConfig()
	cfg.Kinds = []core.KindConfig{
		{
			Label:     "Feature",
			AutoLevel: core.MinorLevel,
		},
	}

	then.WithTempDirConfig(t, cfg)

	builder := strings.Builder{}
	next := NewNext()

	next.SetOut(&builder)
	then.CreateFile(t, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "head.tpl.md")

	var changeBytes bytes.Buffer

	minorChange := core.Change{
		Kind: "Feature",
	}
	_, err := minorChange.WriteTo(&changeBytes)
	then.Nil(t, err)
	then.WriteFile(t, changeBytes.Bytes(), cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	err = next.Run(next.Command, []string{"auto"})
	then.Nil(t, err)
	then.Equals(t, "v0.2.0", builder.String())
}

func TestNextVersionWithPrereleaseAndMeta(t *testing.T) {
	cfg := nextTestConfig()
	then.WithTempDirConfig(t, cfg)

	next := NewNext()
	builder := strings.Builder{}

	next.Prerelease = []string{"b1"}
	next.Meta = []string{"hash"}

	next.SetOut(&builder)
	then.CreateFile(t, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "head.tpl.md")

	err := next.Run(next.Command, []string{"patch"})
	then.Nil(t, err)
	then.Equals(t, "v0.1.1-b1+hash", builder.String())
}

func TestNextVersionWithoutAnyChangesIsV1(t *testing.T) {
	cfg := nextTestConfig()
	then.WithTempDirConfig(t, cfg)

	builder := strings.Builder{}

	next := NewNext()
	next.SetOut(&builder)

	err := next.Run(next.Command, []string{"major"})
	then.Nil(t, err)
	then.Equals(t, "v1.0.0", builder.String())
}

func TestErrorNextVersionBadConfig(t *testing.T) {
	then.WithTempDir(t)

	next := NewNext()
	builder := strings.Builder{}

	next.SetOut(&builder)

	err := next.Run(next.Command, []string{"major"})
	then.NotNil(t, err)
}

func TestErrorNextPartNotSupported(t *testing.T) {
	cfg := nextTestConfig()
	then.WithTempDirConfig(t, cfg)

	next := NewNext()
	builder := strings.Builder{}

	next.SetOut(&builder)
	then.CreateFile(t, cfg.ChangesDir, "v0.0.1.md")

	err := next.Run(next.Command, []string{"notsupported"})
	then.NotNil(t, err)
}

func TestErrorNextUnableToGetChanges(t *testing.T) {
	cfg := nextTestConfig()
	then.WithTempDirConfig(t, cfg)

	next := NewNext()
	builder := strings.Builder{}
	aVer := []byte("not a valid change")

	next.SetOut(&builder)
	then.WriteFile(t, aVer, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	// bad yaml will fail to load changes
	err := next.Run(next.Command, []string{"auto"})
	then.NotNil(t, err)
}
