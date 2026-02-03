package core

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/miniscruff/changie/then"
)

func TestAppendFileAppendsTwoFiles(t *testing.T) {
	then.WithTempDir(t)

	rootPath := "root.txt"
	appendPath := "append.txt"

	rootFile, err := os.Create(rootPath)
	then.Nil(t, err)

	_, err = rootFile.WriteString("root")
	then.Nil(t, err)

	err = os.WriteFile(appendPath, []byte(" append"), CreateFileMode)
	then.Nil(t, err)

	err = AppendFile(rootFile, appendPath)
	then.Nil(t, err)

	rootFile.Close()
	then.FileContents(t, "root append", rootPath)
}

func TestCanWriteNewLines(t *testing.T) {
	var writer strings.Builder

	err := WriteNewlines(&writer, 3)
	then.Nil(t, err)
	then.Equals(t, "\n\n\n", writer.String())
}

func TestSkipNewLinesIfCountIsZero(t *testing.T) {
	var writer strings.Builder

	err := WriteNewlines(&writer, 0)
	then.Nil(t, err)
	then.Equals(t, "", writer.String())
}

func TestErrorNewLinesOnBadWriter(t *testing.T) {
	writer := then.NewErrWriter()
	writer.Raised(t, WriteNewlines(writer, 3))
}

func TestCanSplitEnvVars(t *testing.T) {
	vars := []string{
		"A=b",
		"B=5+5=10",
		"CHANGIE_NAME=some_name",
	}
	expected := map[string]string{
		"A":            "b",
		"B":            "5+5=10",
		"CHANGIE_NAME": "some_name",
	}

	then.MapEquals(t, expected, EnvVarMap(vars))
}

func TestDoesNotLoadEnvVarsIfNotConfigured(t *testing.T) {
	cfg := Config{}
	vars := []string{
		"A=b",
		"B=5+5=10",
		"CHANGIE_NAME=some_name",
	}

	envs := LoadEnvVars(&cfg, vars)
	then.Equals(t, len(envs), 0)
}

func TestLoadEnvsWithPrefix(t *testing.T) {
	cfg := Config{
		EnvPrefix: "CHANGIE_",
	}
	vars := []string{
		"A=b",
		"B=5+5=10",
		"CHANGIE_NAME=some_name",
	}
	expected := map[string]string{
		"NAME": "some_name",
	}

	envs := LoadEnvVars(&cfg, vars)
	then.MapEquals(t, expected, envs)
}

func TestValidBumpLevel(t *testing.T) {
	for _, lvl := range []string{
		MajorLevel,
		MinorLevel,
		PatchLevel,
		AutoLevel,
	} {
		then.True(t, ValidBumpLevel(lvl))
	}
}

func TestInvalidBumpLevel(t *testing.T) {
	then.False(t, ValidBumpLevel("invalid"))
}

func TestFileExists(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "does_exist.txt")

	exists, err := FileExists("does_exist.txt")
	then.True(t, exists)
	then.Nil(t, err)
}

func TestFileDoesNotExist(t *testing.T) {
	then.WithTempDir(t)

	exists, err := FileExists("does_not_exist.txt")
	then.False(t, exists)
	then.Nil(t, err)
}

func TestFileExistError(t *testing.T) {
	then.WithTempDir(t)

	exists, err := FileExists("\000x")
	then.False(t, exists)
	then.NotNil(t, err)
}

func TestCreateTempFileSuccess(t *testing.T) {
	file, err := createTempFile("windows", "txt")
	defer os.Remove(file)

	then.Nil(t, err)
	then.FileContents(t, string(bom), file)
}

func TestBuildCommandToEditFile(t *testing.T) {
	t.Setenv("EDITOR", "vim")

	cmd, err := BuildCommand("test.md")
	then.Nil(t, err)

	osCmd, ok := cmd.(*exec.Cmd)
	then.True(t, ok)
	then.SliceEquals(t, osCmd.Args, []string{"vim", "test.md"})
}

func TestBuildCommandFailsWithNoEditorConfig(t *testing.T) {
	t.Setenv("EDITOR", "")

	_, err := BuildCommand("test.md")
	then.NotNil(t, err)
}

func TestBuildCommandFailsWithBadEditorConfig(t *testing.T) {
	t.Setenv("EDITOR", "bad shell command\n '")

	_, err := BuildCommand("test.md")
	then.NotNil(t, err)
}

func TestGetBodyFromEditorSuccess(t *testing.T) {
	then.WithTempDir(t)
	mockRunner := &dummyEditorRunner{
		filename: "body.txt",
		body:     []byte("some body text"),
		t:        t,
	}

	body, err := getBodyTextWithEditor(mockRunner, "body.txt")
	then.Nil(t, err)
	then.Equals(t, "some body text", body)
}

func TestGetBodyFromEditorBadRunner(t *testing.T) {
	then.WithTempDir(t)

	mockErr := errors.New("bad runner")
	mockRunner := &errRunner{err: mockErr}

	_, err := getBodyTextWithEditor(mockRunner, "body.txt")
	then.Err(t, mockErr, err)
}

func TestGetBodyFromEditorBadReadFile(t *testing.T) {
	then.WithTempDir(t)
	mockRunner := &dummyEditorRunner{
		filename: "body.txt",
		body:     []byte("some body text"),
		t:        t,
	}

	_, err := getBodyTextWithEditor(mockRunner, "diff_file.txt")
	then.NotNil(t, err)
}

type dummyEditorRunner struct {
	filename string
	body     []byte
	t        *testing.T
}

func (d *dummyEditorRunner) Run() error {
	then.WriteFile(d.t, d.body, d.filename)
	return nil
}

type errRunner struct {
	err error
}

func (r *errRunner) Run() error {
	return r.err
}
