package core

import (
	"errors"
	"fmt"
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
	err := change.Write(&builder)

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
	config := Config{}
	SortByConfig(config).Sort(changes)

	then.Equals(t, "third", changes[0].Body)
	then.Equals(t, "second", changes[1].Body)
	then.Equals(t, "first", changes[2].Body)
}

func TestSortByKindThenTime(t *testing.T) {
	config := Config{
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
	config := Config{
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

func TestAskPromptsForBody(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("body stuff"),
		[]byte{13},
	)

	config := Config{}
	c := &Change{}
	then.Nil(t, c.AskPrompts(config, reader))

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.Equals(t, 0, len(c.Custom))
	then.Equals(t, "body stuff", c.Body)
}

func TestAskComponentKindBody(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{106, 13},
		[]byte{106, 106, 13},
		[]byte("body here"),
		[]byte{13},
	)

	config := Config{
		Components: []string{"cli", "tests", "utils"},
		Kinds: []KindConfig{
			{Label: "added"},
			{Label: "changed"},
			{Label: "removed"},
		},
	}
	c := &Change{}
	then.Nil(t, c.AskPrompts(config, reader))

	then.Equals(t, "tests", c.Component)
	then.Equals(t, "removed", c.Kind)
	then.Equals(t, len(c.Custom), 0)
	then.Equals(t, "body here", c.Body)
}

func TestBodyKindPostProcess(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)

	config := Config{
		Kinds: []KindConfig{
			{
				Label: "added",
				Post: []PostProcessConfig{
					{Key: "Post", Value: "{{.Body}}+{{.Custom.Issue}}"},
				},
			},
		},
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
		},
	}
	c := &Change{
		Body: "our body",
		Kind: "added",
		Custom: map[string]string{
			"Issue": "25",
		},
	}

	then.Nil(t, c.AskPrompts(config, reader))

	then.Equals(t, "", c.Component)
	then.Equals(t, "added", c.Kind)
	then.Equals(t, "our body", c.Body)
	then.Equals(t, "25", c.Custom["Issue"])
	then.Equals(t, "our body+25", c.Custom["Post"])
}

func TestBodyCustom(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("body again"),
		[]byte{13},
		[]byte("custom check value"),
		[]byte{13},
	)

	config := Config{
		CustomChoices: []Custom{
			{Key: "check", Type: CustomString},
		},
	}
	c := &Change{}
	then.Nil(t, c.AskPrompts(config, reader))

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.Equals(t, "custom check value", c.Custom["check"])
	then.Equals(t, "body again", c.Body)
}

func TestBodyCustomWithExistingCustomValue(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("body again"),
		[]byte{13},
		[]byte("256"),
		[]byte{13},
	)

	config := Config{
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomString},
			{Key: "Project", Label: "Custom Project Prompt", Type: CustomString},
		},
	}
	c := &Change{
		Custom: map[string]string{
			"Project": "Changie",
		},
	}

	then.Nil(t, c.AskPrompts(config, reader))

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.Equals(t, "256", c.Custom["Issue"])
	then.Equals(t, "Changie", c.Custom["Project"])
	then.Equals(t, "body again", c.Body)
}

func TestSkippedBodyGlobalChoicesKindWithAdditional(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{13},
		[]byte("breaking value"),
		[]byte{13},
	)

	config := Config{
		CustomChoices: []Custom{
			{Key: "skipped", Type: CustomString, Label: "a"},
		},
		Kinds: []KindConfig{
			{
				Label:             "added",
				SkipBody:          true,
				SkipGlobalChoices: true,
				AdditionalChoices: []Custom{
					{Key: "break", Type: CustomString, Label: "b"},
				},
			},
			{Label: "not used"},
		},
	}

	c := &Change{}
	then.Nil(t, c.AskPrompts(config, reader))

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Custom["skipped"])
	then.Equals(t, "breaking value", c.Custom["break"])
	then.Equals(t, "", c.Body)
}

func TestBodyAndPostProcess(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)

	config := Config{
		Post: []PostProcessConfig{
			{Key: "Post", Value: "Body again {{.Body}}"},
		},
	}

	c := &Change{
		Body: "our body",
	}

	then.Nil(t, c.AskPrompts(config, reader))

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.Equals(t, "Body again our body", c.Custom["Post"])
	then.Equals(t, "our body", c.Body)
}

func TestBodyAndPostProcessSkipGlobalPost(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := Config{
		Kinds: []KindConfig{
			{
				Label:          "added",
				SkipGlobalPost: true,
				Post: []PostProcessConfig{
					{Key: "Post", Value: "{{.Body}}+{{.Custom.Issue}}"},
				},
			},
		},
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
		},
		Post: []PostProcessConfig{
			{Key: "GlobalPost", Value: "should be skipped"},
		},
	}

	c := &Change{
		Body: "our body",
		Kind: "added",
		Custom: map[string]string{
			"Issue": "30",
		},
	}

	then.Nil(t, c.AskPrompts(config, reader))

	then.Equals(t, "", c.Component)
	then.Equals(t, "added", c.Kind)
	then.Equals(t, "our body", c.Body)
	then.Equals(t, "our body+30", c.Custom["Post"])
	then.Equals(t, "", c.Custom["GlobalPost"])
}

