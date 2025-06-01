package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/then"
)

func newTestConfig() *core.Config {
	return &core.Config{
		ChangesDir:    "news",
		UnreleasedDir: "future",
		HeaderPath:    "header.rst",
		ChangelogPath: "news.md",
		VersionExt:    "md",
		VersionFormat: "## {{.Version}}",
		KindFormat:    "### {{.Kind}}",
		ChangeFormat:  "* {{.Body}}",
		EnvPrefix:     "ENVPREFIX_",
		Kinds: []core.KindConfig{
			{Label: "added"},
			{Label: "removed"},
			{Label: "other"},
		},
	}
}

func newMockTime() time.Time {
	return time.Date(2021, 5, 22, 13, 30, 10, 5, time.UTC)
}

func TestNewWithEnvVars(t *testing.T) {
	cfg := newTestConfig()
	cfg.Post = []core.PostProcessConfig{
		{Key: "TestPost", Value: "{{.Env.TESTCONTENT}}"},
	}
	then.WithTempDirConfig(t, cfg)
	reader, writer := then.WithReadWritePipe(t)

	t.Setenv("ENVPREFIX_TESTCONTENT", "Test content")

	then.DelayWrite(
		t, writer,
		[]byte{106, 13},
		[]byte("a message with testcontent"),
		[]byte{13},
	)

	cmd := NewNew(
		newMockTime,
		core.NewTemplateCache(),
	)
	cmd.SetIn(reader)

	then.Nil(t, os.MkdirAll(filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir), 0755))

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	futurePath := filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir)
	fileInfos, err := os.ReadDir(futurePath)
	then.Nil(t, err)
	then.Equals(t, 1, len(fileInfos))
	then.Equals(t, ".yaml", filepath.Ext(fileInfos[0].Name()))

	changeContent := fmt.Sprintf(
		"kind: removed\nbody: a message with testcontent\ntime: %s\ncustom:\n    TestPost: Test content\n",
		newMockTime().Format(time.RFC3339Nano),
	)

	then.FileExists(t, futurePath, fileInfos[0].Name())
	then.FileContents(t, changeContent, futurePath, fileInfos[0].Name())
}

func TestNewCreatesNewFileAfterPrompts(t *testing.T) {
	cfg := newTestConfig()
	then.WithTempDirConfig(t, cfg)
	reader, writer := then.WithReadWritePipe(t)

	then.DelayWrite(
		t, writer,
		[]byte{106, 13},
		[]byte("a message"),
		[]byte{13},
	)

	cmd := NewNew(
		newMockTime,
		core.NewTemplateCache(),
	)
	cmd.SetIn(reader)

	// Do not create the unreleased dir, let new do it for us.

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	futurePath := filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir)
	fileInfos, err := os.ReadDir(futurePath)
	then.Nil(t, err)
	then.Equals(t, 1, len(fileInfos))
	then.Equals(t, ".yaml", filepath.Ext(fileInfos[0].Name()))

	changeContent := fmt.Sprintf(
		"kind: removed\nbody: a message\ntime: %s\n",
		newMockTime().Format(time.RFC3339Nano),
	)

	then.FileExists(t, futurePath, fileInfos[0].Name())
	then.FileContents(t, changeContent, futurePath, fileInfos[0].Name())
}

func TestErrorNewBadCustomValues(t *testing.T) {
	cfg := newTestConfig()
	then.WithTempDirConfig(t, cfg)

	cmd := NewNew(
		newMockTime,
		core.NewTemplateCache(),
	)
	cmd.Custom = []string{"bad-format"}

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestErrorOnBadConfig(t *testing.T) {
	then.WithTempDir(t)

	cmd := NewNew(
		newMockTime,
		core.NewTemplateCache(),
	)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestErrorNewOnBadPrompts(t *testing.T) {
	cfg := newTestConfig()
	then.WithTempDirConfig(t, cfg)
	reader, writer := then.WithReadWritePipe(t)

	cmd := NewNew(
		newMockTime,
		core.NewTemplateCache(),
	)
	cmd.SetIn(reader)

	then.DelayWrite(
		t, writer,
		[]byte{106, 13},
		[]byte("a message"),
		[]byte{3}, // 3=ctrl+c to quit
	)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestErrorNewFragmentTemplate(t *testing.T) {
	cfg := newTestConfig()
	cfg.FragmentFileFormat = "{{...asdf}}"
	then.WithTempDirConfig(t, cfg)
	reader, writer := then.WithReadWritePipe(t)

	cmd := NewNew(
		newMockTime,
		core.NewTemplateCache(),
	)
	cmd.SetIn(reader)

	then.DelayWrite(
		t, writer,
		[]byte{106, 13},
		[]byte("a message"),
		[]byte{13},
	)

	err := cmd.Run(cmd.Command, nil)
	then.NotNil(t, err)
}

func TestNewOutputsToCmdOutWhenDry(t *testing.T) {
	cfg := newTestConfig()
	cfg.Kinds = []core.KindConfig{}
	then.WithTempDirConfig(t, cfg)
	reader, writer := then.WithReadWritePipe(t)

	outWriter := strings.Builder{}
	cmd := NewNew(
		newMockTime,
		core.NewTemplateCache(),
	)
	cmd.DryRun = true
	cmd.BodyEditor = false
	cmd.SetIn(reader)
	cmd.SetOut(&outWriter)

	then.DelayWrite(
		t, writer,
		[]byte("another body"),
		[]byte{13},
	)

	changeContent := fmt.Sprintf(
		"body: another body\ntime: %s\n",
		newMockTime().Format(time.RFC3339Nano),
	)

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)
	then.Equals(t, changeContent, outWriter.String())
}

func TestNewFragmentTemplateSlash(t *testing.T) {
	cfg := newTestConfig()
	cfg.Components = []string{"test/component"}
	then.WithTempDirConfig(t, cfg)
	reader, writer := then.WithReadWritePipe(t)

	then.DelayWrite(
		t, writer,
		[]byte{106, 13},
		[]byte{106, 13},
		[]byte("a message"),
		[]byte{13},
	)

	cmd := NewNew(
		newMockTime,
		core.NewTemplateCache(),
	)
	cmd.SetIn(reader)

	then.Nil(t, os.MkdirAll(filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir), 0755))

	err := cmd.Run(cmd.Command, nil)
	then.Nil(t, err)

	futurePath := filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir)
	fileInfos, err := os.ReadDir(futurePath)
	then.Nil(t, err)
	then.Equals(t, 1, len(fileInfos))
	then.Equals(t, ".yaml", filepath.Ext(fileInfos[0].Name()))

	changeContent := fmt.Sprintf(
		"component: test/component\nkind: removed\nbody: a message\ntime: %s\n",
		newMockTime().Format(time.RFC3339Nano),
	)

	then.FileExists(t, futurePath, fileInfos[0].Name())
	then.FileContents(t, changeContent, futurePath, fileInfos[0].Name())
}

func TestPromptEnabled(t *testing.T) {
	t.Run("prompts enabled by default", func(t *testing.T) {
		n := NewNew(nil, nil)
		then.True(t, n.parsePromptEnabled())
	})

	t.Run("prompts disabled with flag", func(t *testing.T) {
		n := NewNew(nil, nil)
		n.Interactive = false
		then.False(t, n.parsePromptEnabled())
	})

	t.Run("prompts disabled with CI env var", func(t *testing.T) {
		n := NewNew(nil, nil)

		t.Setenv("CI", "true")
		then.False(t, n.parsePromptEnabled())
	})
}
