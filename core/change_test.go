package core

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/miniscruff/changie/then"
)

var (
	orderedTimes = []time.Time{
		time.Date(2018, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2017, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local),
	}
)

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
	then.WithTempDir(t)
	then.Nil(t, os.WriteFile("some_file.yaml", []byte("kind: A\nbody: hey\n"), CreateFileMode))

	change, err := LoadChange("some_file.yaml")

	then.Nil(t, err)
	then.Equals(t, "A", change.Kind)
	then.Equals(t, "hey", change.Body)
}

func TestBadYamlFile(t *testing.T) {
	then.WithTempDir(t)
	then.Nil(t, os.WriteFile("some_file.yaml", []byte("not a yaml file---"), CreateFileMode))

	_, err := LoadChange("some_file.yaml")
	then.NotNil(t, err)
}

func TestSortByTime(t *testing.T) {
	changes := []Change{
		{Body: "third", Time: orderedTimes[2]},
		{Body: "second", Time: orderedTimes[1]},
		{Body: "first", Time: orderedTimes[0]},
	}
	cfg := &Config{}

	sort.Slice(changes, ChangeLess(cfg, changes))

	then.Equals(t, "third", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
	then.Equals(t, "first", changes[2].Body)
}

func TestSortByKindThenTime(t *testing.T) {
	cfg := &Config{
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
	sort.Slice(changes, ChangeLess(cfg, changes))

	then.Equals(t, "first", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
	then.Equals(t, "third", changes[2].Body)
	then.Equals(t, "fourth", changes[3].Body)
}

func TestSortByComponentThenKind(t *testing.T) {
	cfg := &Config{
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
	sort.Slice(changes, ChangeLess(cfg, changes))

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

func TestPostProcess_NoPostConfigs(t *testing.T) {
	cfg := &Config{}
	change := &Change{}

	then.Nil(t, change.PostProcess(cfg, nil))
	then.MapLen(t, 0, change.Custom)
}

func TestPostProcess_NilKindGlobalPost(t *testing.T) {
	cfg := &Config{
		Post: []PostProcessConfig{
			{
				Key:   "NilKind",
				Value: "Yes",
			},
		},
	}
	change := &Change{
		Custom: make(map[string]string),
	}

	then.Nil(t, change.PostProcess(cfg, nil))
	then.MapLen(t, 1, change.Custom)
	then.Equals(t, "Yes", change.Custom["NilKind"])
}

func TestPostProcess_KindPost(t *testing.T) {
	cfg := &Config{}
	change := &Change{
		Custom: make(map[string]string),
	}
	kindCfg := &KindConfig{
		Post: []PostProcessConfig{
			{
				Key:   "NotNilKind",
				Value: "AlsoYes",
			},
		},
	}

	then.Nil(t, change.PostProcess(cfg, kindCfg))
	then.MapLen(t, 1, change.Custom)
	then.Equals(t, "AlsoYes", change.Custom["NotNilKind"])
}

func TestPostProcess_InvalidPostTemplate(t *testing.T) {
	cfg := &Config{}
	change := &Change{
		Custom: make(map[string]string),
	}
	kindCfg := &KindConfig{
		Post: []PostProcessConfig{
			{
				Key:   "BrokenTemplate",
				Value: "{{ ..-... }}",
			},
		},
	}

	then.NotNil(t, change.PostProcess(cfg, kindCfg))
}
