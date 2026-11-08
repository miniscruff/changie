package core

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/miniscruff/changie/then"
)

func specificTimeNow() time.Time {
	return time.Date(2023, time.November, 31, 18, 30, 0, 0, time.UTC)
}

func TestAskPromptsForBody(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("body stuff"),
		[]byte{13},
	)

	config := &Config{}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}
	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	then.SliceLen(t, 1, changes)
	c := changes[0]
	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.MapLen(t, 0, c.Custom)
	then.Equals(t, "body stuff", c.Body)
}

func TestAskPromptsForBodyWithProject(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{32},
		[]byte{13},
		[]byte("body stuff"),
		[]byte{13},
	)

	config := &Config{
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
			{Label: "Other", Key: "other"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	then.SliceLen(t, 1, changes)
	c := changes[0]
	then.Equals(t, "client", c.Project)
	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.Equals(t, 0, len(c.Custom))
	then.Equals(t, "body stuff", c.Body)
}

func TestAskPromptsForBodyWithProjectErrBadProject(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)

	config := &Config{
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
			{Label: "Other", Key: "other"},
		},
	}

	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Projects:      []string{"missing"},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errProjectNotFound, err)
}

func TestAskPromptsFailIfDisabled(t *testing.T) {
	config := &Config{
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
			{Label: "Other", Key: "other"},
		},
		Components: []string{"cli", "tests", "utils"},
		Kinds: []KindConfig{
			{Label: "added"},
			{Label: "changed"},
			{Label: "removed"},
		},
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
		},
	}

	t.Run("project", func(t *testing.T) {
		prompts := &Prompts{
			Config:        config,
			StdinReader:   nil,
			TimeNow:       specificTimeNow,
			Enabled:       false,
			LegacyPrompts: promptsLegacy,
		}

		_, err := prompts.BuildChanges()
		then.Err(t, errProjectMissingPromptDisabled, err)
	})

	t.Run("component", func(t *testing.T) {
		prompts := &Prompts{
			Config:        config,
			StdinReader:   nil,
			TimeNow:       specificTimeNow,
			Enabled:       false,
			LegacyPrompts: promptsLegacy,
		}

		prompts.Projects = []string{"other"}
		_, err := prompts.BuildChanges()
		then.Err(t, errComponentMissingPromptDisabled, err)
	})

	t.Run("kind", func(t *testing.T) {
		prompts := &Prompts{
			Config:        config,
			StdinReader:   nil,
			TimeNow:       specificTimeNow,
			Enabled:       false,
			LegacyPrompts: promptsLegacy,
		}

		prompts.Projects = []string{"other"}
		prompts.Component = "cli"

		_, err := prompts.BuildChanges()
		then.Err(t, errKindMissingPromptDisabled, err)
	})

	t.Run("body", func(t *testing.T) {
		prompts := &Prompts{
			Config:        config,
			StdinReader:   nil,
			TimeNow:       specificTimeNow,
			Enabled:       false,
			LegacyPrompts: promptsLegacy,
		}

		prompts.Projects = []string{"other"}
		prompts.Component = "cli"
		prompts.Kind = "added"

		_, err := prompts.BuildChanges()
		then.Err(t, errBodyMissingPromptDisabled, err)
	})

	t.Run("custom", func(t *testing.T) {
		prompts := &Prompts{
			Config:        config,
			StdinReader:   nil,
			TimeNow:       specificTimeNow,
			Enabled:       false,
			LegacyPrompts: promptsLegacy,
		}

		prompts.Projects = []string{"other"}
		prompts.Component = "cli"
		prompts.Kind = "added"
		prompts.Body = "some body"

		_, err := prompts.BuildChanges()
		then.Err(t, errCustomMissingPromptDisabled, err)
	})
}

