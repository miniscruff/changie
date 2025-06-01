package cmd

import (
	"strings"
	"testing"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func diffConfig() *core.Config {
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

func TestDiff(t *testing.T) {
	cfg := diffConfig()
	then.WithTempDirConfig(t, cfg)

	then.WriteFile(t, []byte("1\n"), cfg.ChangesDir, "v0.1.0.md")
	then.WriteFile(t, []byte("2\n"), cfg.ChangesDir, "v0.2.0.md")
	then.WriteFile(t, []byte("3\n"), cfg.ChangesDir, "v0.3.0.md")
	then.WriteFile(t, []byte("4\n"), cfg.ChangesDir, "v1.0.0.md")
	then.WriteFile(t, []byte("5\n"), cfg.ChangesDir, "v1.1.0.md")

	t.Run("with N", func(t *testing.T) {
		expected := "5\n4\n"

		cmd := NewDiff()

		builder := strings.Builder{}
		cmd.SetOut(&builder)

		err := cmd.Run(cmd.Command, []string{"2"})
		then.Nil(t, err)
		then.Equals(t, expected, builder.String())
	})

	t.Run("with triple dots", func(t *testing.T) {
		expected := "4\n3\n2\n"

		cmd := NewDiff()

		builder := strings.Builder{}
		cmd.SetOut(&builder)

		err := cmd.Run(cmd.Command, []string{"v0.2.0...v1.0.0"})
		then.Nil(t, err)
		then.Equals(t, expected, builder.String())
	})

	t.Run("with constraint", func(t *testing.T) {
		expected := "5\n4\n"

		cmd := NewDiff()

		builder := strings.Builder{}
		cmd.SetOut(&builder)

		err := cmd.Run(cmd.Command, []string{">=v1"})
		then.Nil(t, err)
		then.Equals(t, expected, builder.String())
	})
}
