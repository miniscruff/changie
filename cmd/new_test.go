package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"

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
		Kinds: []core.KindConfig{
			{Label: "added"},
			{Label: "removed"},
			{Label: "other"},
		},
	}
}

func newResetVars() {
	_newDryRun = false
	_body = ""
	_kind = ""
	_component = ""
	_custom = nil
}

func newMockTime() time.Time {
	return time.Date(2021, 5, 22, 13, 30, 10, 5, time.UTC)
}

func TestNewCreatesNewFileAfterPrompts(t *testing.T) {
	cfg := newTestConfig()
	_, afs := then.WithAferoFSConfig(t, cfg)
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{106, 13},
		[]byte("a message"),
		[]byte{13},
	)

	err := newPipeline(newConfig{
		afs:           afs,
		timeNow:       newMockTime,
		stdinReader:   reader,
		templateCache: core.NewTemplateCache(),
	})
	then.Nil(t, err)

	futurePath := filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir)
	fileInfos, err := afs.ReadDir(futurePath)
	then.Nil(t, err)
	then.Equals(t, 1, len(fileInfos))
	then.Equals(t, ".yaml", filepath.Ext(fileInfos[0].Name()))

	changeContent := fmt.Sprintf(
		"kind: removed\nbody: a message\ntime: %s\n",
		newMockTime().Format(time.RFC3339Nano),
	)
	changePath := filepath.Join(futurePath, fileInfos[0].Name())
	then.FileContents(t, afs, changePath, changeContent)
}

func TestErrorNewBadCustomValues(t *testing.T) {
	_, afs := then.WithAferoFSConfig(t, newTestConfig())
	_custom = []string{"bad-format"}

	t.Cleanup(newResetVars)

	err := newPipeline(newConfig{
		afs:           afs,
		timeNow:       newMockTime,
		stdinReader:   nil,
		templateCache: core.NewTemplateCache(),
	})
	then.NotNil(t, err)
}

func TestErrorOnBadConfig(t *testing.T) {
	_, afs := then.WithAferoFS()
	err := newPipeline(newConfig{
		afs:           afs,
		timeNow:       newMockTime,
		stdinReader:   os.Stdin,
		templateCache: core.NewTemplateCache(),
	})
	then.NotNil(t, err)
}

func TestErrorNewOnBadWrite(t *testing.T) {
	cfg := newTestConfig()
	cfg.Kinds = []core.KindConfig{}
	fs, afs := then.WithAferoFSConfig(t, cfg)

	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("body stuff"),
		[]byte{13},
	)

	mockError := errors.New("dummy mock error")
	fs.MockCreate = func(name string) (afero.File, error) {
		return nil, mockError
	}

	err := newPipeline(newConfig{
		afs:           afs,
		timeNow:       newMockTime,
		stdinReader:   reader,
		templateCache: core.NewTemplateCache(),
	})
	then.Err(t, mockError, err)
}

func TestErrorNewFragmentTemplate(t *testing.T) {
	cfg := newTestConfig()
	cfg.FragmentFileFormat = "{{...asdf}}"
	_, afs := then.WithAferoFSConfig(t, cfg)

	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{106, 13},
		[]byte("a message"),
		[]byte{13},
	)

	err := newPipeline(newConfig{
		afs:           afs,
		timeNow:       newMockTime,
		stdinReader:   reader,
		templateCache: core.NewTemplateCache(),
	})
	then.NotNil(t, err)
}

func TestNewOutputsToCmdOutWhenDry(t *testing.T) {
	_newDryRun = true

	t.Cleanup(newResetVars)

	cfg := newTestConfig()
	cfg.Kinds = []core.KindConfig{}
	_, afs := then.WithAferoFSConfig(t, cfg)

	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("another body"),
		[]byte{13},
	)

	outWriter := strings.Builder{}

	changeContent := fmt.Sprintf(
		"body: another body\ntime: %s\n",
		newMockTime().Format(time.RFC3339Nano),
	)

	err := newPipeline(newConfig{
		afs:           afs,
		timeNow:       newMockTime,
		stdinReader:   reader,
		templateCache: core.NewTemplateCache(),
		cmdOut:        &outWriter,
	})
	then.Nil(t, err)
	then.Equals(t, changeContent, outWriter.String())

	futurePath := filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir)
	_, err = afs.ReadDir(futurePath)
	then.NotNil(t, err)
}

func TestErrorNewBadBody(t *testing.T) {
	_, afs := then.WithAferoFSConfig(t, newTestConfig())
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{13},
		[]byte("ctrl-c next"),
		[]byte{3},
	)

	err := newPipeline(newConfig{
		afs:           afs,
		timeNow:       newMockTime,
		stdinReader:   reader,
		templateCache: core.NewTemplateCache(),
	})
	then.NotNil(t, err)
}
