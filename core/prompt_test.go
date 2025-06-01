package core

import (
	"bytes"
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Projects:    []string{"missing"},
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errProjectNotFound, err)
}

func TestAskPromptsFailIfDisabled(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("body stuff"),
		[]byte{13},
	)

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
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Enabled:     false,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errProjectMissingPromptDisabled, err)

	prompts.Projects = []string{"other"}
	_, err = prompts.BuildChanges()
	then.Err(t, errComponentMissingPromptDisabled, err)

	prompts.Component = "cli"
	_, err = prompts.BuildChanges()
	then.Err(t, errKindMissingPromptDisabled, err)

	prompts.Kind = "added"
	_, err = prompts.BuildChanges()
	then.Err(t, errBodyMissingPromptDisabled, err)

	prompts.Body = "some body"
	_, err = prompts.BuildChanges()
	then.Err(t, errCustomMissingPromptDisabled, err)
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errProjectRequired, err)
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Body:        "our body",
		Customs:     make(map[string]string),
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errInputTooShort, err)
}

func TestErrorInvalidPost(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Post: []PostProcessConfig{
			{Key: "Post", Value: "invalid {{++...thing}}"},
		},
	}
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Body:        "our body",
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Body:        "our body",
		Projects:    []string{"client"},
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Kind:        "not kind",
		Body:        "body",
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Component:   "d",
		Body:        "body",
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errInvalidComponent, err)
}

func TestErrorComponentGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{}
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Component:   "we shouldn't have a component",
		Body:        "body",
	}

	_, err := prompts.BuildChanges()
	then.Err(t, errComponentProvidedWhenNotConfigured, err)
}

func TestErrorKindGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{}
	prompts := &Prompts{
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Kind:        "we shouldn't have a kind",
		Body:        "body",
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Body:        "body",
		Kind:        "kind",
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Kind:        "kind",
		Body:        "body",
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Component:   "a",
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	c := changes[0]
	then.Equals(t, "a", c.Component)
	then.Equals(t, "skip component body", c.Body)
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Kind:        "kind",
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
		Config:      config,
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Body:        "skip body body",
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
	}

	_, err := prompts.BuildChanges()
	then.NotNil(t, err)
}
