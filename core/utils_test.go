package core

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/miniscruff/changie/shared"
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

	err = AppendFile(os.Open, rootFile, appendPath)
	then.Nil(t, err)

	rootFile.Close()
	then.FileContents(t, "root append", rootPath)
}

func TestErrorAppendFileIfOpenFails(t *testing.T) {
	mockError := errors.New("bad open")
	builder := &strings.Builder{}
	badOpen := func(filename string) (*os.File, error) {
		return nil, mockError
	}

	err := AppendFile(badOpen, builder, "dummy.txt")
	then.Err(t, mockError, err)
}

func TestGetAllVersionsReturnsAllVersions(t *testing.T) {
	config := &Config{
		HeaderPath: "header.md",
	}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockIsDir: true, MockName: "dir"},
			&then.MockDirEntry{MockName: "header.md"},
			&then.MockDirEntry{MockName: "v0.1.0.md"},
			&then.MockDirEntry{MockName: "v0.2.0.md"},
			&then.MockDirEntry{MockName: "not-sem-ver.md"},
		}, nil
	}

	vers, err := GetAllVersions(mockRead, config, false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0", vers[0].Original())
	then.Equals(t, "v0.1.0", vers[1].Original())
}

func TestGetAllVersionsReturnsAllVersionsInProject(t *testing.T) {
	config := &Config{
		HeaderPath: "header.md",
	}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		then.Equals(t, "patcher", dirname)

		return []os.DirEntry{
			&then.MockDirEntry{MockIsDir: true, MockName: "dir"},
			&then.MockDirEntry{MockName: "header.md"},
			&then.MockDirEntry{MockName: "v0.1.0.md"},
			&then.MockDirEntry{MockName: "v0.2.0.md"},
			&then.MockDirEntry{MockName: "not-sem-ver.md"},
		}, nil
	}

	vers, err := GetAllVersions(mockRead, config, false, "patcher")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0", vers[0].Original())
	then.Equals(t, "v0.1.0", vers[1].Original())
}

func TestGetLatestVersionReturnsMostRecent(t *testing.T) {
	config := &Config{}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: "v0.1.0.md"},
			&then.MockDirEntry{MockName: "v0.2.0.md"},
		}, nil
	}

	ver, err := GetLatestVersion(mockRead, config, false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0", ver.Original())
}

func TestGetLatestReturnsRC(t *testing.T) {
	config := &Config{}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: "v0.1.0.md"},
			&then.MockDirEntry{MockName: "v0.2.0-rc1.md"},
		}, nil
	}

	ver, err := GetLatestVersion(mockRead, config, false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0-rc1", ver.Original())
}

func TestGetLatestCanSkipRC(t *testing.T) {
	config := &Config{}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: "v0.1.0.md"},
			&then.MockDirEntry{MockName: "v0.2.0-rc1.md"},
		}, nil
	}

	ver, err := GetLatestVersion(mockRead, config, true, "")
	then.Nil(t, err)
	then.Equals(t, "v0.1.0", ver.Original())
}

func TestGetLatestReturnsZerosIfNoVersionsExist(t *testing.T) {
	config := &Config{}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, nil
	}

	ver, err := GetLatestVersion(mockRead, config, false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.0.0", ver.Original())
}

func TestErrorAllVersionsBadReadDir(t *testing.T) {
	config := &Config{}
	mockError := errors.New("bad stuff")
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, mockError
	}

	vers, err := GetAllVersions(mockRead, config, false, "")
	then.Equals(t, len(vers), 0)
	then.Err(t, mockError, err)
}

func TestErrorLatestVersionBadReadDir(t *testing.T) {
	config := &Config{}
	mockError := errors.New("bad stuff")
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, mockError
	}

	ver, err := GetLatestVersion(mockRead, config, false, "")
	then.Equals(t, nil, ver)
	then.Err(t, mockError, err)
}

func TestErrorNextVersionBadReadDir(t *testing.T) {
	config := &Config{}
	mockError := errors.New("bad stuff")
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, mockError
	}

	ver, err := GetNextVersion(mockRead, config, "major", nil, nil, nil, "")
	then.Equals(t, nil, ver)
	then.Err(t, mockError, err)
}

func TestErrorNextVersionBadVersion(t *testing.T) {
	config := &Config{}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: "v0.1.0.md"},
		}, nil
	}

	ver, err := GetNextVersion(mockRead, config, "a", []string{}, []string{}, nil, "")
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
			config := &Config{}
			mockRead := func(dirname string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&then.MockDirEntry{MockName: tc.latestVersion + ".md"},
				}, nil
			}

			ver, err := GetNextVersion(mockRead, config, tc.partOrVersion, tc.prerelease, tc.meta, nil, "")
			then.Nil(t, err)
			then.Equals(t, tc.expected, ver.Original())
		})
	}
}

