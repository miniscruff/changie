package core

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/miniscruff/changie/then"
)

func TestSaveConfig(t *testing.T) {
	then.WithTempDir(t)

	config := Config{
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

	err := config.Save()
	then.Nil(t, err)
	then.FileContents(t, configYaml, ConfigPaths[0])
}

func TestSaveConfigWithCustomChoicesAndOptionals(t *testing.T) {
	then.WithTempDir(t)

	config := Config{
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

	err := config.Save()
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
	then.Nil(t, os.Chdir("a"))
	then.Nil(t, os.Chdir("b"))
	then.Nil(t, os.Chdir("c"))

	_, err := LoadConfig()
	then.Err(t, ErrConfigNotFound, err)
}

func TestLoadConfigMultipleLayersDown(t *testing.T) {
	then.WithTempDir(t)
	then.WriteFile(t, []byte("changesDir: C\nheaderPath: header.rst\n"), ".changie.yml")

	then.Nil(t, os.MkdirAll(filepath.Join("a", "b", "c"), CreateDirMode))
	then.Nil(t, os.Chdir("a"))
	then.Nil(t, os.Chdir("b"))
	then.Nil(t, os.Chdir("c"))

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
	config := Config{
		Kinds: []KindConfig{
			{Label: "A", Format: "unused"},
			{Label: "unused", Format: ""},
			{Label: "C", Format: "KF"},
		},
	}
	format := config.KindHeader("C")
	then.Equals(t, "KF", format)
}

func TestGetDefaultHeaderForKind(t *testing.T) {
	config := Config{
		Kinds: []KindConfig{
			{Label: "ignored"},
		},
		KindFormat: "KF",
	}
	format := config.KindHeader("unused label")
	then.Equals(t, "KF", format)
}

func TestGetChangeFormatFromKindLabel(t *testing.T) {
	config := Config{
		Kinds: []KindConfig{
			{Label: "A", ChangeFormat: "unused"},
			{Label: "unused", ChangeFormat: ""},
			{Label: "C", ChangeFormat: "CF"},
		},
	}
	format := config.ChangeFormatForKind("C")
	then.Equals(t, "CF", format)
}

func TestGetDefaultChangeFormatIfNoCustomOnesExist(t *testing.T) {
	config := Config{
		Kinds: []KindConfig{
			{Label: "unused", ChangeFormat: "ignored"},
		},
		ChangeFormat: "CF",
	}
	format := config.ChangeFormatForKind("C")
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
