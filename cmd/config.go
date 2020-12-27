package cmd

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"strconv"
)

const (
	configPath   string     = ".changie.yaml"
	CustomString CustomType = "string"
	CustomInt    CustomType = "int"
	CustomEnum   CustomType = "enum"
)

var (
	invalidPromptType = errors.New("Invalid prompt type")
	invalidIntInput   = errors.New("Invalid number")
	intTooLow         = errors.New("Input below minimum")
	intTooHigh        = errors.New("Input above maximum")
)

type EnumWrapper struct {
	*promptui.Select
}

func (e *EnumWrapper) Run() (string, error) {
	_, value, err := e.Select.Run()
	return value, err
}

type Prompt interface {
	Run() (string, error)
}

type CustomType string

type Custom struct {
	Type        CustomType
	Label       string   `yaml:",omitempty"`
	MinInt      *int64   `yaml:"minInt,omitempty"`
	MaxInt      *int64   `yaml:"maxInt,omitempty"`
	EnumOptions []string `yaml:"enumOptions,omitempty"`
}

func (c Custom) CreatePrompt(name string, stdinReader io.ReadCloser) (Prompt, error) {
	label := name
	if c.Label != "" {
		label = c.Label
	}

	switch c.Type {
	case CustomString:
		return &promptui.Prompt{
			Label: label,
			Stdin: stdinReader,
		}, nil
	case CustomInt:
		return &promptui.Prompt{
			Label: label,
			Stdin: stdinReader,
			Validate: func(input string) error {
				value, err := strconv.ParseInt(input, 10, 64)
				if err != nil {
					return invalidIntInput
				}
				if c.MinInt != nil && value < *c.MinInt {
					return fmt.Errorf("%w: %v < %v", intTooLow, value, c.MinInt)
				}
				if c.MaxInt != nil && value > *c.MaxInt {
					return fmt.Errorf("%w: %v > %v", intTooHigh, value, c.MinInt)
				}
				return nil
			},
		}, nil
	case CustomEnum:
		return &EnumWrapper{
			Select: &promptui.Select{
				Label: label,
				Stdin: stdinReader,
				Items: c.EnumOptions,
			}}, nil
	}
	return nil, invalidPromptType
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

func (config Config) Save(wf WriteFiler) error {
	bs, _ := yaml.Marshal(&config)
	return wf(configPath, bs, os.ModePerm)
}

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
