package core

import (
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
	err = rep.Execute(ReplaceData{})
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
	err = rep.Execute(ReplaceData{
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
	err = rep.Execute(ReplaceData{
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
	err = rep.Execute(ReplaceData{
		VersionNoPrefix: "1.2.3",
	})
	then.Nil(t, err)
	then.FileContents(t, endData, filepath)
}

func TestErrorBadTemplateParse(t *testing.T) {
	then.WithTempDir(t)

	rep := Replacement{
		Replace: "{{.bad..}}",
	}
	err := rep.Execute(ReplaceData{})
	then.NotNil(t, err)
}

func TestErrorBadRegex(t *testing.T) {
	then.WithTempDir(t)

	filepath := "insensitive.txt"
	startData := `# yaml file
Version: 0.0.1
level1:
	level2:
		version: 0.0.1
`

	err := os.WriteFile(filepath, []byte(startData), os.ModePerm)
	then.Nil(t, err)

	rep := Replacement{
		Path:  filepath,
		Find:  "a(b",
		Flags: "im",
	}
	err = rep.Execute(ReplaceData{
		VersionNoPrefix: "1.2.3",
	})
	then.NotNil(t, err)
}

func TestErrorBadTemplateExec(t *testing.T) {
	then.WithTempDir(t)

	rep := Replacement{
		Replace: "{{.bad}}",
	}
	err := rep.Execute(ReplaceData{})
	then.NotNil(t, err)
}

func TestErrorBadFileRead(t *testing.T) {
	then.WithTempDir(t)

	rep := Replacement{
		Path: "does not exist",
	}
	err := rep.Execute(ReplaceData{})
	then.Err(t, ErrNoReplacementFilesFound, err)
}

func TestGlobs(t *testing.T) {
	then.WithTempDir(t)

	toReplace := []byte(`{
  "version": "1.0.0"
}`)
	then.WriteFile(t, toReplace, "a", "b", "c.json")
	then.WriteFile(t, toReplace, "a", "b", "d.xml")
	then.WriteFile(t, toReplace, "a", "b", "e.jsonl")
	then.WriteFile(t, toReplace, "a", "b", "f.png")
	then.WriteFile(t, toReplace, "a", "c", "g.json")

	rep := Replacement{
		Path:    "a/*/*.json",
		Find:    `  "version": ".*"`,
		Replace: `  "version": "{{.VersionNoPrefix}}"`,
	}

	err := rep.Execute(ReplaceData{
		VersionNoPrefix: "1.1.0",
	})
	then.Nil(t, err)

	changedFile := `{
  "version": "1.1.0"
}`
	unchangedFile := `{
  "version": "1.0.0"
}`

	then.FileContents(t, changedFile, "a", "b", "c.json")
	then.FileContents(t, unchangedFile, "a", "b", "d.xml")
	then.FileContents(t, unchangedFile, "a", "b", "e.jsonl")
	then.FileContents(t, unchangedFile, "a", "b", "f.png")
	then.FileContents(t, changedFile, "a", "c", "g.json")
}

func TestBadGlob(t *testing.T) {
	rep := Replacement{
		Path: `[]`,
	}

	err := rep.Execute(ReplaceData{})
	then.NotNil(t, err)
}
