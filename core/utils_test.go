package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/miniscruff/changie/then"
)

func utilsTestConfig() *Config {
	return &Config{
		ChangesDir:         "news",
		UnreleasedDir:      "future",
		HeaderPath:         "",
		ChangelogPath:      "news.md",
		VersionExt:         "md",
		VersionFormat:      "## {{.Version}}",
		KindFormat:         "### {{.Kind}}",
		ChangeFormat:       "* {{.Body}}",
		FragmentFileFormat: "",
		Kinds: []KindConfig{
			{Label: "added"},
			{Label: "removed"},
			{Label: "other"},
		},
	}
}

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

func TestGetAllVersionsReturnsAllVersions(t *testing.T) {
	then.WithTempDir(t)

	files := []string{
		"v0.1.0.md",
		"v0.2.0.md",
		"header.md",
		"not-sem-ver.md",
	}
	for _, fp := range files {
		_, err := os.Create(fp)
		then.Nil(t, err)
	}

	config := &Config{
		HeaderPath: "header.md",
		ChangesDir: ".",
	}

	vers, err := GetAllVersions(config, false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0", vers[0].Original())
	then.Equals(t, "v0.1.0", vers[1].Original())
}

func TestGetAllVersionsReturnsAllVersionsInProject(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "patcher", "header.md")
	then.CreateFile(t, "patcher", "v0.1.0.md")
	then.CreateFile(t, "patcher", "v0.2.0.md")
	then.CreateFile(t, "patcher", "not-sem-ver.md")

	config := &Config{
		HeaderPath: "header.md",
		ChangesDir: ".",
	}

	vers, err := GetAllVersions(config, false, "patcher")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0", vers[0].Original())
	then.Equals(t, "v0.1.0", vers[1].Original())
}

func TestGetLatestVersionReturnsMostRecent(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "header.md")
	then.CreateFile(t, "v0.1.0.md")
	then.CreateFile(t, "v0.2.0.md")
	then.CreateFile(t, "not-sem-ver.md")

	config := &Config{
		ChangesDir: ".",
	}

	ver, err := GetLatestVersion(config, false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0", ver.Original())
}

func TestGetLatestReturnsRC(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "header.md")
	then.CreateFile(t, "v0.1.0.md")
	then.CreateFile(t, "v0.2.0-rc1.md")
	then.CreateFile(t, "not-sem-ver.md")

	config := &Config{
		ChangesDir: ".",
	}

	ver, err := GetLatestVersion(config, false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0-rc1", ver.Original())
}

func TestGetLatestCanSkipRC(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "header.md")
	then.CreateFile(t, "v0.1.0.md")
	then.CreateFile(t, "v0.2.0-rc1.md")
	then.CreateFile(t, "not-sem-ver.md")

	config := &Config{
		ChangesDir: ".",
	}

	ver, err := GetLatestVersion(config, true, "")
	then.Nil(t, err)
	then.Equals(t, "v0.1.0", ver.Original())
}

func TestGetLatestReturnsZerosIfNoVersionsExist(t *testing.T) {
	then.WithTempDir(t)

	config := &Config{ChangesDir: "."}

	ver, err := GetLatestVersion(config, false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.0.0", ver.Original())
}

func TestErrorLatestVersionBadReadDir(t *testing.T) {
	then.WithTempDir(t)

	config := &Config{ChangesDir: "\\."}

	ver, err := GetLatestVersion(config, false, "")
	then.Equals(t, nil, ver)
	then.NotNil(t, err)
}

func TestErrorNextVersionBadReadDir(t *testing.T) {
	then.WithTempDir(t)

	config := &Config{ChangesDir: "\\."}

	ver, err := GetNextVersion(config, "major", nil, nil, nil, "")
	then.Equals(t, nil, ver)
	then.NotNil(t, err)
}

func TestErrorNextVersionBadVersion(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "v0.1.0.md")

	config := &Config{ChangesDir: "."}

	ver, err := GetNextVersion(config, "a", []string{}, []string{}, nil, "")
	then.Equals(t, ver, nil)
	then.Err(t, ErrBadVersionOrPart, err)
}

