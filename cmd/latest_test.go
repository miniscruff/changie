package cmd

import (
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

	then.CreateFile(t, afs, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.2.0-rc1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "head.tpl.md")

	res, err := latestPipeline(afs, false)
	then.Nil(t, err)
	then.Equals(t, "v0.2.0-rc1", res)
}

func TestLatestEchoLatestNonPrerelease(t *testing.T) {
	cfg := latestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)

	then.CreateFile(t, afs, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.2.0-rc1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "head.tpl.md")

	res, err := latestPipeline(afs, true)
	then.Nil(t, err)
	then.Equals(t, "v0.1.0", res)
}

func TestLatestWithoutPrefix(t *testing.T) {
	cfg := latestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	removePrefix = true

	t.Cleanup(latestCleanArgs)
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.0.1.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "v0.1.0.md")
	then.CreateFile(t, afs, cfg.ChangesDir, "head.tpl.md")

	res, err := latestPipeline(afs, false)
	then.Nil(t, err)
	then.Equals(t, "0.1.0", res)
}

func TestErrorLatestBadConfig(t *testing.T) {
	_, afs := then.WithAferoFS()

	_, err := latestPipeline(afs, false)
	then.NotNil(t, err)
}

func TestErrorLatestNoVersions(t *testing.T) {
	cfg := latestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)

	// no files, means bad read for get versions
	_, err := latestPipeline(afs, false)
	then.NotNil(t, err)
}