func TestErrorInvalidBody(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("abc"),
		[]byte{13},
	)

	var min int64 = 5

	config := Config{
		Body: BodyConfig{
			MinLength: &min,
		},
	}
	c := &Change{}
	then.NotNil(t, c.AskPrompts(config, reader))
}

func TestErrorInvalidPost(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := Config{
		Post: []PostProcessConfig{
			{Key: "Post", Value: "invalid {{++...thing}}"},
		},
	}
	c := &Change{
		Body: "our body",
	}
	then.NotNil(t, c.AskPrompts(config, reader))
}

func TestErrorInvalidCustomType(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("invalid custom type"),
		[]byte{13},
	)

	config := Config{
		CustomChoices: []Custom{
			{Key: "check", Type: "bad type", Label: "a"},
		},
	}

	c := &Change{}
	then.NotNil(t, c.AskPrompts(config, reader))
}

func TestErrorInvalidComponent(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{3},
	)

	config := Config{
		Components: []string{"a", "b"},
	}
	c := &Change{}
	then.NotNil(t, c.AskPrompts(config, reader))
}

func TestErrorInvalidKind(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{3},
	)

	config := Config{
		Kinds: []KindConfig{
			{Label: "a"},
			{Label: "b"},
		},
	}
	c := &Change{}
	then.NotNil(t, c.AskPrompts(config, reader))
}

func TestErrorFaultBody(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{3},
	)

	config := Config{}
	c := &Change{}
	then.NotNil(t, c.AskPrompts(config, reader))
}

func TestErrorInvalidCustomValue(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("custom body"),
		[]byte{13},
		[]byte("control C out"),
		[]byte{3},
	)

	config := Config{
		CustomChoices: []Custom{
			{Key: "check", Type: CustomString, Label: "a"},
		},
	}
	c := &Change{}
	then.NotNil(t, c.AskPrompts(config, reader))
}

func TestErrorBadKindLabel(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
	}
	c := &Change{
		Kind: "not kind",
		Body: "body",
	}

	err := c.AskPrompts(config, reader)
	then.Err(t, errInvalidKind, err)
}

func TestErrorBadComponentInput(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := Config{
		Components: []string{"a", "b", "c"},
	}
	c := &Change{
		Component: "d",
		Body:      "body",
	}

	err := c.AskPrompts(config, reader)
	then.Err(t, errInvalidComponent, err)
}

func TestErrorComponentGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := Config{}
	c := &Change{
		Component: "we shouldn't have a component",
		Body:      "body",
	}

	err := c.AskPrompts(config, reader)
	then.Err(t, errComponentProvidedWhenNotConfigured, err)
}

func TestErrorKindGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := Config{}
	c := &Change{
		Kind: "we shouldn't have a kind",
		Body: "body",
	}

	err := c.AskPrompts(config, reader)
	then.Err(t, errKindProvidedWhenNotConfigured, err)
}

func TestErrorCustomGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := Config{
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
		},
	}
	c := &Change{
		Custom: map[string]string{
			"MissingKey": "40",
		},
	}

	err := c.AskPrompts(config, reader)
	then.Err(t, errCustomProvidedNotConfigured, err)
}

func TestErrorCustomGivenDoesNotPassValidation(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	minValue := int64(50)
	config := Config{
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt, MinInt: &minValue},
		},
	}
	c := &Change{
		Custom: map[string]string{
			"Issue": "40",
		},
	}

	err := c.AskPrompts(config, reader)
	then.Err(t, errIntTooLow, err)
}

func TestErrorBodyGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := Config{
		Kinds: []KindConfig{
			{Label: "kind", SkipBody: true},
		},
	}
	c := &Change{
		Body: "body",
		Kind: "kind",
	}

	err := c.AskPrompts(config, reader)
	then.Err(t, errKindDoesNotAcceptBody, err)
}

func TestErrorBodyGivenDoesNotPassValidation(t *testing.T) {
	var min int64 = 10

	reader, _ := then.WithReadWritePipe(t)
	config := Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
		Body: BodyConfig{
			MinLength: &min,
		},
	}
	c := &Change{
		Kind: "kind",
		Body: "body",
	}

	err := c.AskPrompts(config, reader)
	then.Err(t, errInputTooShort, err)
}

func TestSkipPromptForComponentIfSet(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("skip component body"),
		[]byte{13},
	)

	config := Config{
		Components: []string{"a", "b"},
	}
	c := &Change{
		Component: "a",
	}

	then.Nil(t, c.AskPrompts(config, reader))
	then.Equals(t, "a", c.Component)
	then.Equals(t, "skip component body", c.Body)
}

func TestSkipPromptForKindIfSet(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("skip kind body"),
		[]byte{13},
	)

	config := Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
	}
	c := &Change{
		Kind: "kind",
	}

	then.Nil(t, c.AskPrompts(config, reader))
	then.Equals(t, "kind", c.Kind)
	then.Equals(t, "skip kind body", c.Body)
}

func TestSkipPromptForBodyIfSet(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("kind"),
		[]byte{13},
	)

	config := Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
	}
	c := &Change{
		Body: "skip body body",
	}

	then.Nil(t, c.AskPrompts(config, reader))
	then.Equals(t, "kind", c.Kind)
	then.Equals(t, "skip body body", c.Body)
}
