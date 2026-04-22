package core

/*
import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/miniscruff/changie/then"
)

func TestSaveConfig(t *testing.T) {
	then.WithTempDir(t)

	cfg := Config{
		ChangesDir:        "Changes",
		UnreleasedDir:     "Unrel",
		HeaderPath:        "header.tpl.md",
		VersionHeaderPath: "header.md",
		VersionExt:        "md",
		ChangeFormat:      "chng",
		ChangelogPath:     "CHANGELOG.md",
	}

	configYaml := `changesDir: Changes
unreleasedDir: Unrel
headerPath: header.tpl.md
changelogPath: CHANGELOG.md
versionExt: md
versionHeaderPath: header.md
changeFormat: chng
`

	err := cfg.Save()
	then.Nil(t, err)
	then.FileContents(t, configYaml, ConfigPaths[0])
}

func TestSaveConfigWithCustomChoicesAndOptionals(t *testing.T) {
	then.WithTempDir(t)

	cfg := Config{
		ChangesDir:        "Changes",
		UnreleasedDir:     "Unrel",
		HeaderPath:        "header.tpl.md",
		VersionHeaderPath: "vheader",
		VersionFooterPath: "vfooter",
		VersionExt:        "md",
		ChangelogPath:     "CHANGELOG.md",
		VersionFormat:     "vers",
		ComponentFormat:   "comp",
		KindFormat:        "kind",
		ChangeFormat:      "chng",
		Components:        []string{"A", "D", "G"},
		HeaderFormat:      "head",
		FooterFormat:      "foot",
		Kinds: []KindConfig{
			{Label: "B"},
			{Label: "C"},
			{Label: "E"},
		},
		CustomChoices: []Custom{
			{
				Key:   "first",
				Type:  CustomString,
				Label: "First name",
			},
		},
	}

	configYaml := `changesDir: Changes
unreleasedDir: Unrel
headerPath: header.tpl.md
changelogPath: CHANGELOG.md
versionExt: md
versionHeaderPath: vheader
versionFooterPath: vfooter
versionFormat: vers
componentFormat: comp
kindFormat: kind
changeFormat: chng
headerFormat: head
footerFormat: foot
components:
    - A
    - D
    - G
kinds:
    - label: B
    - label: C
    - label: E
custom:
    - key: first
      type: string
      label: First name
`

	err := cfg.Save()
	then.Nil(t, err)
	then.FileContents(t, configYaml, ConfigPaths[0])
}

func TestCanCheckIfFileDoesExist(t *testing.T) {
	cfg := &Config{}
	then.WithTempDirConfig(t, cfg)

	exist, err := cfg.Exists()
	then.True(t, exist)
	then.Nil(t, err)
}

func TestCanCheckIfFileDoesNotExist(t *testing.T) {
	cfg := &Config{}

	then.WithTempDir(t)

	exist, err := cfg.Exists()
	then.False(t, exist)
	then.Nil(t, err)
}

func TestLoadConfigFromAlternatePath(t *testing.T) {
	then.WithTempDir(t)

	err := os.WriteFile(ConfigPaths[1], []byte("changesDir: C\nheaderPath: header.rst\n"), CreateFileMode)
	then.Nil(t, err)

	config, err := LoadConfig()
	then.Nil(t, err)
	then.Equals(t, "C", config.ChangesDir)
	then.Equals(t, "header.rst", config.HeaderPath)
}

func TestLoadConfigFromEnvVar(t *testing.T) {
	then.WithTempDir(t)
	t.Setenv("CHANGIE_CONFIG_PATH", filepath.Join("custom", "changie.yaml"))

	then.WriteFile(t, []byte("changesDir: C\nheaderPath: header.rst\n"), "custom", "changie.yaml")

	config, err := LoadConfig()
	then.Nil(t, err)
	then.Equals(t, "C", config.ChangesDir)
	then.Equals(t, "header.rst", config.HeaderPath)
}

func TestLoadConfigFromEnvVarMissingFile(t *testing.T) {
	then.WithTempDir(t)
	t.Setenv("CHANGIE_CONFIG_PATH", filepath.Join("custom", "missing.yaml"))

	_, err := LoadConfig()
	then.NotNil(t, err)
}

func TestLoadConfigMultipleLayersWithoutAConfig(t *testing.T) {
	then.WithTempDir(t)

	then.Nil(t, os.MkdirAll(filepath.Join("a", "b", "c"), CreateDirMode))
	t.Chdir("a")
	t.Chdir("b")
	t.Chdir("c")

	_, err := LoadConfig()
	then.Err(t, ErrConfigNotFound, err)
}

func TestLoadConfigMultipleLayersDown(t *testing.T) {
	then.WithTempDir(t)
	then.WriteFile(t, []byte("changesDir: C\nheaderPath: header.rst\n"), ".changie.yml")

	then.Nil(t, os.MkdirAll(filepath.Join("a", "b", "c"), CreateDirMode))
	t.Chdir("a")
	t.Chdir("b")
	t.Chdir("c")

	config, err := LoadConfig()
	then.Nil(t, err)
	then.Equals(t, "C", config.ChangesDir)
	then.Equals(t, "header.rst", config.HeaderPath)
}

func TestDefaultFragmentTemplateWithProjects(t *testing.T) {
	then.WithTempDir(t)

	configYaml := []byte(`kinds:
  - label: B
  - label: C
  - label: E
components:
  - A
  - D
  - G
projects:
  - label: User Feeds
    key: user_feeds
    changelog: users/feeds/CHANGELOG.md
  - label: User Management
    key: user_management
    changelog: users/management/CHANGELOG.md
`)
	then.WriteFile(t, configYaml, ConfigPaths[0])

	defaultFragmentFormat := fmt.Sprintf("{{.Project}}-{{.Component}}-{{.Kind}}-{{.Time.Format \"%v\"}}", timeFormat)

	config, err := LoadConfig()
	then.Nil(t, err)
	then.Equals(t, defaultFragmentFormat, config.FragmentFileFormat)
}

func TestDefaultFragmentTemplateWithKindsAndComponents(t *testing.T) {
	then.WithTempDir(t)

	yamlBytes := []byte(`kinds:
  - label: B
  - label: C
  - label: E
components:
  - A
  - D
  - G
`)
	then.WriteFile(t, yamlBytes, ConfigPaths[0])

	defaultFragmentFormat := fmt.Sprintf("{{.Component}}-{{.Kind}}-{{.Time.Format \"%v\"}}", timeFormat)

	config, err := LoadConfig()
	then.Nil(t, err)
	then.Equals(t, defaultFragmentFormat, config.FragmentFileFormat)
}

func TestDefaultFragmentTemplateWithKinds(t *testing.T) {
	then.WithTempDir(t)

	cfgYaml := []byte(`kinds:
  - label: B
  - label: C
  - label: E
    `)
	then.WriteFile(t, cfgYaml, ConfigPaths[0])

	defaultFragmentFormat := fmt.Sprintf("{{.Kind}}-{{.Time.Format \"%v\"}}", timeFormat)

	config, err := LoadConfig()
	then.Nil(t, err)
	then.Equals(t, defaultFragmentFormat, config.FragmentFileFormat)
}

func TestDefaultFragmentTemplateWithoutKindsOrComponents(t *testing.T) {
	then.WithTempDir(t)
	then.WriteFile(t, []byte("unreleasedDir: unrel"), ConfigPaths[0])

	defaultFragmentFormat := fmt.Sprintf("{{.Time.Format \"%v\"}}", timeFormat)

	config, err := LoadConfig()
	then.Nil(t, err)
	then.Equals(t, defaultFragmentFormat, config.FragmentFileFormat)
}

func TestErrorBadYamlFile(t *testing.T) {
	then.WithTempDir(t)

	cfgYaml := []byte("not a valid yaml file---")
	then.WriteFile(t, cfgYaml, ConfigPaths[0])

	_, err := LoadConfig()
	then.NotNil(t, err)
}

func TestGetHeaderFromKindLabel(t *testing.T) {
	cfg := Config{
		Kinds: []KindConfig{
			{Label: "A", Format: "unused"},
			{Label: "unused", Format: ""},
			{Label: "C", Format: "KF"},
		},
	}
	format := cfg.KindHeader("C")
	then.Equals(t, "KF", format)
}

func TestGetHeaderFromKindKey(t *testing.T) {
	cfg := Config{
		Kinds: []KindConfig{
			{Label: "A", Format: "unused"},
			{Label: "unused", Format: ""},
			{Key: "C", Format: "KF"},
		},
	}
	format := cfg.KindHeader("C")
	then.Equals(t, "KF", format)
}

func TestGetDefaultHeaderForKind(t *testing.T) {
	cfg := Config{
		Kinds: []KindConfig{
			{Label: "ignored"},
		},
		KindFormat: "KF",
	}
	format := cfg.KindHeader("unused label")
	then.Equals(t, "KF", format)
}

func TestGetChangeFormatFromKindLabel(t *testing.T) {
	cfg := Config{
		Kinds: []KindConfig{
			{Label: "A", ChangeFormat: "unused"},
			{Label: "unused", ChangeFormat: ""},
			{Label: "C", ChangeFormat: "CF"},
		},
	}
	format := cfg.ChangeFormatForKind("C")
	then.Equals(t, "CF", format)
}

func TestGetChangeFormatFromKindKey(t *testing.T) {
	cfg := Config{
		Kinds: []KindConfig{
			{Label: "A", ChangeFormat: "unused"},
			{Label: "unused", ChangeFormat: ""},
			{Key: "C", ChangeFormat: "CF"},
		},
	}
	format := cfg.ChangeFormatForKind("C")
	then.Equals(t, "CF", format)
}

func TestGetDefaultChangeFormatIfNoCustomOnesExist(t *testing.T) {
	cfg := Config{
		Kinds: []KindConfig{
			{Label: "unused", ChangeFormat: "ignored"},
		},
		ChangeFormat: "CF",
	}
	format := cfg.ChangeFormatForKind("C")
	then.Equals(t, "CF", format)
}

func TestBodyConfigCreatePrompt(t *testing.T) {
	cust := BodyConfig{}.CreateCustom()

	then.Equals(t, "Body", cust.Label)
	then.Equals(t, CustomString, cust.Type)
}

func TestBodyConfigCreateBlockPrompt(t *testing.T) {
	cust := BodyConfig{
		UseBlock: true,
	}.CreateCustom()

	then.Equals(t, "Body", cust.Label)
	then.Equals(t, CustomBlock, cust.Type)
}

func TestBodyConfigWithMinAndMax(t *testing.T) {
	var max int64 = 10

	cust := BodyConfig{MaxLength: &max}.CreateCustom()
	then.Equals(t, 10, *cust.MaxLength)
}

func TestConfigProjectFindsProjectByLabel(t *testing.T) {
	cfg := &Config{
		Projects: []ProjectConfig{
			{
				Label: "Other",
				Key:   "other",
			},
			{
				Label: "CLI",
				Key:   "cli",
			},
		},
	}

	proj, err := cfg.Project("CLI")
	then.Nil(t, err)
	then.Equals(t, "cli", proj.Key)
}

func TestConfigProjectFindsProjectByKey(t *testing.T) {
	cfg := &Config{
		Projects: []ProjectConfig{
			{
				Label: "Other",
				Key:   "other",
			},
			{
				Label: "CLI",
				Key:   "cli",
			},
		},
	}

	proj, err := cfg.Project("cli")
	then.Nil(t, err)
	then.Equals(t, "CLI", proj.Label)
}

func TestConfigProjectNoProjectsReturnsEmptyConfig(t *testing.T) {
	cfg := &Config{}

	proj, err := cfg.Project("")
	then.Nil(t, err)
	then.Equals(t, "", proj.Label)
	then.Equals(t, "", proj.Key)
	then.Equals(t, "", proj.ChangelogPath)
}

func TestConfigProjectCanGetLabels(t *testing.T) {
	cfg := &Config{
		Projects: []ProjectConfig{
			{
				Label: "Labeled",
				Key:   "labeled",
			},
			{
				Key: "key_only",
			},
		},
	}

	labels := cfg.ProjectLabels()
	then.SliceEquals(t, []string{"Labeled", "key_only"}, labels)
}

func TestConfigProjectErrorNoLabelOrKeyAndRequired(t *testing.T) {
	cfg := &Config{
		Projects: []ProjectConfig{
			{
				Label: "Other",
				Key:   "other",
			},
		},
	}

	_, err := cfg.Project("")
	then.Err(t, errProjectRequired, err)
}

func TestConfigProjectErrorNotFound(t *testing.T) {
	cfg := &Config{
		Projects: []ProjectConfig{
			{
				Label: "Other",
				Key:   "other",
			},
		},
	}

	_, err := cfg.Project("emailer")
	then.Err(t, errProjectNotFound, err)
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

	cfg := &Config{
		HeaderPath: "header.md",
		ChangesDir: ".",
	}

	vers, err := cfg.AllVersions(false, "")
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

	cfg := &Config{
		HeaderPath: "header.md",
		ChangesDir: ".",
	}

	vers, err := cfg.AllVersions(false, "patcher")
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

	cfg := &Config{
		ChangesDir: ".",
	}

	ver, err := cfg.LatestVersion(false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0", ver.Original())
}

func TestGetLatestReturnsRC(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "header.md")
	then.CreateFile(t, "v0.1.0.md")
	then.CreateFile(t, "v0.2.0-rc1.md")
	then.CreateFile(t, "not-sem-ver.md")

	cfg := &Config{
		ChangesDir: ".",
	}

	ver, err := cfg.LatestVersion(false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.2.0-rc1", ver.Original())
}

func TestGetLatestCanSkipRC(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "header.md")
	then.CreateFile(t, "v0.1.0.md")
	then.CreateFile(t, "v0.2.0-rc1.md")
	then.CreateFile(t, "not-sem-ver.md")

	cfg := &Config{
		ChangesDir: ".",
	}

	ver, err := cfg.LatestVersion(true, "")
	then.Nil(t, err)
	then.Equals(t, "v0.1.0", ver.Original())
}

func TestGetLatestReturnsZerosIfNoVersionsExist(t *testing.T) {
	then.WithTempDir(t)

	cfg := &Config{ChangesDir: "."}

	ver, err := cfg.LatestVersion(false, "")
	then.Nil(t, err)
	then.Equals(t, "v0.0.0", ver.Original())
}

func TestErrorNextVersionBadReadDir(t *testing.T) {
	then.WithTempDir(t)

	cfg := &Config{ChangesDir: "\\."}

	ver, err := cfg.NextVersion("major", nil, nil, nil, "")
	then.Equals(t, "v1.0.0", ver.Original())
	then.Nil(t, err)
}

func TestErrorNextVersionBadVersion(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "v0.1.0.md")

	cfg := &Config{ChangesDir: "."}

	ver, err := cfg.NextVersion("a", []string{}, []string{}, nil, "")
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

			cfg := &Config{
				ChangesDir: ".",
			}

			ver, err := cfg.NextVersion(tc.partOrVersion, tc.prerelease, tc.meta, nil, "")
			then.Nil(t, err)
			then.Equals(t, tc.expected, ver.Original())
		})
	}
}

func TestNextVersionOptionsWithNoneAutoLevel(t *testing.T) {
	then.WithTempDir(t)

	cfg := &Config{
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

	ver, err := cfg.NextVersion("auto", nil, nil, changes, "")
	then.Nil(t, err)
	then.Equals(t, "v0.3.0", ver.Original())
}

func TestNextVersionOptionsNoneAutoLevelOnly(t *testing.T) {
	then.WithTempDir(t)

	latestVersion := "v0.2.3"
	_, err := os.Create(latestVersion + ".md")
	then.Nil(t, err)

	cfg := &Config{
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

	ver, err := cfg.NextVersion("auto", nil, nil, changes, "")
	then.Equals(t, ver, nil)
	then.Err(t, ErrNoChangesFoundForAuto, err)
}

func TestErrorNextVersionAutoMissingKind(t *testing.T) {
	then.WithTempDir(t)

	_, err := os.Create("v0.2.3.md")
	then.Nil(t, err)

	cfg := &Config{
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

	_, err = cfg.NextVersion("auto", nil, nil, changes, "")
	then.Err(t, ErrMissingAutoLevel, err)
}

func TestErrorNextVersionBadPrerelease(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "v0.2.5.md")

	cfg := &Config{ChangesDir: "."}

	_, err := cfg.NextVersion("patch", []string{"0005"}, nil, nil, "")
	then.NotNil(t, err)
}

func TestErrorNextVersionBadMeta(t *testing.T) {
	then.WithTempDir(t)
	then.CreateFile(t, "v0.2.5.md")

	cfg := &Config{ChangesDir: "."}

	_, err := cfg.NextVersion("patch", nil, []string{"&&*&"}, nil, "")
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

	cfg := &Config{
		ChangesDir:    ".chng",
		UnreleasedDir: "unrel",
	}

	files, err := cfg.ChangeFiles([]string{"alpha", "beta"})
	then.Nil(t, err)
	then.SliceEquals(t, expected, files)
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
			res, err := cfg.HighestAutoLevel(tc.changes)
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
	_, err := cfg.HighestAutoLevel(changes)
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
	_, err := cfg.HighestAutoLevel(changes)
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

	changes, err := cfg.Changes(nil, "")
	then.Nil(t, err)
	then.SliceLen(t, 3, changes)
	then.Equals(t, "second", changes[0].Body)
	then.Equals(t, "first", changes[1].Body)
	then.Equals(t, "third", changes[2].Body)
}

func TestGetAllChangesErrorIfNoKindWhenConfigured(t *testing.T) {
	then.WithTempDir(t)

	cfg := utilsTestConfig()
	orderedTimes := []time.Time{
		time.Date(2019, 5, 25, 20, 45, 0, 0, time.UTC),
	}

	changes := []Change{
		{Kind: "added not found", Body: "ignored", Time: orderedTimes[0]},
	}
	for i, c := range changes {
		then.WriteFileTo(t, c, cfg.ChangesDir, cfg.UnreleasedDir, fmt.Sprintf("%d.yaml", i))
	}

	then.CreateFile(t, cfg.ChangesDir, cfg.UnreleasedDir, "ignored.txt")

	_, err := cfg.Changes(nil, "")
	then.Err(t, ErrKindNotFound, err)
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

	changes, err := cfg.Changes(nil, "web_hook_sender")
	then.Nil(t, err)
	then.Equals(t, 2, len(changes))
	then.Equals(t, "first", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
	then.Equals(t, "added", changes[0].KindKey)
	then.Equals(t, "added", changes[1].KindKey)
	then.Equals(t, "Added Label", changes[0].KindLabel)
	then.Equals(t, "Added Label", changes[1].KindLabel)
}
*/
