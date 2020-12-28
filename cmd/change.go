package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const timeFormat string = "20060102-150405"

// Change represents an atomic change to a project
type Change struct {
	Kind   string
	Body   string
	Custom map[string]string `yaml:",omitempty"`
}

// SaveUnreleased will save an unreleased change to the unreleased directory
func (change Change) SaveUnreleased(wf WriteFiler, tn TimeNow, config Config) error {
	bs, _ := yaml.Marshal(&change)
	timeString := tn().Format(timeFormat)
	filePath := fmt.Sprintf(
		"%s/%s/%s-%s.yaml",
		config.ChangesDir,
		config.UnreleasedDir,
		change.Kind,
		timeString,
	)

	return wf(filePath, bs, os.ModePerm)
}

// LoadChange will load a change from file path
func LoadChange(path string, rf ReadFiler) (Change, error) {
	var c Change

	bs, err := rf(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
