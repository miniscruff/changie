package core

// this file has all the sub and alternate config types
// cause otherwise config.go would get very big

// Configuration options for newlines before and after different elements.
type NewlinesConfig struct {
	After  int `yaml:"after,omitempty" default:"0"`
	Before int `yaml:"before,omitempty" default:"0"`
}

type PromptFormat struct {
	// key: key of our prompt output used in file/fragment templates
	Key string
	// label: display label when rendering prompts ( defaults to key if empty )
	Label string
	// format: line format of our component ( defaults to line format of all components )
	Format string
}

func (p *PromptFormat) Equals(keyOrLabel string) bool {
	return p.Key == keyOrLabel || p.Label == keyOrLabel
}

func (p *PromptFormat) KeyOrLabel() string {
	if p.Key != "" {
		return p.Key
	}

	return p.Label
}

// ( custom type for what data to pass to template cache, we need to document )
type FileWriterFormat struct {
	FilePath string //: file path + template cache to generate text
	Format   string //: string + template cache to generate text

	Newlines NewlinesConfig
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

type FragmentConfig struct {
	// Directory for unrelease fragments relative to root dir
	Dir string `yaml:"dir" default:"unreleased"`
	//template format for generating the file name of fragments ( from FragmentFileFormat )
	FileFormat string

	// todo: docs
	Prompts []Custom `yaml:"prompts,omitempty"`
}

type ProjectOptions struct {
	Key          string          // : key of our component used in file/fragment outputs
	Label        string          //: output label when selecting in prompts ( defaults to key if empty )
	Replacements []Replacement   `yaml:"replacements,omitempty"`
	Changelog    ChangelogConfig `yaml:"changelog"`
}

type ProjectConfig struct {
	VersionSeparator string
	Options          []ProjectOptions
}

type ComponentOptions struct {
	// - line_prompt_type ( inline these values )
	Prompt PromptFormat
	// post: post process options ( just like kind config, ran before kinds if both configured )
	// todo docs
	// is this meant to be here or up a layer?
	Post []PostProcessConfig `yaml:"post,omitempty"`
}

type ComponentConfig struct {
	// line format: output format of our component line, can be empty
	Prompt   PromptFormat
	Newlines NewlinesConfig
	Options  []ComponentOptions
}

func (c ComponentConfig) Names() []string {
	names := make([]string, len(c.Options))
	for i, o := range c.Options {
		names[i] = o.Prompt.KeyOrLabel()
	}

	return names
}

type KindOptions struct {
	// - line_prompt_type ( inline these values )
	Prompt   PromptFormat
	Newlines NewlinesConfig

	// todo?
	// change format: per kind option of the change format ( takes precedence over default change line format )
	Change ChangeConfig

	Post      []PostProcessConfig `yaml:"post,omitempty"`
	AutoLevel string              `yaml:"auto,omitempty"`

	// rename skip to "UseX=true" by default instead of skip=false
	// skip: ( extra level needed? )
	// body: unchanged
	// default prompts: unchanged
	// default post: unchanged
	// ??? prompts: ( from additional choices )
	// ??? should this be wrapped in a prompts and have skip global under it?
}

type KindConfig struct {
	// line format: output format of our kind line, required
	Prompt   PromptFormat
	Newlines NewlinesConfig
	Options  []KindOptions
}

type ChangeConfig struct {
	Prompt   PromptFormat
	Newlines NewlinesConfig

	MinLength *int64
	MaxLength *int64
	UseBlock  bool
}

func (c ChangeConfig) CreateCustom() *Custom {
	customType := CustomString
	if c.UseBlock {
		customType = CustomBlock
	}

	return &Custom{
		Label:     "Body",
		Type:      customType,
		MinLength: c.MinLength,
		MaxLength: c.MaxLength,
	}
}

func (c ChangeConfig) Validate(input string) error {
	cu := Custom{
		Label:     "Body",
		Type:      CustomString,
		MinLength: c.MinLength,
		MaxLength: c.MaxLength,
	}

	return cu.Validate(input)
}

type ReleaseNotesConfig struct {
	// extension: File extension for release notes ( from VersionExt )
	Extension string
	Version   FileWriterFormat
	Header    FileWriterFormat
	Footer    FileWriterFormat
	Newlines  NewlinesConfig
}

// changelog: ( shared type between root and project.options[*].changelog )
type ChangelogConfig struct {
	// output: output path to changelog file ( from ChangelogPath )
	Output       string
	Header       FileWriterFormat
	Footer       FileWriterFormat
	Newlines     NewlinesConfig
	Post         []PostProcessConfig `yaml:"post,omitempty"`
	Replacements []Replacement       `yaml:"replacements,omitempty"`
}
