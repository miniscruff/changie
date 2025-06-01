package cmd

import (
	"strings"
	"testing"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func latestConfig() *core.Config {
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

func TestLatestVersionEchosLatestVersion(t *testing.T) {
	cfg := latestConfig()
	then.WithTempDirConfig(t, cfg)

	then.CreateFile(t, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.2.0-rc1.md")
	then.CreateFile(t, cfg.ChangesDir, "head.tpl.md")

	cmd := NewLatest()

	builder := strings.Builder{}
	cmd.SetOut(&builder)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, "v0.2.0-rc1", builder.String())
}

func TestLatestEchoLatestNonPrerelease(t *testing.T) {
	cfg := latestConfig()
	then.WithTempDirConfig(t, cfg)

	builder := strings.Builder{}

	then.CreateFile(t, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.2.0-rc1.md")
	then.CreateFile(t, cfg.ChangesDir, "head.tpl.md")

	cmd := NewLatest()
	cmd.SkipPrereleases = true
	cmd.SetOut(&builder)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, "v0.1.0", builder.String())
}

func TestLatestWithoutPrefix(t *testing.T) {
	cfg := latestConfig()
	then.WithTempDirConfig(t, cfg)

	builder := strings.Builder{}

	then.CreateFile(t, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "head.tpl.md")

	cmd := NewLatest()
	cmd.RemovePrefix = true
	cmd.SetOut(&builder)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, "0.1.0", builder.String())
}

func TestLatestVersionWithProject(t *testing.T) {
	cfg := latestConfig()
	cfg.Projects = []core.ProjectConfig{
		{
			Label: "r project",
			Key:   "r",
		},
	}
	cfg.ProjectsVersionSeparator = "#"
	then.WithTempDirConfig(t, cfg)

	then.CreateFile(t, cfg.ChangesDir, "r", "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "r", "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "r", "v0.2.0-rc1.md")
	then.CreateFile(t, cfg.ChangesDir, "r", "head.tpl.md")

	builder := strings.Builder{}
	cmd := NewLatest()
	cmd.Project = "r"

	cmd.SetOut(&builder)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, "r#v0.2.0-rc1", builder.String())
}

func TestLatestVersionWithBadProject(t *testing.T) {
	cfg := latestConfig()
	cfg.Projects = []core.ProjectConfig{
		{
			Label: "r project",
			Key:   "r",
		},
	}
	cfg.ProjectsVersionSeparator = "#"
	then.WithTempDirConfig(t, cfg)

	then.CreateFile(t, cfg.ChangesDir, "r", "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "r", "v0.1.0.md")
	then.CreateFile(t, cfg.ChangesDir, "r", "v0.2.0-rc1.md")
	then.CreateFile(t, cfg.ChangesDir, "r", "head.tpl.md")

	cmd := NewLatest()
	cmd.Project = "missing_project_again"

	builder := strings.Builder{}
	cmd.SetOut(&builder)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestErrorLatestBadConfig(t *testing.T) {
	then.WithTempDir(t)

	cmd := NewLatest()
	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestErrorLatestBadWrite(t *testing.T) {
	cfg := latestConfig()
	then.WithTempDirConfig(t, cfg)
	w := then.NewErrWriter()

	then.CreateFile(t, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, cfg.ChangesDir, "v0.1.0.md")

	cmd := NewLatest()
	cmd.SetOut(w)
	w.Raised(t, cmd.Run(cmd.Command, nil))
}