func TestAskPromptsForBodyWithProjectErrBadInput(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{3},
	)

	config := &Config{
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
			{Label: "Other", Key: "other"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}

func TestAskPromptsProjectRequiredBeforeContinuing(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		// enter without a project shows an error and stays on projects
		[]byte{13},
		[]byte{32},
		[]byte{13},
		[]byte("body stuff"),
		[]byte{13},
	)

	config := &Config{
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
			{Label: "Other", Key: "other"},
		},
	}
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Enabled:     true,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)
	then.SliceLen(t, 1, changes)
	then.Equals(t, "client", changes[0].Project)
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

	config := &Config{
		Components: []string{"cli", "tests", "utils"},
		Kinds: []KindConfig{
			{Label: "added"},
			{Label: "changed"},
			{Label: "removed"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}
	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]
	then.Equals(t, "tests", c.Component)
	then.Equals(t, "removed", c.Kind)
	then.Equals(t, len(c.Custom), 0)
	then.Equals(t, "body here", c.Body)
}

func TestBodyKindPostProcess(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)

	config := &Config{
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
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Body:        "our body",
		Kind:        "added",
		Customs: map[string]string{
			"Issue": "25",
		},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}
	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]
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

	config := &Config{
		CustomChoices: []Custom{
			{Key: "check", Type: CustomString},
		},
	}

	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}
	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]

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

	config := &Config{
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomString},
			{Key: "Project", Label: "Custom Project Prompt", Type: CustomString},
		},
	}
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Customs: map[string]string{
			"Project": "Changie",
		},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}
	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]

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

	config := &Config{
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

	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}
	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Custom["skipped"])
	then.Equals(t, "breaking value", c.Custom["break"])
	then.Equals(t, "", c.Body)
}

func TestBodyAndPostProcess(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)

	config := &Config{
		Post: []PostProcessConfig{
			{Key: "Post", Value: "Body again {{.Body}}"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Body:          "our body",
		Customs:       make(map[string]string),
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.Equals(t, "Body again our body", c.Custom["Post"])
	then.Equals(t, "our body", c.Body)
}

func TestBodyAndPostProcessSkipGlobalPost(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
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

	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Body:        "our body",
		Kind:        "added",
		Customs: map[string]string{
			"Issue": "30",
		},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]

	then.Equals(t, "", c.Component)
	then.Equals(t, "added", c.Kind)
	then.Equals(t, "our body", c.Body)
	then.Equals(t, "our body+30", c.Custom["Post"])
	then.Equals(t, "", c.Custom["GlobalPost"])
}

func TestAskPromptsInvalidBodyBeforeContinuing(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		// too short shows an error and stays on the body
		[]byte("abc"),
		[]byte{13},
		[]byte("de"),
		[]byte{13},
	)

	var min int64 = 5

	config := &Config{
		Body: BodyConfig{
			MinLength: &min,
		},
	}

	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Enabled:     true,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)
	then.Equals(t, "abcde", changes[0].Body)
}

func TestErrorInvalidPost(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Post: []PostProcessConfig{
			{Key: "Post", Value: "invalid {{++...thing}}"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Body:          "our body",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}

func TestErrorInvalidPostWithProjects(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Post: []PostProcessConfig{
			{Key: "Post", Value: "invalid {{++...thing}}"},
		},
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
			{Label: "Other", Key: "other"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Body:          "our body",
		Projects:      []string{"client"},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}

func TestErrorInvalidCustomType(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("invalid custom type"),
		[]byte{13},
	)

	config := &Config{
		CustomChoices: []Custom{
			{Key: "check", Type: "bad type", Label: "a"},
		},
	}

	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}

func TestErrorInvalidComponent(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{3},
	)

	config := &Config{
		Components: []string{"a", "b"},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}

func TestErrorInvalidKind(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{3},
	)

	config := &Config{
		Kinds: []KindConfig{
			{Label: "a"},
			{Label: "b"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}

func TestErrorFaultBody(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{3},
	)

	config := &Config{}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
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

	config := &Config{
		CustomChoices: []Custom{
			{Key: "check", Type: CustomString, Label: "a"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}

func TestErrorBadKindLabel(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Kind:          "not kind",
		Body:          "body",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errInvalidKind, err)
}

func TestErrorBadComponentInput(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Components: []string{"a", "b", "c"},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Component:     "d",
		Body:          "body",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errInvalidComponent, err)
}

func TestErrorComponentGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Component:     "we shouldn't have a component",
		Body:          "body",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errComponentProvidedWhenNotConfigured, err)
}

func TestErrorKindGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Kind:          "we shouldn't have a kind",
		Body:          "body",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errKindProvidedWhenNotConfigured, err)
}

func TestErrorCustomGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
		},
	}
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Customs: map[string]string{
			"MissingKey": "40",
		},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errCustomProvidedNotConfigured, err)
}

func TestErrorCustomGivenDoesNotPassValidation(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	minValue := int64(50)
	config := &Config{
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt, MinInt: &minValue},
		},
	}
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Customs: map[string]string{
			"Issue": "40",
		},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errIntTooLow, err)
}

func TestErrorBodyGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Kinds: []KindConfig{
			{Label: "kind", SkipBody: true},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Body:          "body",
		Kind:          "kind",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errKindDoesNotAcceptBody, err)
}

func TestErrorBodyGivenDoesNotPassValidation(t *testing.T) {
	var min int64 = 10

	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
		Body: BodyConfig{
			MinLength: &min,
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Kind:          "kind",
		Body:          "body",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errInputTooShort, err)
}

func TestSkipPromptForComponentIfSet(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("skip component body"),
		[]byte{13},
	)

	config := &Config{
		Components: []string{"a", "b"},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Component:     "a",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]
	then.Equals(t, "a", c.Component)
	then.Equals(t, "skip component body", c.Body)
}

func TestSkipPromptForCustomIfOptionalAndDisabled(t *testing.T) {
	config := &Config{
		CustomChoices: []Custom{
			{
				Key:      "cust",
				Optional: true,
				Type:     "string",
			},
		},
	}
	prompts := &Prompts{
		Config:        config,
		TimeNow:       specificTimeNow,
		Enabled:       false,
		LegacyPrompts: promptsLegacy,
		Body:          "skip custom input",
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]
	then.Equals(t, "skip custom input", c.Body)
	then.MapLen(t, 0, c.Custom)
	then.Equals(t, "", c.Custom["cust"])
}

func TestSkipPromptForPromptsWithCustomPromptsInKindConfig(t *testing.T) {
	config := &Config{
		Kinds: []KindConfig{
			{
				Key: "dependency",
				AdditionalChoices: []Custom{
					{
						Key:  "name",
						Type: CustomString,
					},
					{
						Key:  "from",
						Type: CustomString,
					},
					{
						Key:  "to",
						Type: CustomString,
					},
				},
				SkipBody: true,
			},
		},
	}
	prompts := &Prompts{
		Config:      config,
		StdinReader: bytes.NewReader(nil),
		TimeNow:     specificTimeNow,
		Kind:        "dependency",
		Customs: map[string]string{
			"name": "go",
			"from": "1.20",
			"to":   "1.22",
		},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]
	then.Equals(t, "go", c.Custom["name"])
	then.Equals(t, "1.20", c.Custom["from"])
	then.Equals(t, "1.22", c.Custom["to"])
}

func TestSkipPromptForKindIfSet(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("skip kind body"),
		[]byte{13},
	)

	config := &Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Kind:          "kind",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]
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

	config := &Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Body:          "skip body body",
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]
	then.Equals(t, "kind", c.Kind)
	then.Equals(t, "skip body body", c.Body)
}

