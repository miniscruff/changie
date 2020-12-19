package cmd

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

const timeFormat string = "20060102-150405"

type Change struct {
	Kind   string
	Body   string
	Custom map[string]string `yaml:",omitempty"`
}

func (change Change) SaveUnreleased(wf WriteFiler, tn TimeNow, config Config) error {
	bs, _ := yaml.Marshal(&change)
	timeString := tn().Format(timeFormat)
	filePath := fmt.Sprintf("%s/%s/%s-%s.yaml", config.ChangesDir, config.UnreleasedDir, change.Kind, timeString)

	return wf.WriteFile(filePath, bs, os.ModePerm)
}

func LoadChange(path string, rf ReadFiler) (Change, error) {
	var c Change
	bs, err := rf.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
