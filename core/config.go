package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/miniscruff/changie/shared"
)

const (
	configEnvVar        = "CHANGIE_CONFIG_PATH"
	timeFormat   string = "20060102-150405"

	CreateFileMode os.FileMode = 0644
	CreateDirMode  os.FileMode = 0755

	AutoLevel  = "auto"
	MajorLevel = "major"
	MinorLevel = "minor"
	PatchLevel = "patch"
	NoneLevel  = "none"
	EmptyLevel = ""
)

var ConfigPaths []string = []string{
	".changie.yaml",
	".changie.yml",
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
// [Template Cache](#templatecache-type) handles all the template executions.
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
	// TODO: Leave comments for redirects / renames later.
	RootDir      string             `yaml:"rootDir" required:"true"`
	EnvPrefix    string             `yaml:"envPrefix,omitempty"`
	Fragment     FragmentConfig     `yaml:"fragment"`
	Project      ProjectConfig      `yaml:"project"`
	Component    ComponentConfig    `yaml:"component"`
	Kind         KindConfig         `yaml:"kind"`
	Change       ChangeConfig       `yaml:"change"`
	ReleaseNotes ReleaseNotesConfig `yaml:"releaseNotes"`
	Changelog    ChangelogConfig    `yaml:"changelog"`

	// Directory for change files, header file and unreleased files.
	// Relative to project root.
	// example: yaml
	// changesDir: .changes
	ChangesDir string `yaml:"changesDir" required:"true"`
	// Directory for all unreleased change files.
	// Relative to [changesDir](#config-changesdir).
	// example: yaml
	// unreleasedDir: unreleased
	UnreleasedDir string `yaml:"unreleasedDir" required:"true"`
	// Header content included at the top of the merged changelog.
	// A default header file is created when initializing that follows "Keep a Changelog".
	//
	// Filepath for your changelog header file.
	// Relative to [changesDir](#config-changesdir).
	// example: yaml
	// headerPath: header.tpl.md
	HeaderPath string `yaml:"headerPath"`
	// Filepath for the generated changelog file.
	// Relative to project root.
	// ChangelogPath is not required if you are using projects.
	// example: yaml
	// changelogPath: CHANGELOG.md
	ChangelogPath string `yaml:"changelogPath"`
	// File extension for generated version files.
	// This should probably match your changelog path file.
	// Must not include the period.
	// example: yaml
	// # for markdown changelogs
	// versionExt: md
	VersionExt string `yaml:"versionExt" required:"true"`
	// Filepath for your version header file relative to [unreleasedDir](#config-unreleaseddir).
	// It is also possible to use the '--header-path' parameter when using the [batch command](../cli/changie_batch.md).
	VersionHeaderPath string `yaml:"versionHeaderPath,omitempty"`
	// Filepath for your version footer file relative to [unreleasedDir](#config-unreleaseddir).
	// It is also possible to use the '--footer-path' parameter when using the [batch command](../cli/changie_batch.md).
	VersionFooterPath string `yaml:"versionFooterPath,omitempty"`
	// Customize the file name generated for new fragments.
	// The default uses the component and kind only if configured for your project.
	// The file is placed in the unreleased directory, so the full path is:
	//
	// "{{.ChangesDir}}/{{.UnreleasedDir}}/{{.FragmentFileFormat}}.yaml"
	// example: yaml
	// fragmentFileFormat: "{{.Kind}}-{{.Custom.Issue}}"
	FragmentFileFormat string `yaml:"fragmentFileFormat,omitempty" default:"{{.Project}}-{{.Component}}-{{.Kind}}-{{.Time.Format \"20060102-150405\"}}" templateType:"Change"`
	// Template used to generate version headers.
	VersionFormat string `yaml:"versionFormat,omitempty" templateType:"BatchData"`
	// Template used to generate component headers.
	// If format is empty no header will be included.
	// If components are disabled, the format is unused.
	ComponentFormat string `yaml:"componentFormat,omitempty" templateType:"ComponentData"`
	// Template used to generate kind headers.
	// If format is empty no header will be included.
	// If kinds are disabled, the format is unused.
	KindFormat string `yaml:"kindFormat,omitempty" templateType:"KindData"`
	// Template used to generate change lines in version files and changelog.
	// Custom values are created through custom choices and can be accessible through the Custom argument.
	// example: yaml
	// changeFormat: '* [#{{.Custom.Issue}}](https://github.com/miniscruff/changie/issues/{{.Custom.Issue}}) {{.Body}}'
	ChangeFormat string `yaml:"changeFormat" templateType:"Change"`
	// Template used to generate a version header.
	HeaderFormat string `yaml:"headerFormat,omitempty" templateType:"BatchData"`
	// Template used to generate a version footer.
	// example: yaml
	// # config yaml
	// custom:
	// - key: Author
	//   type: string
	//   minLength: 3
	// footerFormat: |
	//   ### Contributors
	//   {{- range (customs .Changes "Author" | uniq) }}
	//   * [{{.}}](https://github.com/{{.}})
	//   {{- end}}
	FooterFormat string `yaml:"footerFormat,omitempty" templateType:"BatchData"`
	// Options to customize the body prompt
	Body BodyConfig `yaml:"body,omitempty"`
	// Components are an additional layer of organization suited for projects that want to split
	// change fragments by an area or tag of the project.
	// An example could be splitting your changelogs by packages for a monorepo.
	// If no components are listed then the component prompt will be skipped and no component header included.
	// By default no components are configured.
	// example: yaml
	// components:
	// - API
	// - CLI
	// - Frontend
	Components []string `yaml:"components,omitempty"`
	// Kinds are another optional layer of changelogs suited for specifying what type of change we are
	// making.
	// If configured, developers will be prompted to select a kind.
	//
	// The default list comes from keep a changelog and includes; added, changed, removed, deprecated, fixed, and security.
	// example: yaml
	// kinds:
	// - label: Added
	// - label: Changed
	// - label: Deprecated
	// - label: Removed
	// - label: Fixed
	// - label: Security
	Kinds []KindConfigOld `yaml:"kinds,omitempty"`
	// Custom choices allow you to ask for additional information when creating a new change fragment.
	// These custom choices are included in the [change custom](#change-custom) value.
	// example: yaml
	// # github issue and author name
	// custom:
	// - key: Issue
	//   type: int
	//   minInt: 1
	// - key: Author
	//   label: GitHub Name
	//   type: string
	//   minLength: 3
	CustomChoices []Custom `yaml:"custom,omitempty"`
	// Replacements to run when merging a changelog.
	// example: yaml
	// # nodejs package.json replacement
	// replacements:
	// - path: package.json
	//   find: '  "version": ".*",'
	//   replace: '  "version": "{{.VersionNoPrefix}}",'
	Replacements []Replacement `yaml:"replacements,omitempty"`
	// Newline options allow you to add extra lines between elements written by changie.
	Newlines NewlinesConfigOld `yaml:"newlines,omitempty"`
	// Post process options when saving a new change fragment.
	// example: yaml
	// # build a GitHub link from author choice
	// post:
	// - key: AuthorLink
	//   value: "https://github.com/{{.Custom.Author}}
	// changeFormat: "* {{.Body}} by [{{.Custom.Author}}]({{.Custom.AuthorLink}})"
	Post []PostProcessConfig `yaml:"post,omitempty"`
	// Projects allow you to specify child projects as part of a monorepo setup.
	// example: yaml
	// projects:
	//   - label: UI
	//     key: ui
	//     changelog: ui/CHANGELOG.md
	//   - label: Email Sender
	//     key: email_sender
	//     changelog: services/email/CHANGELOG.md
	Projects []ProjectConfigOptions `yaml:"projects,omitempty"`
	// ProjectsVersionSeparator is used to determine the final version when using projects.
	// The result is: project key + projectVersionSeparator + latest/next version.
	// example: yaml
	// projectsVersionSeparator: "_"
	ProjectsVersionSeparator string `yaml:"projectsVersionSeparator,omitempty"`

	cachedEnvVars map[string]string
}

