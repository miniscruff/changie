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
	_, afs := then.WithAferoFSConfig(t, nextTestConfig())
	builder := strings.Builder{}

	// major and minor are not tested directly
	// as next version is tested in utils
	then.CreateFile(t, afs, "chgs", "v0.0.1.md")
	then.CreateFile(t, afs, "chgs", "v0.1.0.md")
	then.CreateFile(t, afs, "chgs", "head.tpl.md")

	err := nextPipeline(afs, &builder, "patch", nil, nil, nil)
	then.Nil(t, err)
	then.Equals(t, "v0.1.1", builder.String())
}

func TestNextVersionWithAuto(t *testing.T) {
	cfg := nextTestConfig()
    cfg.Kinds = []core.KindConfig{
        {
            Label: "Feature",
            AutoLevel: core.MinorLevel,
        },
    }

	builder := strings.Builder{}
	_, afs := then.WithAferoFSConfig(t, cfg)

	then.CreateFile(t, afs, "chgs", "v0.0.1.md")
	then.CreateFile(t, afs, "chgs", "v0.1.0.md")
	then.CreateFile(t, afs, "chgs", "head.tpl.md")

    minorChange := core.Change{
        Kind: "Feature",
    }
    var changeBytes bytes.Buffer
    minorChange.Write(&changeBytes)
    then.WriteFile(t, afs, changeBytes.Bytes(), cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	err := nextPipeline(afs, &builder, "auto", nil, nil, nil)
	then.Nil(t, err)
    then.Equals(t, "v0.2.0", builder.String())
}

func TestNextVersionWithPrereleaseAndMeta(t *testing.T) {
	_, afs := then.WithAferoFSConfig(t, nextTestConfig())
	builder := strings.Builder{}

	then.CreateFile(t, afs, "chgs", "v0.0.1.md")
	then.CreateFile(t, afs, "chgs", "v0.1.0.md")
	then.CreateFile(t, afs, "chgs", "head.tpl.md")

	err := nextPipeline(afs, &builder, "patch", []string{"b1"}, []string{"hash"}, nil)
	then.Nil(t, err)
	then.Equals(t, "v0.1.1-b1+hash", builder.String())
}

func TestErrorNextVersionBadConfig(t *testing.T) {
	_, afs := then.WithAferoFS()
	builder := strings.Builder{}

	err := nextPipeline(afs, &builder, "major", nil, nil, nil)
	then.NotNil(t, err)
}

func TestErrorNextPartNotSupported(t *testing.T) {
	_, afs := then.WithAferoFSConfig(t, nextTestConfig())
	builder := strings.Builder{}

	then.CreateFile(t, afs, "chgs", "v0.0.1.md")

	err := nextPipeline(afs, &builder, "notsupported", nil, nil, nil)
	then.NotNil(t, err)
}

func TestErrorNextUnableToGetChanges(t *testing.T) {
    cfg := nextTestConfig()
	builder := strings.Builder{}
	_, afs := then.WithAferoFSConfig(t, cfg)

	aVer := []byte("not a valid change")
	then.WriteFile(t, afs, aVer, cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")

	// bad yaml will fail to load changes
	err := nextPipeline(afs, &builder, "auto", nil, nil, nil)
	then.NotNil(t, err)
}

func TestErrorNextUnableToGetVersions(t *testing.T) {
	_, afs := then.WithAferoFSConfig(t, nextTestConfig())
	builder := strings.Builder{}

	// no files, means bad read for get versions
	err := nextPipeline(afs, &builder, "major", nil, nil, nil)
	then.NotNil(t, err)
}
