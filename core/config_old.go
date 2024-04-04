package core

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

// Kind config allows you to customize the options depending on what kind was selected.
type KindConfigOld struct {
	// Label is the value used in the prompt when selecting a kind.
	// example: yaml
	// label: Feature
	Label string `yaml:",omitempty" required:"true"`
	// Format will override the root kind format when building the kind header.
	// example: yaml
	// format: '### {{.Kind}} **Breaking Changes**'
	Format string `yaml:",omitempty"`
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
type ProjectConfigOptionsOld struct {
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

// Configuration options for newlines before and after different elements.
type NewlinesConfigOld struct {
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