func (c *Config) KindHeader(label string) string {
	for _, kindConfig := range c.Kinds {
		if kindConfig.Format != "" && kindConfig.Label == label {
			return kindConfig.Format
		}
	}

	return c.KindFormat
}

func (c *Config) ChangeFormatForKind(label string) string {
	for _, kindConfig := range c.Kinds {
		if kindConfig.ChangeFormat != "" && kindConfig.Label == label {
			return kindConfig.ChangeFormat
		}
	}

	return c.ChangeFormat
}

func (c *Config) EnvVars() map[string]string {
	if c.cachedEnvVars == nil {
		c.cachedEnvVars = LoadEnvVars(c, os.Environ())
	}

	return c.cachedEnvVars
}

// Save will save the config as a yaml file to the default path
func (c *Config) Save(wf shared.WriteFiler) error {
	bs, _ := yaml.Marshal(&c)
	return wf(ConfigPaths[0], bs, CreateFileMode)
}

func (c *Config) ProjectFromLabel(labelOrKey string) (*ProjectConfigOptions, error) {
	if len(c.Projects) == 0 {
		return &ProjectConfigOptions{}, nil
	}

	if len(labelOrKey) == 0 {
		return nil, errProjectRequired
	}

	for _, pc := range c.Projects {
		if labelOrKey == pc.Label || labelOrKey == pc.Key {
			return &pc, nil
		}
	}

	return nil, errProjectNotFound
}

func (c *Config) ProjectLabels() []string {
	projectLabels := make([]string, len(c.Projects))

	for i, pc := range c.Projects {
		if len(pc.Label) > 0 {
			projectLabels[i] = pc.Label
		} else {
			projectLabels[i] = pc.Key
		}
	}

	return projectLabels
}

// Exists returns whether or not a config already exists
func (c *Config) Exists() (bool, error) {
	for _, p := range ConfigPaths {
		if exists, err := FileExists(p); exists || err != nil {
			return exists, err
		}
	}

	return false, nil
}

// LoadConfig will load the config from the default path
func LoadConfig(rf shared.ReadFiler) (*Config, error) {
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
		return nil, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	// load backward incompatible configs
	if c.FragmentFileFormat == "" {
		if len(c.Projects) > 0 {
			c.FragmentFileFormat += "{{.Project}}-"
		}

		if len(c.Components) > 0 {
			c.FragmentFileFormat += "{{.Component}}-"
		}

		if len(c.Kinds) > 0 {
			c.FragmentFileFormat += "{{.Kind}}-"
		}

		c.FragmentFileFormat += fmt.Sprintf("{{.Time.Format \"%v\"}}", timeFormat)
	}

	return &c, nil
}
