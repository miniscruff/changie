package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const timeFormat string = "20060102-150405"

// Change represents an atomic change to a project
type Change struct {
	Component string `yaml:",omitempty"`
	Kind      string `yaml:",omitempty"`
	Body      string
	Time      time.Time
	Custom    map[string]string `yaml:",omitempty"`
}

// SaveUnreleased will save an unreleased change to the unreleased directory
func (change Change) SaveUnreleased(wf WriteFiler, config Config) error {
	bs, _ := yaml.Marshal(&change)
	nameParts := make([]string, 0)

	if change.Component != "" {
		nameParts = append(nameParts, change.Component)
	}

	if change.Kind != "" {
		nameParts = append(nameParts, change.Kind)
	}

	nameParts = append(nameParts, change.Time.Format(timeFormat))

	filePath := fmt.Sprintf(
		"%s/%s/%s.yaml",
		config.ChangesDir,
		config.UnreleasedDir,
		strings.Join(nameParts, "-"),
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

type changesConfigSorter struct {
	changes []Change
	config  Config
}

func SortByConfig(config Config) *changesConfigSorter {
	return &changesConfigSorter{
		config: config,
	}
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (s *changesConfigSorter) Sort(changes []Change) {
	s.changes = changes
	sort.Sort(s)
}

// Len is part of sort.Interface.
func (s *changesConfigSorter) Len() int {
	return len(s.changes)
}

// Swap is part of sort.Interface.
func (s *changesConfigSorter) Swap(i, j int) {
	s.changes[i], s.changes[j] = s.changes[j], s.changes[i]
}

// Less will compare two Change values with the config settings.
// * Components, if enabled, are sorted by index in config
// * Kind, if enabled, are sorted by index in config
// * Time sorted newest first
func (s *changesConfigSorter) Less(i, j int) bool {
	a, b := &s.changes[i], &s.changes[j]

	// Start by sorting by component index
	if len(s.config.Components) > 0 && a.Component != b.Component {
		for _, c := range s.config.Components {
			if a.Component == c {
				return true
			} else if b.Component == c {
				return false
			}
		}
	}

	// Then sort by kind index
	if len(s.config.Kinds) > 0 && a.Kind != b.Kind {
		for _, k := range s.config.Kinds {
			if a.Kind == k {
				return true
			} else if b.Kind == k {
				return false
			}
		}
	}

	// Finish sort by newest first
	return a.Time.After(b.Time)
}
