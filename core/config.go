package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
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

var ErrConfigNotFound = errors.New("no changie config found")

// GetVersions will return, in semver sorted order, all released versions
type GetVersions func(Config) ([]*semver.Version, error)

// Kind config allows you to customize the options depending on what kind was selected.
type KindConfig struct {
	// Key is the value used for lookups and file names for kinds.
	// By default it will use label if no key is provided.
	// example: yaml
	// key: feature
	Key string `yaml:"key,omitempty"`
	// Label is the value used in the prompt when selecting a kind.
	// example: yaml
	// label: Feature
	Label string `yaml:"label,omitempty" required:"true"`
	// Format will override the root kind format when building the kind header.
	// example: yaml
	// format: '### {{.Kind}} **Breaking Changes**'
	Format string `yaml:"format,omitempty"`
	// Change format will override the root change format when building changes specific to this kind.
	// example: yaml
	// changeFormat: 'Breaking: {{.Custom.Body}}
	ChangeFormat string `yaml:"changeFormat,omitempty"`
	// Additional choices allows adding choices per kind
	AdditionalChoices []Custom `yaml:"additionalChoices,omitempty"`
	// Post process options when saving a new change fragment specific to this kind.
	Post []PostProcessConfig `yaml:"post,omitempty"`
	// Skip global choices allows skipping the parent choices options.
	SkipGlobalChoices bool `yaml:"skipGlobalChoices,omitempty" default:"false"`
	// Skip body allows skipping the parent body prompt.
	SkipBody bool `yaml:"skipBody,omitempty" default:"false"`
	// Skip global post allows skipping the parent post processing.
	SkipGlobalPost bool `yaml:"skipGlobalPost,omitempty" default:"false"`
	// Auto determines what value to bump when using `batch auto` or `next auto`.
	// Possible values are major, minor, patch or none and the highest one is used if
	// multiple changes are found. none will not bump the version.
	// Only none changes is not a valid bump and will fail to batch.
	// example: yaml
	// auto: minor
	AutoLevel string `yaml:"auto,omitempty"`
}

// KeyOrLabel returns the kind config key if set, otherwise the label
func (kc *KindConfig) KeyOrLabel() string {
	if kc.Key != "" {
		return kc.Key
	}

	return kc.Label
}

// Body config allows you to customize the default body prompt
type BodyConfig struct {
	// Min length specifies the minimum body length
	MinLength *int64 `yaml:"minLength,omitempty" default:"no min"`
	// Max length specifies the maximum body length
	MaxLength *int64 `yaml:"maxLength,omitempty" default:"no max"`
	// Block allows multiline text inputs for body messages
	UseBlock bool `yaml:"block,omitempty" default:"false"`
}

func (b BodyConfig) CreateCustom() *Custom {
	customType := CustomString
	if b.UseBlock {
		customType = CustomBlock
	}

	return &Custom{
		Label:     "Body",
		Type:      customType,
		MinLength: b.MinLength,
		MaxLength: b.MaxLength,
	}
}

func (b BodyConfig) Validate(input string) error {
	c := Custom{
		Label:     "Body",
		Type:      CustomString,
		MinLength: b.MinLength,
		MaxLength: b.MaxLength,
	}

	return c.Validate(input)
}

// Configuration options for newlines before and after different elements.
type NewlinesConfig struct {
	// Add newlines after change fragment
	AfterChange int `yaml:"afterChange,omitempty" default:"0"`
	// Add newlines after the header file in the merged changelog
	AfterChangelogHeader int `yaml:"afterChangelogHeader,omitempty" default:"0"`
	// Add newlines after adding a version to the changelog
	AfterChangelogVersion int `yaml:"afterChangelogVersion,omitempty" default:"0"`
	// Add newlines after component
	AfterComponent int `yaml:"afterComponent,omitempty" default:"0"`
	// Add newlines after footer file
	AfterFooterFile int `yaml:"afterFooterFile,omitempty" default:"0"`
	// Add newlines after footer template
	AfterFooterTemplate int `yaml:"afterFooter,omitempty" default:"0"`
	// Add newlines after header file
	AfterHeaderFile int `yaml:"afterHeaderFile,omitempty" default:"0"`
	// Add newlines after header template
	AfterHeaderTemplate int `yaml:"afterHeaderTemplate,omitempty" default:"0"`
	// Add newlines after kind
	AfterKind int `yaml:"afterKind,omitempty" default:"0"`
	// Add newlines after version
	AfterVersion int `yaml:"afterVersion,omitempty" default:"0"`
	// Add newlines before change fragment
	BeforeChange int `yaml:"beforeChange,omitempty" default:"0"`
	// Add newlines before adding a version to the changelog
	BeforeChangelogVersion int `yaml:"beforeChangelogVersion,omitempty" default:"0"`
	// Add newlines before component
	BeforeComponent int `yaml:"beforeComponent,omitempty" default:"0"`
	// Add newlines before footer file
	BeforeFooterFile int `yaml:"beforeFooterFile,omitempty" default:"0"`
	// Add newlines before footer template
	BeforeFooterTemplate int `yaml:"beforeFooterTemplate,omitempty" default:"0"`
	// Add newlines before header file
	BeforeHeaderFile int `yaml:"beforeHeaderFile,omitempty" default:"0"`
	// Add newlines before header template
	BeforeHeaderTemplate int `yaml:"beforeHeaderTemplate,omitempty" default:"0"`
	// Add newlines before kind
	BeforeKind int `yaml:"beforeKind,omitempty" default:"0"`
	// Add newlines before version
	BeforeVersion int `yaml:"beforeVersion,omitempty" default:"0"`
	// Add newlines at the end of the version file
	EndOfVersion int `yaml:"endOfVersion,omitempty" default:"0"`
}

