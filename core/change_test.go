package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/miniscruff/changie/then"
)

var (
	orderedTimes = []time.Time{
		time.Date(2018, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2017, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local),
	}
	changeIncrementer = 0
)

// writeChangeFile will write a change file with an auto-incrementing index to prevent
// same second clobbering
func writeChangeFile(t *testing.T, cfg *Config, change Change) {
	// set our time as an arbitrary amount from jan 1 2000 so
	// each change is 1 hour later then the last
	if change.Time.Year() == 0 {
		diff := time.Duration(changeIncrementer) * time.Hour
		change.Time = time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC).Add(diff)
	}

	bs, _ := yaml.Marshal(&change)
	name := fmt.Sprintf("change-%d.yaml", changeIncrementer)
	changeIncrementer++

	then.WriteFile(t, bs, cfg.ChangesDir, cfg.UnreleasedDir, name)
}

func TestWriteChange(t *testing.T) {
	mockTime := time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local)
	change := Change{
		Body: "some body message",
		Time: mockTime,
	}

	changesYaml := fmt.Sprintf(
		"body: some body message\ntime: %s\n",
		mockTime.Format(time.RFC3339Nano),
	)

	var builder strings.Builder
	_, err := change.WriteTo(&builder)

	then.Equals(t, changesYaml, builder.String())
	then.Nil(t, err)
}

func TestLoadChangeFromPath(t *testing.T) {
	mockRf := func(filepath string) ([]byte, error) {
		then.Equals(t, "some_file.yaml", filepath)
		return []byte("kind: A\nbody: hey\n"), nil
	}

	change, err := LoadChange("some_file.yaml", mockRf)

	then.Nil(t, err)
	then.Equals(t, "A", change.Kind)
	then.Equals(t, "hey", change.Body)
}

func TestBadRead(t *testing.T) {
	mockErr := errors.New("bad file")

	mockRf := func(filepath string) ([]byte, error) {
		then.Equals(t, "some_file.yaml", filepath)
		return []byte(""), mockErr
	}

	_, err := LoadChange("some_file.yaml", mockRf)
	then.Err(t, mockErr, err)
}

func TestBadYamlFile(t *testing.T) {
	mockRf := func(filepath string) ([]byte, error) {
		then.Equals(t, "some_file.yaml", filepath)
		return []byte("not a yaml file---"), nil
	}

	_, err := LoadChange("some_file.yaml", mockRf)
	then.NotNil(t, err)
}

func TestSortByTime(t *testing.T) {
	changes := []Change{
		{Body: "third", Time: orderedTimes[2]},
		{Body: "second", Time: orderedTimes[1]},
		{Body: "first", Time: orderedTimes[0]},
	}
	config := &Config{}
	SortByConfig(config).Sort(changes)

	then.Equals(t, "third", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
	then.Equals(t, "first", changes[2].Body)
}

func TestSortByKindThenTime(t *testing.T) {
	config := &Config{
		Kinds: []KindConfig{
			{Label: "A"},
			{Label: "B"},
		},
	}
	changes := []Change{
		{Kind: "B", Body: "fourth", Time: orderedTimes[0]},
		{Kind: "A", Body: "second", Time: orderedTimes[1]},
		{Kind: "B", Body: "third", Time: orderedTimes[1]},
		{Kind: "A", Body: "first", Time: orderedTimes[2]},
	}
	SortByConfig(config).Sort(changes)

	then.Equals(t, "first", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
	then.Equals(t, "third", changes[2].Body)
	then.Equals(t, "fourth", changes[3].Body)
}

func TestSortByComponentThenKind(t *testing.T) {
	config := &Config{
		Kinds: []KindConfig{
			{Label: "D"},
			{Label: "E"},
		},
		Components: []string{"A", "B", "C"},
	}
	changes := []Change{
		{Body: "fourth", Component: "B", Kind: "E"},
		{Body: "second", Component: "A", Kind: "E"},
		{Body: "third", Component: "B", Kind: "D"},
		{Body: "first", Component: "A", Kind: "D"},
	}
	SortByConfig(config).Sort(changes)

	then.Equals(t, "first", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
	then.Equals(t, "third", changes[2].Body)
	then.Equals(t, "fourth", changes[3].Body)
}

func TestLoadEnvVars(t *testing.T) {
	config := Config{
		EnvPrefix: "TEST_CHANGIE_",
	}

	t.Setenv("TEST_CHANGIE_ALPHA", "Beta")
	t.Setenv("IGNORE", "Delta")

	expectedEnvVars := map[string]string{
		"ALPHA": "Beta",
	}

	then.MapEquals(t, expectedEnvVars, config.EnvVars())

	// will use the cached results
	t.Setenv("TEST_CHANGIE_UNUSED", "NotRead")
	then.MapEquals(t, expectedEnvVars, config.EnvVars())
}
