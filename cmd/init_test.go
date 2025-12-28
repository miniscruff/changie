package cmd

import (
	"testing"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func initConfig() *core.Config {
	return &core.Config{
		// RootDir:    ".changes",
		// UnreleasedDir: "unreleased",
		// HeaderPath:    "header.tpl.md",
		// ChangelogPath: "CHANGELOG.md",
		// VersionExt:    "md",
		// VersionFormat: "",
		// KindFormat:    "",
		// ChangeFormat:  "",
		// Kinds:         []core.KindConfig{},
	}
}

func TestInitBuildsDefaultSkeleton(t *testing.T) {
	then.WithTempDir(t)

	cfg := initConfig()
	cmd := NewInit()
	err := cmd.Run(cmd.Command, nil)

	then.Nil(t, err)
	then.FileContents(t, defaultHeader, cfg.RootDir, cfg.Changelog.Header.FilePath)
	then.FileContents(t, defaultChangelog, cfg.Changelog.Output)
	then.FileContents(t, "", cfg.RootDir, cfg.Fragment.Dir, ".gitkeep")
}

func TestInitBuildsConfigIfConfigExistsIfForced(t *testing.T) {
	cfg := initConfig()
	then.WithTempDirConfig(t, cfg)

	cmd := NewInit()
	cmd.Force = true
	err := cmd.Run(cmd.Command, nil)

	then.Nil(t, err)
	then.FileContents(t, defaultHeader, cfg.RootDir, cfg.Changelog.Header.FilePath)
	then.FileContents(t, defaultChangelog, cfg.Changelog.Output)
	then.FileContents(t, "", cfg.RootDir, cfg.Fragment.Dir, ".gitkeep")
}

func TestErrorInitFileExists(t *testing.T) {
	cfg := initConfig()
	then.WithTempDirConfig(t, cfg)

	cmd := NewInit()
	err := cmd.Run(cmd.Command, nil)
	then.Err(t, errConfigExists, err)
}
