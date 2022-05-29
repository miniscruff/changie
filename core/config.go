package core

import (
	"fmt"
	"io"
	"os"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v2"

	"github.com/miniscruff/changie/shared"
)

const configEnvVar = "CHANGIE_CONFIG_PATH"
const CreateFileMode os.FileMode = 0644
const CreateDirMode os.FileMode = 0755
const timeFormat string = "20060102-150405"

var ConfigPaths []string = []string{
	".changie.yaml",
	".changie.yml",
}

// GetVersions will return, in semver sorted order, all released versions
type GetVersions func(shared.ReadDirer, Config) ([]*semver.Version, error)

type KindConfig struct {
	Label             string   `yaml:",omitempty"`
	Format            string   `yaml:",omitempty"`
	ChangeFormat      string   `yaml:"changeFormat,omitempty"`
	SkipGlobalChoices bool     `yaml:"skipGlobalChoices,omitempty"`
	SkipBody          bool     `yaml:"skipBody,omitempty"`
	AdditionalChoices []Custom `yaml:"additionalChoices,omitempty"`
}

func (kc KindConfig) String() string {
	return kc.Label
}

type BodyConfig struct {
	MinLength *int64 `yaml:"minLength,omitempty"`
	MaxLength *int64 `yaml:"maxLength,omitempty"`
}

func (b BodyConfig) CreatePrompt(stdinReader io.ReadCloser) Prompt {
	p, _ := Custom{
		Label:     "Body",
		Type:      CustomString,
		MinLength: b.MinLength,
		MaxLength: b.MaxLength,
	}.CreatePrompt(stdinReader)

	return p
}

// Config handles configuration for a project.
//
// Custom configuration path:
//
// By default Changie will try and load ".changie.yaml" and ".changie.yml" before running commands.
// If you want to change this path set the environment variable `CHANGIE_CONFIG_PATH` to the desired file.
//
// `export CHANGIE_CONFIG_PATH=./tools/changie.yaml`
//
// **Templating**
//
// Changie utilizes [go template](https://golang.org/pkg/text/template/) and [sprig functions](https://masterminds.github.io/sprig/) for formatting.
// In addition to that, custom template functions are available when working with changes.
//
// When batching changes into a version, the headers and footers are placed as such:
//
// 1. Header file
// 1. Header template
// 1. All changes
// 1. Footer template
// 1. Footer file
//
// All elements are optional and will be added together if all are provided.
type Config struct {
	// Directory for change files, header file and unreleased files.
	// Relative to project root.
	ChangesDir string `yaml:"changesDir" required:"true"`
	// Directory for all unreleased change files.
	// Relative to [changesDir](#config-changesdir).
	UnreleasedDir string `yaml:"unreleasedDir" required:"true"`
	// When merging all versions into one changelog file content is added to the top from a header file.
	// A default header file is created when initializing that follows "Keep a Changelog".
	//
	// Filepath for your changelog header file.
	// Relative to [changesDir](#config-changesdir).
	HeaderPath string `yaml:"headerPath" required:"true"`
	// Filepath for the generated changelog file.
	// Relative to project root.
	ChangelogPath string `yaml:"changelogPath" required:"true"`
	// File extension for generated version files.
	// This should probably match your changelog path file.
	// Must not include the period.
	VersionExt string `yaml:"versionExt" required:"true"`
	// Filepath for your version header file relative to [unreleasedDir](#config-unreleaseddir).
	// It is also possible to use the '--header-path' parameter when using the [batch command](/cli/changie_batch).
	VersionHeaderPath string `yaml:"versionHeaderPath,omitempty"`
	// Filepath for your version footer file relative to [unreleasedDir](#config-unreleaseddir).
	// It is also possible to use the '--footer-path' parameter when using the [batch command](/cli/changie_batch).
	VersionFooterPath string `yaml:"versionFooterPath,omitempty"`
	// Customize the file name generated for new fragments.
	// The default uses the component and kind only if configured for your project.
	// The file is placed in the unreleased directory, so the full path is:
	//
	// "{{.ChangesDir}}/{{.UnreleasedDir}}/{{.FragmentFileFormat}}.yaml"
	// example: yaml
	// fragmentFileFormat: "{{.Kind}}-{{.Custom.Issue}}"
	FragmentFileFormat string `yaml:"fragmentFileFormat,omitempty" default:"{{.Component}}-{{.Kind}}-{{.Time.Format \"20060102-150405\"}}" templateType:"Change"`
	VersionFormat      string `yaml:"versionFormat,omitempty" templateType:"BatchData"`
	ComponentFormat    string `yaml:"componentFormat,omitempty" templateType:"TODO"`
	KindFormat         string `yaml:"kindFormat,omitempty" templateType:"TODO"`
	ChangeFormat       string `yaml:"changeFormat" templateType:"Change"`
	HeaderFormat       string `yaml:"headerFormat,omitempty" templateType:"BatchData"`
	FooterFormat       string `yaml:"footerFormat,omitempty" templateType:"BatchData"`
	// custom
	Body          BodyConfig    `yaml:"body,omitempty"`
	Components    []string      `yaml:"components,omitempty"`
	Kinds         []KindConfig  `yaml:"kinds,omitempty"`
	CustomChoices []Custom      `yaml:"custom,omitempty"`
	Replacements  []Replacement `yaml:"replacements,omitempty"`
}

func (c Config) KindHeader(label string) string {
	for _, kindConfig := range c.Kinds {
		if kindConfig.Format != "" && kindConfig.Label == label {
			return kindConfig.Format
		}
	}

	return c.KindFormat
}

func (c Config) ChangeFormatForKind(label string) string {
	for _, kindConfig := range c.Kinds {
		if kindConfig.ChangeFormat != "" && kindConfig.Label == label {
			return kindConfig.ChangeFormat
		}
	}

	return c.ChangeFormat
}

// Save will save the config as a yaml file to the default path
func (c Config) Save(wf shared.WriteFiler) error {
	bs, _ := yaml.Marshal(&c)
	return wf(ConfigPaths[0], bs, CreateFileMode)
}

// LoadConfig will load the config from the default path
func LoadConfig(rf shared.ReadFiler) (Config, error) {
	var (
		c   Config
		bs  []byte
		err error
	)

	customPath := os.Getenv(configEnvVar)
	if customPath != "" {
		bs, err = rf(customPath)
	} else {
		for _, path := range ConfigPaths {
			bs, err = rf(path)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, err
	}

	// load backward incompatible configs
	if c.FragmentFileFormat == "" {
		if len(c.Components) > 0 {
			c.FragmentFileFormat = "{{.Component}}-"
		}

		if len(c.Kinds) > 0 {
			c.FragmentFileFormat += "{{.Kind}}-"
		}

		c.FragmentFileFormat += fmt.Sprintf("{{.Time.Format \"%v\"}}", timeFormat)
	}

	return c, nil
}
