package cmd

import (
	"errors"
	"os"
	"path/filepath"
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
	cmd := NewInit(os.WriteFile)
	err := cmd.Run(cmd.Command, nil)

	then.Nil(t, err)
	then.FileContents(t, defaultHeader, cfg.ChangesDir, cfg.HeaderPath)
	then.FileContents(t, defaultChangelog, cfg.ChangelogPath)
	then.FileContents(t, "", cfg.ChangesDir, cfg.UnreleasedDir, ".gitkeep")
}

func TestInitBuildsConfigIfConfigExistsIfForced(t *testing.T) {
	cfg := initConfig()
	then.WithTempDirConfig(t, cfg)

	cmd := NewInit(os.WriteFile)
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

	cmd := NewInit(os.WriteFile)
	err := cmd.Run(cmd.Command, nil)
	then.Err(t, errConfigExists, err)
}

func TestErrorInitBadWriteFiles(t *testing.T) {
	cfg := initConfig()
	mockError := errors.New("bad write file")

	for _, tc := range []struct {
		name string
		path []string
	}{
		{
			name: "ChangelogPath",
			path: []string{cfg.ChangelogPath},
		},
		{
			name: "ConfigFile",
			path: []string{core.ConfigPaths[0]},
		},
		{
			name: "HeaderFile",
			path: []string{cfg.ChangesDir, cfg.HeaderPath},
		},
		{
			name: "GitKeep",
			path: []string{cfg.ChangesDir, cfg.UnreleasedDir, ".gitkeep"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			then.WithTempDir(t)

			mockWriteFile := func(path string, data []byte, perm os.FileMode) error {
				t.Log(path, filepath.Join(tc.path...))

				if path == filepath.Join(tc.path...) {
					return mockError
				}

				return nil
			}

			cmd := NewInit(mockWriteFile)
			err := cmd.Run(cmd.Command, nil)
			then.Err(t, mockError, err)
		})
	}
}
