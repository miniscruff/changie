package cmd

import (
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
	Label       string   `yaml:",omitempty"`
	MinInt      *int64   `yaml:"minInt,omitempty"`
	MaxInt      *int64   `yaml:"maxInt,omitempty"`
	EnumOptions []string `yaml:"enumOptions,omitempty"`
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
