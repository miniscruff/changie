package core

import (
	"os"
	"strings"
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

func testGlobs(t *testing.T, paths []string, rep Replacement) {
	toReplace := `{
  "version": "1.0.0"
}`
	replaceWith := "1.2.3"

	for _, path := range paths {
		split := strings.Split(path, "/")
		if split[0] != path {
			err := os.Mkdir(split[0], os.ModePerm)
			then.Nil(t, err)
		}
		err := os.WriteFile(path, []byte(toReplace), os.ModePerm)
		then.Nil(t, err)
	}

	err := rep.Execute(ReplaceData{
		VersionNoPrefix: replaceWith,
	})
	then.Nil(t, err)

	expected := `{
  "version": "` + replaceWith + `"
}`

	for _, path := range paths {
		then.FileContents(t, expected, path)
	}
}

func TestGlobs(t *testing.T) {
	then.WithTempDir(t)

	paths := []string{"a.json", "b.json"}
	rep := Replacement{
		Path:    "*.json",
		Find:    `  "version": ".*"`,
		Replace: `  "version": "{{.VersionNoPrefix}}"`,
	}
	testGlobs(t, paths, rep)

	paths = []string{"c/a.json", "d/b.json"}
	rep.Path = "*/*.json"
	testGlobs(t, paths, rep)

	paths = []string{"e/a.json", "f/b.json"}
	rep.Path = "[ef]/[ab].json"
	testGlobs(t, paths, rep)

	paths = []string{"f.json", "b.jsop", "c.jsof"}
	rep.Path = "*.jso?"
	testGlobs(t, paths, rep)
}

func TestBadGlob(t *testing.T) {
	rep := Replacement{
		Path: `[]`,
	}

	err := rep.Execute(ReplaceData{})
	then.NotNil(t, err)
}