func TestNextVersionOptions(t *testing.T) {
	for _, tc := range []struct {
		name          string
		latestVersion string
		partOrVersion string
		prerelease    []string
		meta          []string
		expected      string
	}{
		{
			name:          "BrandNewVersion",
			latestVersion: "v0.2.0",
			partOrVersion: "v0.4.0",
			expected:      "v0.4.0",
		},
		{
			name:          "OnlyMajorVersion",
			latestVersion: "v0.1.5",
			partOrVersion: "v1",
			expected:      "v1",
		},
		{
			name:          "MajorAndMinor",
			latestVersion: "v0.1.5",
			partOrVersion: "v1.2",
			expected:      "v1.2",
		},
		{
			name:          "ShortSemVerPrerelease",
			latestVersion: "v0.1.5",
			partOrVersion: "v1.5",
			prerelease:    []string{"rc2"},
			expected:      "v1.5.0-rc2",
		},
		{
			name:          "MajorPart",
			latestVersion: "v0.1.5",
			partOrVersion: "major",
			expected:      "v1.0.0",
		},
		{
			name:          "MinorPart",
			latestVersion: "v1.1.5",
			partOrVersion: "minor",
			expected:      "v1.2.0",
		},
		{
			name:          "PatchPart",
			latestVersion: "v2.4.2",
			partOrVersion: "patch",
			expected:      "v2.4.3",
		},
		{
			name:          "WithPrerelease",
			latestVersion: "v0.3.5",
			partOrVersion: "patch",
			prerelease:    []string{"b1", "amd64"},
			expected:      "v0.3.6-b1.amd64",
		},
		{
			name:          "WithMeta",
			latestVersion: "v2.4.2",
			partOrVersion: "patch",
			meta:          []string{"20230507", "githash"},
			expected:      "v2.4.3+20230507.githash",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			then.WithTempDir(t)

			_, err := os.Create(tc.latestVersion + ".md")
			then.Nil(t, err)

			config := &Config{
				ChangesDir: ".",
			}

			ver, err := GetNextVersion(config, tc.partOrVersion, tc.prerelease, tc.meta, nil, "")
			then.Nil(t, err)
			then.Equals(t, tc.expected, ver.Original())
		})
	}
}

func TestNextVersionOptionsWithNoneAutoLevel(t *testing.T) {
	then.WithTempDir(t)

	config := &Config{
		ChangesDir: ".",
		Kinds: []KindConfig{
			{
				Label:     "patch",
				AutoLevel: PatchLevel,
			},
			{
				Label:     "minor",
				AutoLevel: MinorLevel,
			},
			{
				Label:     "major",
				AutoLevel: MajorLevel,
			},
			{
				Label:     "skip",
				AutoLevel: NoneLevel,
			},
		},
	}
	latestVersion := "v0.2.3"

	_, err := os.Create(latestVersion + ".md")
	then.Nil(t, err)

	changes := []Change{
		{
			Kind: "minor",
		},
		{
			Kind: "skip",
		},
	}

	ver, err := GetNextVersion(config, "auto", nil, nil, changes, "")
	then.Nil(t, err)
	then.Equals(t, "v0.3.0", ver.Original())
}

func TestNextVersionOptionsNoneAutoLevelOnly(t *testing.T) {
	then.WithTempDir(t)

	latestVersion := "v0.2.3"
	_, err := os.Create(latestVersion + ".md")
	then.Nil(t, err)

	config := &Config{
		ChangesDir: ".",
		Kinds: []KindConfig{
			{
				Label:     "skip",
				AutoLevel: NoneLevel,
			},
		},
	}
	changes := []Change{
		{
			Kind: "skip",
		},
	}

	ver, err := GetNextVersion(config, "auto", nil, nil, changes, "")
	then.Equals(t, ver, nil)
	then.Err(t, ErrNoChangesFoundForAuto, err)
}

func TestErrorNextVersionAutoMissingKind(t *testing.T) {
	then.WithTempDir(t)

	_, err := os.Create("v0.2.3.md")
	then.Nil(t, err)

	config := &Config{
		ChangesDir: ".",
		Kinds: []KindConfig{
			{
				Label: "missing",
			},
		},
	}
	changes := []Change{
		{
			Kind: "missing",
		},
	}

	_, err = GetNextVersion(config, "auto", nil, nil, changes, "")
	then.Err(t, ErrMissingAutoLevel, err)
}

func TestErrorNextVersionBadPrerelease(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "v0.2.5.md")

	config := &Config{ChangesDir: "."}

	_, err := GetNextVersion(config, "patch", []string{"0005"}, nil, nil, "")
	then.NotNil(t, err)
}

func TestErrorNextVersionBadMeta(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "v0.2.5.md")

	config := &Config{ChangesDir: "."}

	_, err := GetNextVersion(config, "patch", nil, []string{"&&*&"}, nil, "")
	then.NotNil(t, err)
}

