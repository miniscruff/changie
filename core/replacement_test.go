package core

import (
	"errors"
	"os"
	"testing"

	"github.com/miniscruff/changie/then"
)

func TestFindAndReplaceContentsInAFile(t *testing.T) {
    then.WithTempDir(t)
	filepath := "file.txt"
	startData := `first line
second line
third line
ignore me`
	endData := `first line
replaced here
third line
ignore me`

	err := os.WriteFile(filepath, []byte(startData), os.ModePerm)
	then.Nil(t, err)

	rep := Replacement{
		Path:    filepath,
		Find:    "second line",
		Replace: "replaced here",
	}
	err = rep.Execute(os.ReadFile, os.WriteFile, ReplaceData{})
	then.Nil(t, err)
	then.FileContents(t, endData, filepath)
}

func TestFindAndReplaceWithTemplate(t *testing.T) {
    then.WithTempDir(t)
	filepath := "template.txt"
	startData := `{
  "name": "demo-project",
  "version": "1.0.0",
}`
	endData := `{
  "name": "demo-project",
  "version": "1.1.0",
}`

	err := os.WriteFile(filepath, []byte(startData), os.ModePerm)
	then.Nil(t, err)

	rep := Replacement{
		Path:    filepath,
		Find:    `  "version": ".*",`,
		Replace: `  "version": "{{.VersionNoPrefix}}",`,
	}
	err = rep.Execute(os.ReadFile, os.WriteFile, ReplaceData{
		VersionNoPrefix: "1.1.0",
	})
	then.Nil(t, err)
	then.FileContents(t, endData, filepath)
}

func TestFindAndReplaceStartOfLine(t *testing.T) {
    then.WithTempDir(t)
	filepath := "start.txt"
	startData := `# yaml file
version: 0.0.1
level1:
	level2:
		version: 0.0.1
`
	endData := `# yaml file
version: 1.2.3
level1:
	level2:
		version: 0.0.1
`

	err := os.WriteFile(filepath, []byte(startData), os.ModePerm)
	then.Nil(t, err)

	rep := Replacement{
		Path:    filepath,
		Find:    "^version: .*",
		Replace: "version: {{.VersionNoPrefix}}",
	}
	err = rep.Execute(os.ReadFile, os.WriteFile, ReplaceData{
		VersionNoPrefix: "1.2.3",
	})
	then.Nil(t, err)
	then.FileContents(t, endData, filepath)
}

func TestFindAndReplaceCaseInsensitive(t *testing.T) {
    then.WithTempDir(t)
	filepath := "insensitive.txt"
	startData := `# yaml file
Version: 0.0.1
level1:
	level2:
		version: 0.0.1
`
	endData := `# yaml file
version: 1.2.3
level1:
	level2:
		version: 0.0.1
`

	err := os.WriteFile(filepath, []byte(startData), os.ModePerm)
	then.Nil(t, err)

	rep := Replacement{
		Path:    filepath,
		Find:    "^version: .*",
		Replace: "version: {{.VersionNoPrefix}}",
		Flags:   "im",
	}
	err = rep.Execute(os.ReadFile, os.WriteFile, ReplaceData{
		VersionNoPrefix: "1.2.3",
	})
	then.Nil(t, err)
	then.FileContents(t, endData, filepath)
}

func TestErrorBadFileRead(t *testing.T) {
    then.WithTempDir(t)
	rep := Replacement{
		Path: "does not exist",
	}
	err := rep.Execute(os.ReadFile, os.WriteFile, ReplaceData{})
	then.NotNil(t, err)
}

func TestErrorBadTemplateParse(t *testing.T) {
    then.WithTempDir(t)
	rep := Replacement{
		Replace: "{{.bad..}}",
	}
	err := rep.Execute(os.ReadFile, os.WriteFile, ReplaceData{})
	then.NotNil(t, err)
}

func TestErrorBadTemplateExec(t *testing.T) {
    then.WithTempDir(t)
	rep := Replacement{
		Replace: "{{.bad}}",
	}
	err := rep.Execute(os.ReadFile, os.WriteFile, ReplaceData{})
	then.NotNil(t, err)
}

func TestErrorBadWriteFile(t *testing.T) {
    then.WithTempDir(t)
	filepath := "err.txt"
	startData := "some data"
	mockError := errors.New("bad write")
	badWrite := func(path string, data []byte, mode os.FileMode) error {
		return mockError
	}

	err := os.WriteFile(filepath, []byte(startData), os.ModePerm)
	then.Nil(t, err)

	rep := Replacement{
		Path:    filepath,
		Find:    "some",
		Replace: "none",
	}
	err = rep.Execute(os.ReadFile, badWrite, ReplaceData{})
	then.Err(t, mockError, err)
}
