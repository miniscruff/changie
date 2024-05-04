package cmd

import (
	"testing"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func initConfig() *core.Config {
	return &core.Config{
		ChangesDir:    ".changes",
		UnreleasedDir: "unreleased",
		HeaderPath:    "header.tpl.md",
		ChangelogPath: "CHANGELOG.md",
		VersionExt:    "md",
		VersionFormat: "",
		KindFormat:    "",
		ChangeFormat:  "",
		Kinds:         []core.KindConfig{},
	}
}

func TestInitBuildsDefaultSkeleton(t *testing.T) {
	then.WithTempDir(t)

	cfg := initConfig()
	cmd := NewInit()
	err := cmd.Run(cmd.Command, nil)

	then.Nil(t, err)
	then.FileContents(t, defaultHeader, cfg.ChangesDir, cfg.HeaderPath)
	then.FileContents(t, defaultChangelog, cfg.ChangelogPath)
	then.FileContents(t, "", cfg.ChangesDir, cfg.UnreleasedDir, ".gitkeep")
}

func TestInitBuildsConfigIfConfigExistsIfForced(t *testing.T) {
	cfg := initConfig()
	then.WithTempDirConfig(t, cfg)

	cmd := NewInit()
	cmd.Force = true
	err := cmd.Run(cmd.Command, nil)

	then.Nil(t, err)
	then.FileContents(t, defaultHeader, cfg.ChangesDir, cfg.HeaderPath)
	then.FileContents(t, defaultChangelog, cfg.ChangelogPath)
	then.FileContents(t, "", cfg.ChangesDir, cfg.UnreleasedDir, ".gitkeep")
}

func TestErrorInitFileExists(t *testing.T) {
	cfg := initConfig()
	then.WithTempDirConfig(t, cfg)

	cmd := NewInit()
	err := cmd.Run(cmd.Command, nil)
	then.Err(t, errConfigExists, err)
}
