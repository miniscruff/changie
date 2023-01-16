package cmd

import (
	"strings"
	"testing"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func latestCleanArgs() {
	removePrefix = false
}

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
	_, afs := then.WithAferoFSConfig(t, cfg)
	w := strings.Builder{}

	then.CreateFile(t, afs, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.2.0-rc1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "head.tpl.md")

	err := latestPipeline(afs, &w, false)
	then.Nil(t, err)
	then.Equals(t, "v0.2.0-rc1", w.String())
}

func TestLatestEchoLatestNonPrerelease(t *testing.T) {
	cfg := latestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	w := strings.Builder{}

	then.CreateFile(t, afs, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.2.0-rc1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "head.tpl.md")

	err := latestPipeline(afs, &w, true)
	then.Nil(t, err)
	then.Equals(t, "v0.1.0", w.String())
}

func TestLatestWithoutPrefix(t *testing.T) {
	cfg := latestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	w := strings.Builder{}
	removePrefix = true

	t.Cleanup(latestCleanArgs)
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "head.tpl.md")

	err := latestPipeline(afs, &w, false)
	then.Nil(t, err)
	then.Equals(t, "0.1.0", w.String())
}

func TestErrorLatestBadConfig(t *testing.T) {
	_, afs := then.WithAferoFS()
	w := strings.Builder{}

	err := latestPipeline(afs, &w, false)
	then.NotNil(t, err)
}

func TestErrorLatestNoVersions(t *testing.T) {
	cfg := latestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	w := strings.Builder{}

	// no files, means bad read for get versions
	err := latestPipeline(afs, &w, false)
	then.NotNil(t, err)
}

func TestErrorLatestBadWrite(t *testing.T) {
	cfg := latestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	w := then.NewErrWriter()

	then.CreateFile(t, afs, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.0.md")

	w.Raised(t, latestPipeline(afs, w, false))
}
