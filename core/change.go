package core

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Less will compare two Change values with the config settings.
// * Components, if enabled, are sorted by index in config
// * Kind, if enabled, are sorted by index in config
// * Time sorted newest first
func ChangeLess(cfg *Config, changes []Change) func(i, j int) bool {
	return func(i, j int) bool {
		a := changes[i]
		b := changes[j]

		// Start by sorting by component index
		if len(cfg.Components) > 0 && a.Component != b.Component {
			for _, c := range cfg.Components {
				if a.Component == c {
					return true
				} else if b.Component == c {
					return false
				}
			}
		}

		// Then sort by kind index
		if len(cfg.Kinds) > 0 && a.Kind != b.Kind {
			for _, k := range cfg.Kinds {
				if a.Kind == k.Key || a.Kind == k.Label {
					return true
				} else if b.Kind == k.Key || b.Kind == k.Label {
					return false
				}
			}
		}

		// Finish sort by oldest first
		return b.Time.After(a.Time)
	}
}

// Change represents an atomic change to a project.
type Change struct {
	// Project of our change, if one was provided.
	Project string `yaml:",omitempty" default:""`
	// Component of our change, if one was provided.
	Component string `yaml:",omitempty" default:""`
	// Kind of our change, if one was provided.
	Kind string `yaml:",omitempty" default:""`
	// Body message of our change, if one was provided.
	Body string `yaml:",omitempty" default:""`
	// When our change was made.
	Time time.Time `yaml:"" required:"true"`
	// Custom values corresponding to our options where each key-value pair is the key of the custom option
	// and value the one provided in the change.
	// example: yaml
	// custom:
	// - key: Issue
	//   type: int
	// changeFormat: "{{.Body}} from #{{.Custom.Issue}}"
	Custom map[string]string `yaml:",omitempty" default:"nil"`
	// Env vars configured by the system.
	// This is not written in change fragments but instead loaded by the system and accessible for templates.
	// For example if you want to use an env var in [change format](#config-changeformat) you can,
	// but env vars configured when executing `changie new` will not be saved.
	// See [envPrefix](#config-envprefix) for configuration.
	Env map[string]string `yaml:"-" default:"nil"`
	// Filename the change was saved to.
	Filename string `yaml:"-"`
}

// WriteTo will write a change to the writer as YAML
func (change Change) WriteTo(writer io.Writer) (int64, error) {
	bs, _ := yaml.Marshal(&change)
	n, err := writer.Write(bs)

	return int64(n), err
}

func (change *Change) PostProcess(cfg *Config, kind *KindConfig) error {
	postConfigs := make([]PostProcessConfig, 0)

	if kind == nil || !kind.SkipGlobalPost {
		postConfigs = append(postConfigs, cfg.Post...)
	}

	if kind != nil {
		postConfigs = append(postConfigs, kind.Post...)
	}

	if len(postConfigs) == 0 {
		return nil
	}

	templateCache := NewTemplateCache()

	for _, post := range postConfigs {
		writer := strings.Builder{}

		err := templateCache.Execute(post.Value, &writer, change)
		if err != nil {
			return err
		}

		change.Custom[post.Key] = writer.String()
	}

	return nil
}

// LoadChange will load a change from file path
func LoadChange(path string) (Change, error) {
	var c Change

	bs, err := os.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("reading change file '%s': %w", path, err)
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, fmt.Errorf("unmarshaling change file '%s': %w", path, err)
	}

	c.Filename = path

	return c, nil
}
