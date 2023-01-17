package core

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/manifoldco/promptui"

	"github.com/miniscruff/changie/then"
)

func TestSaveConfig(t *testing.T) {
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

	mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
		then.Equals(t, ConfigPaths[0], filepath)
		then.Equals(t, configYaml, string(bytes))

		return nil
	}

	err := config.Save(mockWf)
	then.Nil(t, err)
}

func TestSaveConfigWithCustomChoicesAndOptionals(t *testing.T) {
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

	mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
		then.Equals(t, ConfigPaths[0], filepath)
		then.Equals(t, configYaml, string(bytes))

		return nil
	}

	err := config.Save(mockWf)
	then.Nil(t, err)
}

func TestCanCheckIfFileDoesExist(t *testing.T) {
	config := Config{}
	exister := func(path string) (bool, error) {
		return true, nil
	}

	exist, err := config.Exists(exister)
	then.True(t, exist)
	then.Nil(t, err)
}

func TestCanCheckIfFileDoesNotExist(t *testing.T) {
	config := Config{}
	exister := func(path string) (bool, error) {
		return false, nil
	}

	exist, err := config.Exists(exister)
	then.False(t, exist)
	then.Nil(t, err)
}

func TestErrorWhenCheckingFileExists(t *testing.T) {
	mockError := errors.New("mock error")
	config := Config{}
	exister := func(path string) (bool, error) {
		return false, mockError
	}

	_, err := config.Exists(exister)
	then.Err(t, mockError, err)
}

func TestLoadConfigFromPath(t *testing.T) {
	mockRf := func(filepath string) ([]byte, error) {
		if filepath == ConfigPaths[0] {
			return []byte("changesDir: C\nfragmentFileFormat: \"{{.Custom.Issue}}\"\n"), nil
		}

		return nil, os.ErrNotExist
	}

	config, err := LoadConfig(mockRf)
	then.Nil(t, err)
	then.Equals(t, "C", config.ChangesDir)
	then.Equals(t, "{{.Custom.Issue}}", config.FragmentFileFormat)
}

func TestLoadConfigFromAlternatePath(t *testing.T) {
	mockRf := func(filepath string) ([]byte, error) {
		if filepath == ConfigPaths[1] {
			return []byte("changesDir: C\nheaderPath: header.rst\n"), nil
		}

		return nil, os.ErrNotExist
	}

	config, err := LoadConfig(mockRf)
	then.Nil(t, err)
	then.Equals(t, "C", config.ChangesDir)
	then.Equals(t, "header.rst", config.HeaderPath)
}

func TestLoadConfigFromEnvVar(t *testing.T) {
	customPath := "./custom/changie.yaml"
	mockRf := func(filepath string) ([]byte, error) {
		if filepath == customPath {
			return []byte("changesDir: C\nheaderPath: header.rst\n"), nil
		}

		return nil, os.ErrNotExist
	}

	t.Setenv("CHANGIE_CONFIG_PATH", customPath)

	config, err := LoadConfig(mockRf)
	then.Nil(t, err)
	then.Equals(t, "C", config.ChangesDir)
	then.Equals(t, "header.rst", config.HeaderPath)
}

func TestDefaultFragmentTemplateWithKindsAndComponents(t *testing.T) {
	mockRf := func(filepath string) ([]byte, error) {
		if filepath == ConfigPaths[0] {
			return []byte(`kinds:
- label: B
- label: C
- label: E
components:
- A
- D
- G
`), nil
		}

		return nil, os.ErrNotExist
	}
	defaultFragmentFormat := fmt.Sprintf("{{.Component}}-{{.Kind}}-{{.Time.Format \"%v\"}}", timeFormat)

	config, err := LoadConfig(mockRf)
	then.Nil(t, err)
	then.Equals(t, defaultFragmentFormat, config.FragmentFileFormat)
}

func TestDefaultFragmentTemplateWithKinds(t *testing.T) {
	mockRf := func(filepath string) ([]byte, error) {
		if filepath == ConfigPaths[0] {
			return []byte(`kinds:
- label: B
- label: C
- label: E
`), nil
		}

		return nil, os.ErrNotExist
	}
	defaultFragmentFormat := fmt.Sprintf("{{.Kind}}-{{.Time.Format \"%v\"}}", timeFormat)

	config, err := LoadConfig(mockRf)
	then.Nil(t, err)
	then.Equals(t, defaultFragmentFormat, config.FragmentFileFormat)
}

func TestDefaultFragmentTemplateWithoutKindsOrComponents(t *testing.T) {
	mockRf := func(filepath string) ([]byte, error) {
		if filepath == ConfigPaths[0] {
			return []byte("unreleasedDir: unrel"), nil
		}

		return nil, os.ErrNotExist
	}
	defaultFragmentFormat := fmt.Sprintf("{{.Time.Format \"%v\"}}", timeFormat)

	config, err := LoadConfig(mockRf)
	then.Nil(t, err)
	then.Equals(t, defaultFragmentFormat, config.FragmentFileFormat)
}

func TestErrorBadRead(t *testing.T) {
	mockErr := errors.New("bad file")
	mockRf := func(filepath string) ([]byte, error) {
		return []byte(""), mockErr
	}

	_, err := LoadConfig(mockRf)
	then.Err(t, mockErr, err)
}

func TestErrorBadYamlFile(t *testing.T) {
	mockRf := func(filepath string) ([]byte, error) {
		return []byte("not a valid yaml file---"), nil
	}

	_, err := LoadConfig(mockRf)
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
	p := BodyConfig{}.CreatePrompt(os.Stdin)
	underPrompt, ok := p.(*promptui.Prompt)

	then.True(t, ok)
	then.Equals(t, "Body", underPrompt.Label.(string))
	then.Nil(t, underPrompt.Validate("anything"))
}

func TestBodyConfigWithMinAndMax(t *testing.T) {
	var max int64 = 10

	longInput := "jas dklfjaklsd fjklasjd flkasjdfkl sd"
	p := BodyConfig{MaxLength: &max}.CreatePrompt(os.Stdin)
	underPrompt, ok := p.(*promptui.Prompt)

	then.True(t, ok)
	then.Equals(t, "Body", underPrompt.Label.(string))
	then.Err(t, errInputTooLong, underPrompt.Validate(longInput))
}