func TestCanFindChangeFiles(t *testing.T) {
	then.WithTempDir(t)

	then.Nil(t, os.Mkdir(".chng", CreateDirMode))
	then.Nil(t, os.Mkdir(filepath.Join(".chng", "alpha"), CreateDirMode))
	then.Nil(t, os.Mkdir(filepath.Join(".chng", "beta"), CreateDirMode))
	then.Nil(t, os.Mkdir(filepath.Join(".chng", "unrel"), CreateDirMode))
	then.Nil(t, os.Mkdir(filepath.Join(".chng", "ignored"), CreateDirMode))

	expected := []string{
		filepath.Join(".chng", "alpha", "c.yaml"),
		filepath.Join(".chng", "alpha", "d.yaml"),
		filepath.Join(".chng", "beta", "e.yaml"),
		filepath.Join(".chng", "beta", "f.yaml"),
		filepath.Join(".chng", "unrel", "a.yaml"),
		filepath.Join(".chng", "unrel", "b.yaml"),
	}
	for _, fp := range expected {
		_, err := os.Create(fp)
		then.Nil(t, err)
	}

	for _, fp := range []string{
		filepath.Join(".chng", "beta", "ignored.md"),
		filepath.Join(".chng", "ignored", "g.yaml"),
		filepath.Join(".chng", "h.md"),
	} {
		_, err := os.Create(fp)
		then.Nil(t, err)
	}

	config := &Config{
		ChangesDir:    ".chng",
		UnreleasedDir: "unrel",
	}

	files, err := FindChangeFiles(config, []string{"alpha", "beta"})
	then.Nil(t, err)
	then.SliceEquals(t, expected, files)
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

func TestHighestAutoLevel(t *testing.T) {
	cfg := &Config{
		Kinds: []KindConfig{
			{
				Label:     "patch",
				AutoLevel: PatchLevel,
			},
			{
				Label:     "minor",
				AutoLevel: MinorLevel,
			},
			{
				Label:     "major",
				AutoLevel: MajorLevel,
			},
		},
	}

	for _, tc := range []struct {
		name     string
		changes  []Change
		expected string
	}{
		{
			name:     "single patch",
			expected: PatchLevel,
			changes: []Change{
				{
					Kind: "patch",
				},
			},
		},
		{
			name:     "patch and minor",
			expected: MinorLevel,
			changes: []Change{
				{
					Kind: "patch",
				},
				{
					Kind: "minor",
				},
			},
		},
		{
			name:     "major, minor and patch",
			expected: MajorLevel,
			changes: []Change{
				{
					Kind: "patch",
				},
				{
					Kind: "minor",
				},
				{
					Kind: "major",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, err := HighestAutoLevel(cfg, tc.changes)
			then.Nil(t, err)
			then.Equals(t, tc.expected, res)
		})
	}
}

func TestErrorHighestAutoLevelMissingKindConfig(t *testing.T) {
	cfg := &Config{
		Kinds: []KindConfig{
			{
				Label: "missing",
			},
		},
	}
	changes := []Change{
		{
			Kind: "missing",
		},
	}
	_, err := HighestAutoLevel(cfg, changes)
	then.Err(t, ErrMissingAutoLevel, err)
}

func TestErrorHighestAutoLevelWithNoChanges(t *testing.T) {
	cfg := &Config{
		Kinds: []KindConfig{
			{
				Label: "missing",
			},
		},
	}
	changes := []Change{}
	_, err := HighestAutoLevel(cfg, changes)
	then.Err(t, ErrNoChangesFoundForAuto, err)
}

func TestGetAllChanges(t *testing.T) {
	then.WithTempDir(t)

	cfg := utilsTestConfig()

	orderedTimes := []time.Time{
		time.Date(2019, 5, 25, 20, 45, 0, 0, time.UTC),
		time.Date(2017, 4, 25, 15, 20, 0, 0, time.UTC),
		time.Date(2015, 3, 25, 10, 5, 0, 0, time.UTC),
	}

	changes := []Change{
		{Kind: "removed", Body: "third", Time: orderedTimes[2]},
		{Kind: "added", Body: "first", Time: orderedTimes[0]},
		{Kind: "added", Body: "second", Time: orderedTimes[1]},
	}
	for i, c := range changes {
		then.WriteFileTo(t, c, cfg.ChangesDir, cfg.UnreleasedDir, fmt.Sprintf("%d.yaml", i))
	}

	then.CreateFile(t, cfg.ChangesDir, cfg.UnreleasedDir, "ignored.txt")

	changes, err := GetChanges(cfg, nil, "")
	then.Nil(t, err)
	then.SliceLen(t, 3, changes)
	then.Equals(t, "second", changes[0].Body)
	then.Equals(t, "first", changes[1].Body)
	then.Equals(t, "third", changes[2].Body)
}

func TestGetAllChangesWithProject(t *testing.T) {
	then.WithTempDir(t)

	cfg := utilsTestConfig()
	cfg.Projects = []ProjectConfig{
		{
			Label: "Web Hook",
			Key:   "web_hook_sender",
		},
	}
	cfg.Kinds[0].Key = "added"
	cfg.Kinds[0].Label = "Added Label"

	changes := []Change{
		{Kind: "added", Body: "first", Project: "web_hook_sender"},
		{Kind: "added", Body: "second", Project: "web_hook_sender"},
		{Kind: "removed", Body: "ignored", Project: "skipped"},
	}
	for i, c := range changes {
		then.WriteFileTo(t, c, cfg.ChangesDir, cfg.UnreleasedDir, fmt.Sprintf("%d.yaml", i))
	}

	changes, err := GetChanges(cfg, nil, "web_hook_sender")
	then.Nil(t, err)
	then.Equals(t, 2, len(changes))
	then.Equals(t, "first", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
	then.Equals(t, "added", changes[0].KindKey)
	then.Equals(t, "added", changes[1].KindKey)
	then.Equals(t, "Added Label", changes[0].KindLabel)
	then.Equals(t, "Added Label", changes[1].KindLabel)
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
