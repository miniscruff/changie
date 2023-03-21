package cmd

/*
import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func initCleanArgs() {
	initForce = false
}

func initConfig() core.Config {
	return core.Config{
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

func TestInitBuildsDefaultSkeleton(t *testing.T) {
	cfg := initConfig()
	_, afs := then.WithAferoFS()

	err := initPipeline(afs.MkdirAll, afs.WriteFile, afs.Exists, cfg)

	then.Nil(t, err)
	then.FileContents(t, afs, defaultHeader, cfg.ChangesDir, cfg.HeaderPath)
	then.FileContents(t, afs, defaultChangelog, cfg.ChangelogPath)
	then.FileContents(t, afs, "", cfg.ChangesDir, cfg.UnreleasedDir, ".gitkeep")
}

func TestInitBuildsConfigIfConfigExistsIfForced(t *testing.T) {
	initForce = true
	cfg := initConfig()
	_, afs := then.WithAferoFSConfig(t, &cfg)

	err := initPipeline(afs.MkdirAll, afs.WriteFile, afs.Exists, cfg)

	t.Cleanup(initCleanArgs)
	then.Nil(t, err)
	then.FileContents(t, afs, defaultHeader, cfg.ChangesDir, cfg.HeaderPath)
	then.FileContents(t, afs, defaultChangelog, cfg.ChangelogPath)
	then.FileContents(t, afs, "", cfg.ChangesDir, cfg.UnreleasedDir, ".gitkeep")
}

func TestErrorInitBadMkdir(t *testing.T) {
	cfg := initConfig()
	_, afs := then.WithAferoFS()
	mockError := errors.New("bad mkdir")
	badMkdir := func(path string, mode os.FileMode) error {
		return mockError
	}

	err := initPipeline(badMkdir, afs.WriteFile, afs.Exists, cfg)
	then.Err(t, mockError, err)
}

func TestErrorInitFileExists(t *testing.T) {
	cfg := initConfig()
	_, afs := then.WithAferoFS()
	badFileExists := func(path string) (bool, error) {
		return true, nil
	}

	err := initPipeline(afs.MkdirAll, afs.WriteFile, badFileExists, cfg)
	then.Err(t, errConfigExists, err)
}

func TestErrorInitUnableToCheckIfFileExists(t *testing.T) {
	cfg := initConfig()
	_, afs := then.WithAferoFS()

	badFileExists := func(path string) (bool, error) {
		return false, errors.New("doesn't matter")
	}

	err := initPipeline(afs.MkdirAll, afs.WriteFile, badFileExists, cfg)
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
			_, afs := then.WithAferoFS()
			mockWriteFile := func(path string, data []byte, perm os.FileMode) error {
				if path == filepath.Join(tc.path...) {
					return mockError
				}
				return nil
			}

			err := initPipeline(afs.MkdirAll, mockWriteFile, afs.Exists, cfg)
			then.Err(t, mockError, err)
		})
	}
}
*/
