package main

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"os"
)

const configPath string = ".changie.yaml"

// Config handles configuration for a changie project
type Config struct {
	ChangesDir    string
	UnreleasedDir string
	HeaderPath    string
	ChangelogPath string
	// formats
	// custom questions
}

func (config Config) Save(fs afero.Fs) error {
	bs, err := yaml.Marshal(config)
	if err != nil {
		return nil
	}

	afs := afero.Afero{Fs: fs}
	return afs.WriteFile(configPath, bs, os.ModePerm)
}

func LoadConfig(fs afero.Fs) Config {
	return Config{}
}
