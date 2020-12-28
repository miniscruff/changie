package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
)

const (
	configPath   string     = ".changie.yaml"
	customString CustomType = "string"
	customInt    CustomType = "int"
	customEnum   CustomType = "enum"
)

var (
	errInvalidPromptType = errors.New("invalid prompt type")
	errInvalidIntInput   = errors.New("invalid number")
	errIntTooLow         = errors.New("input below minimum")
	errIntTooHigh        = errors.New("input above maximum")
)

type enumWrapper struct {
	*promptui.Select
}

func (e *enumWrapper) Run() (string, error) {
	_, value, err := e.Select.Run()
	return value, err
}

// Prompt is a small wrapper around the promptui Run method
type Prompt interface {
	Run() (string, error)
}

// CustomType determines the possible custom choice types
type CustomType string

// Custom contains the options for a custom choice for new changes
type Custom struct {
	Type        CustomType
	Label       string   `yaml:",omitempty"`
	MinInt      *int64   `yaml:"minInt,omitempty"`
	MaxInt      *int64   `yaml:"maxInt,omitempty"`
	EnumOptions []string `yaml:"enumOptions,omitempty"`
}

// CreatePrompt will create a promptui select or prompt from a custom choice
func (c Custom) CreatePrompt(name string, stdinReader io.ReadCloser) (Prompt, error) {
	label := name
	if c.Label != "" {
		label = c.Label
	}

	switch c.Type {
	case customString:
		return &promptui.Prompt{
			Label: label,
			Stdin: stdinReader,
		}, nil
	case customInt:
		return &promptui.Prompt{
			Label: label,
			Stdin: stdinReader,
			Validate: func(input string) error {
				value, err := strconv.ParseInt(input, 10, 64)
				if err != nil {
					return errInvalidIntInput
				}
				if c.MinInt != nil && value < *c.MinInt {
					return fmt.Errorf("%w: %v < %v", errIntTooLow, value, c.MinInt)
				}
				if c.MaxInt != nil && value > *c.MaxInt {
					return fmt.Errorf("%w: %v > %v", errIntTooHigh, value, c.MinInt)
				}
				return nil
			},
		}, nil
	case customEnum:
		return &enumWrapper{
			Select: &promptui.Select{
				Label: label,
				Stdin: stdinReader,
				Items: c.EnumOptions,
			}}, nil
	}

	return nil, errInvalidPromptType
}

// Config handles configuration for a changie project
type Config struct {
	ChangesDir    string `yaml:"changesDir"`
	UnreleasedDir string `yaml:"unreleasedDir"`
	HeaderPath    string `yaml:"headerPath"`
	ChangelogPath string `yaml:"changelogPath"`
	VersionExt    string `yaml:"versionExt"`
	// formats
	VersionFormat string `yaml:"versionFormat"`
	KindFormat    string `yaml:"kindFormat"`
	ChangeFormat  string `yaml:"changeFormat"`
	// custom
	Kinds         []string          `yaml:"kinds"`
	CustomChoices map[string]Custom `yaml:"custom,omitempty"`
}

// Save will save the config as a yaml file to the default path
func (config Config) Save(wf WriteFiler) error {
	bs, _ := yaml.Marshal(&config)
	return wf(configPath, bs, os.ModePerm)
}

// LoadConfig will load the config from the default path
func LoadConfig(rf ReadFiler) (Config, error) {
	var c Config

	bs, err := rf(configPath)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