func TestGetBodyTxtWithEditor(t *testing.T) {
	then.WithTempDir(t)
	t.Setenv("EDITOR", "vim")

	expectedBodyMsg := "some body msg"
	config := &Config{}

	prompts := &Prompts{
		Config:     config,
		BodyEditor: true,
		TimeNow:    specificTimeNow,
		EditorCmdBuilder: func(filename string) (EditorRunner, error) {
			return &dummyEditorRunner{
				filename: filename,
				body:     []byte(expectedBodyMsg),
				t:        t,
			}, nil
		},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.MapLen(t, 0, c.Custom)
	then.Equals(t, expectedBodyMsg, c.Body)
}

func TestGetBodyTxtWithEditorUnableToCreateCmd(t *testing.T) {
	t.Setenv("EDITOR", "")
	then.WithTempDir(t)

	config := &Config{}
	prompts := &Prompts{
		Config:     config,
		BodyEditor: true,
		TimeNow:    specificTimeNow,
		EditorCmdBuilder: func(s string) (EditorRunner, error) {
			return BuildCommand(s)
		},
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}

func TestRequireProvided(t *testing.T) {
	config := &Config{
		Kinds: []KindConfig{
			{Label: "Added", Key: "added"},
			{
				Label:             "Dependencies",
				Key:               "deps",
				SkipBody:          true,
				SkipGlobalChoices: true,
				AdditionalChoices: []Custom{
					{Key: "Package", Type: CustomString},
					{Key: "Version", Type: CustomString, Optional: true},
				},
			},
		},
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
			{Key: "Author", Type: CustomString, Optional: true},
		},
	}

	t.Run("all provided", func(t *testing.T) {
		p := &Prompts{
			Config:  config,
			Kind:    "added",
			Body:    "body",
			Customs: map[string]string{"Issue": "15"},
		}

		then.Nil(t, p.requireProvided())
	})

	t.Run("kind skips body and global choices", func(t *testing.T) {
		p := &Prompts{
			Config:  config,
			Kind:    "deps",
			Customs: map[string]string{"Package": "huh"},
		}

		then.Nil(t, p.requireProvided())
	})

	t.Run("kind additional choice missing", func(t *testing.T) {
		p := &Prompts{
			Config: config,
			Kind:   "deps",
		}

		then.Err(t, errCustomMissingPromptDisabled, p.requireProvided())
	})
}

var (
	keyDown     = []byte("\x1b[B")
	keyShiftTab = []byte("\x1b[Z")
)

func TestAskPromptsGoBackAndChangeKind(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		// pick added, type a body, then go back and pick dependencies
		[]byte{13},
		[]byte("body for added"),
		keyShiftTab,
		keyDown,
		[]byte{13},
		[]byte("huh"),
		[]byte{13},
	)

	config := &Config{
		Kinds: []KindConfig{
			{Label: "Added", Key: "added"},
			{
				Label:             "Dependencies",
				Key:               "deps",
				SkipBody:          true,
				AdditionalChoices: []Custom{{Key: "Package", Type: CustomString}},
			},
		},
	}
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Enabled:     true,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	then.SliceLen(t, 1, changes)
	c := changes[0]
	then.Equals(t, "deps", c.Kind)
	// the body typed for added is not used as dependencies skips the body
	then.Equals(t, "", c.Body)
	then.MapEquals(t, map[string]string{"Package": "huh"}, c.Custom)
}

func TestAskPromptsKeepsProvidedCustoms(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("body stuff"),
		[]byte{13},
		[]byte("miniscruff"),
		[]byte{13},
	)

	config := &Config{
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
			{Key: "Author", Type: CustomString},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
		Customs:       map[string]string{"Issue": "15"},
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)
	then.MapEquals(t, map[string]string{"Issue": "15", "Author": "miniscruff"}, changes[0].Custom)
}

func TestAskPromptsNothingToAsk(t *testing.T) {
	prompts := &Prompts{
		Config:        &Config{},
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
		Body:          "body from flag",
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)
	then.Equals(t, "body from flag", changes[0].Body)
}

