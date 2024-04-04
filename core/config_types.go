package core

import (
	"github.com/Masterminds/semver/v3"

	"github.com/miniscruff/changie/shared"
)

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

// NewlineConfig configures how we add newlines before and after templates.
type NewlinesConfig struct {
	// Before is the amount of newlines to add before writing text.
	Before int `yaml:"before"`
	// After is the amount of newlines to add after writing text.
	After int `yaml:"after"`
}

// LinePromptConfig TODO:
type LinePromptConfig struct {
	// Key of our prompt output used in file/fragment templates
	Key string `yaml:"key"`
	// Label is the display text when asking users for prompt values.
	// If not defined, the key is used.
	Label string `yaml:"label"`
	// Line format ...
	// ??? can projects have line formats? What would that do?
	LineFormat string `yaml:"lineFormat"`
}

// FileWriterConfig TODO:
// TODO: ( custom type for what data to pass to template cache, we need to document )
type FileWriterConfig struct {
	Filepath   string         `yaml:"filepath"`
	LineFormat string         `yaml:"lineFormat"`
	Newlines   NewlinesConfig `yaml:"newlines"`
}

type FragmentConfig struct {
	// dir: directory for unreleased fragments relative to root dir ( from UnreleasedDir )
	Dir string `yaml:"dir"`
	// file format: template format for generating the file name of fragments ( from FragmentFileFormat )
	FileFormat string `yaml:"fileFormat"`
	// posts: post processing ( from root posts )
	Post []PostProcessConfig `yaml:"post,omitempty"`
	// prompts: ( from Customs )
	// TODO: rename/recreate type from Custom to Prompt
	Prompts []Custom `yaml:"prompts,omitempty"`
}

type ProjectConfig struct {
	// version separator: ( from project version separator )
	VersionSeperator string `yaml:"versionSeperator"`

	Options ProjectConfigOptions `yaml:"options"`
}

type ProjectConfigOptions struct {
	Label     string          `yaml:"label"`
	Key       string          `yaml:"key"`
	Changelog ChangelogConfig `yaml:"changelog"`
}

type ComponentConfig struct {
	// line format: output format of our component line, can be empty
	LineFormat string `yaml:"lineFormat"`
	// newlines: see newline_type
	Newlines NewlinesConfig `yaml:"newlines"`

	Options []ComponentOptions `yaml:"options"`
}

type ComponentOptions struct {
	// line_prompt_type ( inline these values )
	LinePrompt LinePromptConfig `yaml:",inline"`
	// post: post process options ( just like kind config, ran before kinds if both configured )
	Post PostProcessConfig `yaml:"post"`
}

type KindConfig struct {
	// line format: output format of our kind line, can be empty
	LineFormat string `yaml:"lineFormat"`
	// newlines: see newline_type
	Newlines NewlinesConfig `yaml:"newlines"`
	Options  []KindOptions  `yaml:"options"`
}

type KindOptions struct {
	// line_prompt_type ( inline these values )
	LinePrompt LinePromptConfig `yaml:",inline"`
	// post: post process options ( just like kind config, ran before kinds if both configured )
	Post []PostProcessConfig `yaml:"post"`

	// change format: per kind option of the change format ( takes precedence over default change line format )
	ChangeFormat string `yaml:"changeFormat"`
	// TODO: rename/recreate type from Custom to Prompt
	Prompts []Custom `yaml:"prompts,omitempty"`

	AutoLevel string   `yaml:"auto,omitempty"`
	Skip      KindSkip `yaml:"skip"`
}

type KindSkip struct {
	Body           bool `yaml:"body"`
	DefaultPrompts bool `yaml:"defaultPrompts"`
	DefaultPost    bool `yaml:"defaultPost"`
	// ??? prompts: ( from additional choices )
	// ??? should this be wrapped in a prompts and have skip global under it?
}

type ChangeConfig struct {
	// line_prompt_type ( inline these values ) ( default to body )
	LinePrompt LinePromptConfig `yaml:",inline"`
	// min_length: min length validation
	MinLength *int `yaml:"minLength,omitempty"`
	// max_length: max length validation
	MaxLength *int `yaml:"maxLength,omitempty"`
	// TODO: skip or enabled?
	Skip     bool           `yaml:"skip"`
	Newlines NewlinesConfig `yaml:"newlines"`
}

type ReleaseNotesConfig struct {
	// extension: File extension for release notes ( from VersionExt )
	Extension string `yaml:"extension"`
	// version: see file_writer_type
	Version FileWriterConfig `yaml:"version"`
	// header: see file_writer_type
	Header FileWriterConfig `yaml:"header"`
	// footer: see file_writer_type
	Footer FileWriterConfig `yaml:"footer"`
	// Newlines...
	Newlines NewlinesConfig `yaml:"newlines"`
}

// changelog: ( shared type between root and project.options[*].changelog )
type ChangelogConfig struct {
	Output string `yaml:"output"`
	// header: see file_writer_type
	Header FileWriterConfig `yaml:"header"`
	// footer: see file_writer_type
	Footer FileWriterConfig `yaml:"footer"`
	// replacements: moved to better support projects
	Replacements []Replacement `yaml:"replacements,omitempty"`
	// Newlines...
	Newlines NewlinesConfig `yaml:"newlines"`
}

// GetVersions will return, in semver sorted order, all released versions
// TODO: Might not be used.
type GetVersions func(shared.ReadDirer, Config) ([]*semver.Version, error)