func TestNextVersionOptionsWithNoneAutoLEvel(t *testing.T) {
	config := &Config{
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
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: latestVersion + ".md"},
		}, nil
	}
	changes := []Change{
		{
			Kind: "minor",
		},
		{
			Kind: "skip",
		},
	}

	ver, err := GetNextVersion(mockRead, config, "auto", nil, nil, changes, "")
	then.Nil(t, err)
	then.Equals(t, "v0.3.0", ver.Original())
}

func TestNextVersionOptionsNoneAutoLevelOnly(t *testing.T) {
	config := &Config{
		Kinds: []KindConfig{
			{
				Label:     "skip",
				AutoLevel: NoneLevel,
			},
		},
	}
	latestVersion := "v0.2.3"
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: latestVersion + ".md"},
		}, nil
	}
	changes := []Change{
		{
			Kind: "skip",
		},
	}

	ver, err := GetNextVersion(mockRead, config, "auto", nil, nil, changes, "")
	then.Equals(t, ver, nil)
	then.Err(t, ErrNoChangesFoundForAuto, err)
}

func TestErrorNextVersionAutoMissingKind(t *testing.T) {
	config := &Config{
		Kinds: []KindConfig{
			{
				Label: "missing",
			},
		},
	}
	latestVersion := "v0.2.3"
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: latestVersion + ".md"},
		}, nil
	}
	changes := []Change{
		{
			Kind: "missing",
		},
	}

	_, err := GetNextVersion(mockRead, config, "auto", nil, nil, changes, "")
	then.Err(t, ErrMissingAutoLevel, err)
}

func TestErrorNextVersionBadPrerelease(t *testing.T) {
	config := &Config{}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: "v0.2.5.md"},
		}, nil
	}

	_, err := GetNextVersion(mockRead, config, "patch", []string{"0005"}, nil, nil, "")
	then.NotNil(t, err)
}

func TestErrorNextVersionBadMeta(t *testing.T) {
	config := &Config{}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: "v0.2.5.md"},
		}, nil
	}

	_, err := GetNextVersion(mockRead, config, "patch", nil, []string{"&&*&"}, nil, "")
	then.NotNil(t, err)
}

func TestCanFindChangeFiles(t *testing.T) {
	config := &Config{
		ChangesDir:    ".chng",
		UnreleasedDir: "unrel",
	}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		switch dirname {
		case filepath.Join(".chng", "alpha"):
			return []os.DirEntry{
				&then.MockDirEntry{MockName: "c.yaml"},
				&then.MockDirEntry{MockName: "d.yaml"},
			}, nil
		case filepath.Join(".chng", "beta"):
			return []os.DirEntry{
				&then.MockDirEntry{MockName: "e.yaml"},
				&then.MockDirEntry{MockName: "f.yaml"},
				&then.MockDirEntry{MockName: "ignored.md"},
			}, nil
		case filepath.Join(".chng", "unrel"):
			return []os.DirEntry{
				&then.MockDirEntry{MockName: "a.yaml"},
				&then.MockDirEntry{MockName: "b.yaml"},
			}, nil
		case filepath.Join(".chng", "ignored"):
			return []os.DirEntry{
				&then.MockDirEntry{MockName: "g.yaml"},
			}, nil
		case ".chng":
			return []os.DirEntry{
				&then.MockDirEntry{MockName: "h.md"},
			}, nil
		}

		return nil, nil
	}

	expected := []string{
		filepath.Join(".chng", "alpha", "c.yaml"),
		filepath.Join(".chng", "alpha", "d.yaml"),
		filepath.Join(".chng", "beta", "e.yaml"),
		filepath.Join(".chng", "beta", "f.yaml"),
		filepath.Join(".chng", "unrel", "a.yaml"),
		filepath.Join(".chng", "unrel", "b.yaml"),
	}

	files, err := FindChangeFiles(config, mockRead, []string{"alpha", "beta"})
	then.Nil(t, err)
	then.SliceEquals(t, expected, files)
}

