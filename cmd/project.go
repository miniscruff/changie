package cmd

import (
	"github.com/spf13/afero"
)

// ProjectConfig handles configuration for a changie project
type ProjectConfig struct {
}

// Project handles all the work on a changie project
type Project struct {
	config ProjectConfig
	fs     afero.Fs
}

func NewProject(fs afero.Fs, config ProjectConfig) *Project {
	return &Project{
		fs:     fs,
		config: config,
	}
}

func (p *Project) Init() error {
	var err error

	err = p.fs.MkdirAll("changes/unreleased", 644)
	if err != nil {
		return err
	}

	keepFile, err := p.fs.Create("changes/unreleased/.gitkeep")
	if err != nil {
		return err
	}
	defer keepFile.Close()

	headerFile, err := p.fs.Create("changes/header.tpl.md")
	if err != nil {
		return err
	}
	defer headerFile.Close()

	_, err = headerFile.WriteString(defaultHeader)
	if err != nil {
		return err
	}

	outputFile, err := p.fs.Create("CHANGELOG.md")
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(defaultChangelog)
	if err != nil {
		return err
	}

	return nil
}
