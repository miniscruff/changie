package project

import (
	"fmt"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

const timeFormat string = "20060102-150405"

type Change struct {
	Kind   string
	Body   string
	Custom map[string]string `yaml:",omitempty"`
}

func (change Change) SaveUnreleased(fs afero.Fs, config Config) error {
	bs, err := yaml.Marshal(&change)
	if err != nil {
		return nil
	}

	afs := afero.Afero{Fs: fs}
	timeString := time.Now().Format(timeFormat)
	filePath := fmt.Sprintf("%s/%s/%s-%s.yaml", config.ChangesDir, config.UnreleasedDir, change.Kind, timeString)

	// if this file conflicts with another, try again. It should happen very rarely
	if ok, _ := afs.Exists(filePath); ok {
		return change.SaveUnreleased(fs, config)
	}
	return afs.WriteFile(filePath, bs, os.ModePerm)
}

func LoadChange(path string, afs afero.Afero) (Change, error) {
	var c Change
	bs, err := afs.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