func TestErrorOnFindChangeFilesIfBadRead(t *testing.T) {
	mockErr := errors.New("bad read")
	config := &Config{}
	mockRead := func(dirname string) ([]os.DirEntry, error) {
		return nil, mockErr
	}

	_, err := FindChangeFiles(config, mockRead, []string{"alpha", "beta"})
	then.Err(t, mockErr, err)
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
	cfg := utilsTestConfig()
	then.WithTempDirConfig(t, cfg)

	orderedTimes := []time.Time{
		time.Date(2019, 5, 25, 20, 45, 0, 0, time.UTC),
		time.Date(2017, 4, 25, 15, 20, 0, 0, time.UTC),
		time.Date(2015, 3, 25, 10, 5, 0, 0, time.UTC),
	}
	aPath := filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir, "a.yaml")
	bPath := filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir, "b.yaml")
	cPath := filepath.Join(cfg.ChangesDir, cfg.UnreleasedDir, "c.yaml")

	readDir := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: "a.yaml"},
			&then.MockDirEntry{MockName: "b.yaml"},
			&then.MockDirEntry{MockName: "c.yaml"},
		}, nil
	}
	readFile := func(filename string) ([]byte, error) {
		var c Change

		switch filename {
		case aPath:
			c = Change{Kind: "removed", Body: "third", Time: orderedTimes[2]}
		case bPath:
			c = Change{Kind: "added", Body: "first", Time: orderedTimes[0]}
		case cPath:
			c = Change{Kind: "added", Body: "second", Time: orderedTimes[1]}
		}

		return yaml.Marshal(&c)
	}

	then.CreateFile(t, cfg.ChangesDir, cfg.UnreleasedDir, "ignored.txt")

	changes, err := GetChanges(cfg, nil, readDir, readFile, "")
	then.Nil(t, err)
	then.Equals(t, "second", changes[0].Body)
	then.Equals(t, "first", changes[1].Body)
	then.Equals(t, "third", changes[2].Body)
}

func TestGetAllChangesWithProject(t *testing.T) {
	cfg := utilsTestConfig()
	cfg.Projects = []ProjectConfig{
		{
			Label: "Web Hook",
			Key:   "web_hook_sender",
		},
	}
	then.WithTempDirConfig(t, cfg)

	writeChangeFile(t, cfg, Change{Kind: "added", Body: "first", Project: "web_hook_sender"})
	writeChangeFile(t, cfg, Change{Kind: "added", Body: "second", Project: "web_hook_sender"})
	writeChangeFile(t, cfg, Change{Kind: "removed", Body: "ignored", Project: "skipped"})

	changes, err := GetChanges(cfg, nil, os.ReadDir, os.ReadFile, "web_hook_sender")
	then.Nil(t, err)
	then.Equals(t, 2, len(changes))
	then.Equals(t, "first", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
}

func TestBatchErrorIfUnableToReadDir(t *testing.T) {
	cfg := utilsTestConfig()
	then.WithTempDirConfig(t, cfg)

	mockErr := errors.New("bad mock open")
	readDir := func(dirname string) ([]os.DirEntry, error) {
		return nil, mockErr
	}
	readFile := func(filename string) ([]byte, error) {
		return nil, nil
	}

	then.CreateFile(t, cfg.ChangesDir, cfg.UnreleasedDir, "ignored.txt")

	_, err := GetChanges(cfg, nil, readDir, readFile, "")
	then.Err(t, mockErr, err)
}

func TestBatchErrorBadChangesFile(t *testing.T) {
	cfg := utilsTestConfig()
	then.WithTempDirConfig(t, cfg)

	readDir := func(dirname string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&then.MockDirEntry{MockName: "a.yaml"},
		}, nil
	}
	mockErr := errors.New("bad mock open")
	readFile := func(filename string) ([]byte, error) {
		return nil, mockErr
	}

	then.CreateFile(t, cfg.ChangesDir, cfg.UnreleasedDir, "ignored.txt")

	_, err := GetChanges(cfg, nil, readDir, readFile, "")
	then.Err(t, mockErr, err)
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
	file, err := createTempFile(os.Create, "windows", "txt")
	defer os.Remove(file)

	then.Nil(t, err)
	then.FileContents(t, string(bom), file)
}

func TestCreateTempFileUnableToWriteBom(t *testing.T) {
	var cf shared.CreateFiler = func(filename string) (*os.File, error) {
		return nil, nil // returning nil so that write bom fails
	}

	runt := "windows"
	ext := ".md"

	_, err := createTempFile(cf, runt, ext)
	then.Err(t, os.ErrInvalid, err)
}

func TestCreateTempFileUnableToCreateFile(t *testing.T) {
	var cf shared.CreateFiler = func(filename string) (*os.File, error) {
		return nil, os.ErrPermission // simulating permission denied err
	}

	runt := "windows"
	ext := ".md"

	_, err := createTempFile(cf, runt, ext)
	then.Err(t, os.ErrPermission, err)
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

	body, err := getBodyTextWithEditor(mockRunner, "body.txt", os.ReadFile)
	then.Nil(t, err)
	then.Equals(t, "some body text", body)
}

func TestGetBodyFromEditorBadRunner(t *testing.T) {
	then.WithTempDir(t)

	mockErr := errors.New("bad runner")
	mockRunner := &errRunner{err: mockErr}

	_, err := getBodyTextWithEditor(mockRunner, "body.txt", os.ReadFile)
	then.Err(t, mockErr, err)
}

func TestGetBodyFromEditorBadReadFile(t *testing.T) {
	then.WithTempDir(t)
	mockRunner := &dummyEditorRunner{
		filename: "body.txt",
		body:     []byte("some body text"),
		t:        t,
	}

	_, err := getBodyTextWithEditor(mockRunner, "diff_file.txt", os.ReadFile)
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
