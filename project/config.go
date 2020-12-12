package project

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"os"
)

const (
	configPath   string     = ".changie.yaml"
	CustomString CustomType = "string"
	CustomInt    CustomType = "int"
	CustomEnum   CustomType = "enum"
)

type CustomType string

type Custom struct {
	Type        CustomType
	MinInt      *int64   `yaml:"minInt,omitempty"`
	MaxInt      *int64   `yaml:"maxInt,omitempty"`
	EnumOptions []string `yaml:"enumOptions,omitempty"`
}

// Config handles configuration for a changie project
type Config struct {
	ChangesDir    string   `yaml:"changesDir"`
	UnreleasedDir string   `yaml:"unreleasedDir"`
	HeaderPath    string   `yaml:"headerPath"`
	ChangelogPath string   `yaml:"changelogPath"`
	Kinds         []string `yaml:"kinds"`
	// formats
	CustomChoices map[string]Custom `yaml:"custom,omitempty"`
	// custom questions
}

func (config Config) Save(fs afero.Fs) error {
	bs, err := yaml.Marshal(&config)
	if err != nil {
		return nil
	}

	afs := afero.Afero{Fs: fs}
	return afs.WriteFile(configPath, bs, os.ModePerm)
}

func LoadConfig(fs afero.Fs) (Config, error) {
	var c Config
	afs := afero.Afero{Fs: fs}
	bs, err := afs.ReadFile(configPath)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
