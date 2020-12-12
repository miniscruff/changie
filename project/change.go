package project

import (
	"fmt"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type Change struct {
	Kind   string
	Body   string
	Time   time.Time
	Custom map[string]string
}

func (change Change) SaveUnreleased(fs afero.Fs, config Config) error {
	bs, err := yaml.Marshal(&change)
	if err != nil {
		return nil
	}

	afs := afero.Afero{Fs: fs}
	return afs.WriteFile(
		fmt.Sprintf("%s/%s/%s-suffix.yaml", config.ChangesDir, config.UnreleasedDir, change.Kind),
		bs, os.ModePerm,
	)
}