// PostProcessConfig allows adding additional custom values to a change fragment
// after all the other inputs are complete.
// This will add additional keys to the `custom` section of the fragment.
// If the key already exists as part of a custom choice the value will be overridden.
type PostProcessConfig struct {
	// Key to save the custom value with
	Key string `yaml:"key"`
	// Value of the custom value as a go template
	Value string `yaml:"value" templateType:"Change"`
}

// ProjectConfig extends changie to support multiple changelog files for different projects
// inside one repository.
// example: yaml
// projects:
//   - label: UI
//     key: ui
//     changelog: ui/CHANGELOG.md
//   - label: Email Sender
//     key: email_sender
//     changelog: services/email/CHANGELOG.md
type ProjectConfig struct {
	// Label is the value used in the prompt when selecting a project.
	// example: yaml
	// label: Frontend
	Label string `yaml:"label"`
	// Key is the value used for unreleased and version output paths.
	// example: yaml
	// key: frontend
	Key string `yaml:"key"`
	// ChangelogPath is the path to the changelog for this project.
	// example: yaml
	// changelog: src/frontend/CHANGELOG.md
	ChangelogPath string `yaml:"changelog"`
	// Replacements to run when merging a changelog for our project.
	// example: yaml
	// # nodejs package.json replacement
	// replacements:
	// - path: ui/package.json
	//   find: '  "version": ".*",'
	//   replace: '  "version": "{{.VersionNoPrefix}}",'
	Replacements []Replacement `yaml:"replacements"`
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
	// `{{.ChangesDir}}/{{.UnreleasedDir}}/{{.FragmentFileFormat}}.yaml`
	// example: yaml
	// fragmentFileFormat: "{{.Kind}}-{{.Custom.Issue}}"
	FragmentFileFormat string `yaml:"fragmentFileFormat,omitempty" default:"{{.Project}}-{{.Component}}-{{.Kind}}-{{.Time.Format \"20060102-150405\"}}" templateType:"Change"` //nolint:lll
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
	Kinds []KindConfig `yaml:"kinds,omitempty"`
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
	Newlines NewlinesConfig `yaml:"newlines,omitempty"`
	// Post process options when saving a new change fragment.
	// example: yaml
	// # build a GitHub link from author choice
	// post:
	// - key: AuthorLink
	//   value: "https://github.com/{{.Custom.Author}}
	// changeFormat: "* {{.Body}} by [{{.Custom.Author}}]({{.Custom.AuthorLink}})"
	Post []PostProcessConfig `yaml:"post,omitempty"`
	// Prefix of environment variables to load for templates.
	// The prefix is removed from resulting key map.
	// example: yaml
	// # if we have an environment variable like so:
	// # export CHANGIE_PROJECT=changie
	// # we can use that in our templates if we set the prefix
	// envPrefix: "CHANGIE_"
	// versionFormat: "New release for {{.Env.PROJECT}}"
	EnvPrefix string `yaml:"envPrefix,omitempty"`
	// Projects allow you to specify child projects as part of a monorepo setup.
	// example: yaml
	// projects:
	//   - label: UI
	//     key: ui
	//     changelog: ui/CHANGELOG.md
	//   - label: Email Sender
	//     key: email_sender
	//     changelog: services/email/CHANGELOG.md
	Projects []ProjectConfig `yaml:"projects,omitempty"`
	// ProjectsVersionSeparator is used to determine the final version when using projects.
	// The result is: project key + projectVersionSeparator + latest/next version.
	// example: yaml
	// projectsVersionSeparator: "_"
	ProjectsVersionSeparator string `yaml:"projectsVersionSeparator,omitempty"`

	cachedEnvVars map[string]string
}

func (c *Config) KindFromKeyOrLabel(keyOrLabel string) *KindConfig {
	for _, kc := range c.Kinds {
		if kc.KeyOrLabel() == keyOrLabel {
			return &kc
		}
	}

	return nil
}

func (c *Config) KindHeader(keyOrLabel string) string {
	for _, kc := range c.Kinds {
		if kc.Format != "" && (kc.Key == keyOrLabel || kc.Label == keyOrLabel) {
			return kc.Format
		}
	}

	return c.KindFormat
}

func (c *Config) ChangeFormatForKind(keyOrLabel string) string {
	for _, kc := range c.Kinds {
		if kc.ChangeFormat != "" && (kc.Key == keyOrLabel || kc.Label == keyOrLabel) {
			return kc.ChangeFormat
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
func (c *Config) Save() error {
	bs, _ := yaml.Marshal(&c)
	return os.WriteFile(ConfigPaths[0], bs, CreateFileMode)
}

func (c *Config) Project(labelOrKey string) (*ProjectConfig, error) {
	if len(c.Projects) == 0 {
		return &ProjectConfig{}, nil
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

// findConfigUpwards will recursively look up until we find a changie config file
func findConfigUpwards() ([]byte, error) {
	currDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		for _, path := range ConfigPaths {
			bs, err := os.ReadFile(filepath.Join(currDir, path))
			if err == nil || !errors.Is(err, fs.ErrNotExist) {
				return bs, nil
			}
		}

		err := os.Chdir("..")
		if err != nil {
			return nil, err
		}

		lastDir := currDir

		currDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}

		// Keep going up, until going up is unchanged.
		if lastDir == currDir {
			return nil, ErrConfigNotFound
		}
	}
}

// LoadConfig will load the config from the default path
func LoadConfig() (*Config, error) {
	var (
		c   Config
		bs  []byte
		err error
	)

	customPath := os.Getenv(configEnvVar)
	if customPath != "" {
		bs, err = os.ReadFile(customPath)
	} else {
		bs, err = findConfigUpwards()
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