func TestAskPromptsProjectWithoutKey(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		keyDown,
		[]byte{32},
		[]byte{13},
		[]byte("body stuff"),
		[]byte{13},
	)

	config := &Config{
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
			{Label: "Frontend"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)
	then.SliceLen(t, 1, changes)
	// projects without a key are saved with an empty project, as before
	then.Equals(t, "", changes[0].Project)
}

func TestAskPromptsKindWithoutKey(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		keyDown,
		[]byte{13},
		[]byte("CVE-1"),
		[]byte{13},
	)

	config := &Config{
		Kinds: []KindConfig{
			{Label: "Dependencies", SkipBody: true, SkipGlobalChoices: true},
			{
				Label:             "Security",
				SkipBody:          true,
				SkipGlobalChoices: true,
				AdditionalChoices: []Custom{{Key: "CVE", Type: CustomString}},
			},
		},
		CustomChoices: []Custom{{Key: "Issue", Type: CustomInt}},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: promptsLegacy,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)
	then.SliceLen(t, 1, changes)
	// kinds without a key are found by label for hiding fields and saving
	then.Equals(t, "Security", changes[0].Kind)
	then.Equals(t, "", changes[0].Body)
	then.MapEquals(t, map[string]string{"CVE": "CVE-1"}, changes[0].Custom)
}

func TestAskPromptsForBodyWithProjectErrNoProject(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{13},
	)

	config := &Config{
		Projects: []ProjectConfig{
			{Label: "Client", Key: "client"},
			{Label: "Other", Key: "other"},
		},
	}
	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: true,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errProjectRequired, err)
}

func TestErrorInvalidBody(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("abc"),
		[]byte{13},
	)

	var min int64 = 5

	config := &Config{
		Body: BodyConfig{
			MinLength: &min,
		},
	}

	prompts := &Prompts{
		Config:        config,
		StdinReader:   reader,
		TimeNow:       specificTimeNow,
		Enabled:       true,
		LegacyPrompts: true,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errInputTooShort, err)
}

func TestNewPromptsWarning(t *testing.T) {
	for _, tc := range []struct {
		name    string
		prompts Prompts
		warns   bool
	}{
		{name: "huh form shown", prompts: Prompts{Enabled: true, ShowNewPromptsWarning: true}, warns: true},
		{name: "warning hidden", prompts: Prompts{Enabled: true}},
		{name: "legacy prompts", prompts: Prompts{Enabled: true, LegacyPrompts: true, ShowNewPromptsWarning: true}},
		{name: "prompts disabled", prompts: Prompts{Body: "from flag", ShowNewPromptsWarning: true}},
		{name: "nothing to ask", prompts: Prompts{Enabled: true, Body: "from flag", ShowNewPromptsWarning: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader, writer := then.WithReadWritePipe(t)
			if tc.prompts.Enabled && tc.prompts.Body == "" {
				then.DelayWrite(t, writer, []byte("body stuff"), []byte{13})
			}

			var output bytes.Buffer

			prompts := tc.prompts
			prompts.Config = &Config{}
			prompts.StdinReader = reader
			prompts.TimeNow = specificTimeNow
			prompts.Output = &output

			_, err := prompts.BuildChanges()
			then.Nil(t, err)

			// the form is drawn after the warning on the same output
			then.Equals(t, tc.warns, strings.HasPrefix(output.String(), newPromptsWarning+"\n"))
			then.Equals(t, tc.warns, strings.Contains(output.String(), newPromptsWarning))
		})
	}
}

func TestFormInput(t *testing.T) {
	t.Run("pipe is read as a plain reader", func(t *testing.T) {
		reader, writer, err := os.Pipe()
		then.Nil(t, err)

		t.Cleanup(func() {
			_ = reader.Close()
			_ = writer.Close()
		})

		_, isFile := formInput(reader).(*os.File)
		then.False(t, isFile)
	})

	t.Run("other readers are unchanged", func(t *testing.T) {
		reader := strings.NewReader("input")
		then.Equals[io.Reader](t, reader, formInput(reader))
	})
}
