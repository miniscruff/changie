package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"text/template"

	"github.com/icholy/replace"
	"github.com/manifoldco/promptui"
	"golang.org/x/text/transform"
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
	Key         string
	Type        CustomType
	Label       string   `yaml:",omitempty"`
	MinInt      *int64   `yaml:"minInt,omitempty"`
	MaxInt      *int64   `yaml:"maxInt,omitempty"`
	EnumOptions []string `yaml:"enumOptions,omitempty"`
}

// CreatePrompt will create a promptui select or prompt from a custom choice
func (c Custom) CreatePrompt(stdinReader io.ReadCloser) (Prompt, error) {
	label := c.Key
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

type ReplaceData struct {
	Version         string
	VersionNoPrefix string
}

type Replacement struct {
	Path    string
	Find    string
	Replace string
}

func (r Replacement) Execute(readFile ReadFiler, writeFile WriteFiler, data ReplaceData) error {
	templ, err := template.New("replacement").Parse(r.Replace)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	//nolint:errcheck
	templ.Execute(&buf, data)

	fileData, err := readFile(r.Path)
	if err != nil {
		return err
	}

	transformer := replace.Regexp(regexp.MustCompile(r.Find), buf.Bytes())

	newData, _, _ := transform.Bytes(transformer, fileData)

	err = writeFile(r.Path, newData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Config handles configuration for a changie project
type Config struct {
	ChangesDir        string `yaml:"changesDir"`
	UnreleasedDir     string `yaml:"unreleasedDir"`
	HeaderPath        string `yaml:"headerPath"`
	VersionHeaderPath string `yaml:"versionHeaderPath"`
	ChangelogPath     string `yaml:"changelogPath"`
	VersionExt        string `yaml:"versionExt"`
	// formats
	VersionFormat   string `yaml:"versionFormat"`
	ComponentFormat string `yaml:"componentFormat,omitempty"`
	KindFormat      string `yaml:"kindFormat,omitempty"`
	ChangeFormat    string `yaml:"changeFormat"`
	// custom
	Components    []string      `yaml:"components,omitempty"`
	Kinds         []string      `yaml:"kinds,omitempty"`
	CustomChoices []Custom      `yaml:"custom,omitempty"`
	Replacements  []Replacement `yaml:"replacements,omitempty"`
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
