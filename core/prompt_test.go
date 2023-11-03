package core

import (
	"errors"
	"os"
	"testing"

	"github.com/miniscruff/changie/then"
)

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

	config := &Config{}
	c := &PromptChange{}
	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.Equals(t, 0, len(c.Custom))
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
	c := &PromptChange{}
	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

	then.SliceEquals(t, []string{"client"}, c.Projects)
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
	c := &PromptChange{
		Projects: []string{"missing"},
	}
	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
	then.Err(t, errProjectNotFound, err)
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
	c := &PromptChange{}
	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
	then.NotNil(t, err)
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
	c := &PromptChange{}
	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

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
	c := &PromptChange{
		Body: "our body",
		Kind: "added",
		Custom: map[string]string{
			"Issue": "25",
		},
	}

	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

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
	c := &PromptChange{}
	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

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
	c := &PromptChange{
		Custom: map[string]string{
			"Project": "Changie",
		},
	}

	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

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

	c := &PromptChange{}
	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

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

	c := &PromptChange{
		Body: "our body",
	}

	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

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

	c := &PromptChange{
		Body: "our body",
		Kind: "added",
		Custom: map[string]string{
			"Issue": "30",
		},
	}

	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))

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
	c := &PromptChange{}
	then.NotNil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
}

func TestErrorInvalidPost(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Post: []PostProcessConfig{
			{Key: "Post", Value: "invalid {{++...thing}}"},
		},
	}
	c := &PromptChange{
		Body: "our body",
	}
	then.NotNil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
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

	c := &PromptChange{}
	then.NotNil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
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
	c := &PromptChange{}
	then.NotNil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
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
	c := &PromptChange{}
	then.NotNil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
}

func TestErrorFaultBody(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{3},
	)

	config := &Config{}
	c := &PromptChange{}
	then.NotNil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
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
	c := &PromptChange{}
	then.NotNil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
}

func TestErrorBadKindLabel(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
	}
	c := &PromptChange{
		Kind: "not kind",
		Body: "body",
	}

	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
	then.Err(t, errInvalidKind, err)
}

func TestErrorBadComponentInput(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Components: []string{"a", "b", "c"},
	}
	c := &PromptChange{
		Component: "d",
		Body:      "body",
	}

	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
	then.Err(t, errInvalidComponent, err)
}

func TestErrorComponentGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{}
	c := &PromptChange{
		Component: "we shouldn't have a component",
		Body:      "body",
	}

	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
	then.Err(t, errComponentProvidedWhenNotConfigured, err)
}

func TestErrorKindGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{}
	c := &PromptChange{
		Kind: "we shouldn't have a kind",
		Body: "body",
	}

	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
	then.Err(t, errKindProvidedWhenNotConfigured, err)
}

func TestErrorCustomGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		CustomChoices: []Custom{
			{Key: "Issue", Type: CustomInt},
		},
	}
	c := &PromptChange{
		Custom: map[string]string{
			"MissingKey": "40",
		},
	}

	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
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
	c := &PromptChange{
		Custom: map[string]string{
			"Issue": "40",
		},
	}

	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
	then.Err(t, errIntTooLow, err)
}

func TestErrorBodyGivenWithNoConfiguration(t *testing.T) {
	reader, _ := then.WithReadWritePipe(t)
	config := &Config{
		Kinds: []KindConfig{
			{Label: "kind", SkipBody: true},
		},
	}
	c := &PromptChange{
		Body: "body",
		Kind: "kind",
	}

	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
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
	c := &PromptChange{
		Kind: "kind",
		Body: "body",
	}

	err := c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	})
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
	c := &PromptChange{
		Component: "a",
	}

	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
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

	config := &Config{
		Kinds: []KindConfig{
			{Label: "kind"},
		},
	}
	c := &PromptChange{
		Kind: "kind",
	}

	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
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
	c := &PromptChange{
		Body: "skip body body",
	}

	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      config,
		StdinReader: reader,
	}))
	then.Equals(t, "kind", c.Kind)
	then.Equals(t, "skip body body", c.Body)
}

func TestGetBodyTxtWithEditor(t *testing.T) {
	then.WithTempDir(t)
	t.Setenv("EDITOR", "vim")

	expectedBodyMsg := "some body msg"
	c := &PromptChange{}

	then.Nil(t, c.AskPrompts(PromptContext{
		Config:      &Config{},
		BodyEditor:  true,
		CreateFiler: os.Create,
		EditorCmdBuilder: func(filename string) (EditorRunner, error) {
			return &dummyEditorRunner{
				filename: filename,
				body:     []byte(expectedBodyMsg),
				t:        t,
			}, nil
		},
	}))

	then.Equals(t, "", c.Component)
	then.Equals(t, "", c.Kind)
	then.Equals(t, 0, len(c.Custom))
	then.Equals(t, expectedBodyMsg, c.Body)
}

func TestGetBodyTxtWithEditorUnableToCreateCmd(t *testing.T) {
	t.Setenv("EDITOR", "")
	then.WithTempDir(t)

	c := &PromptChange{}
	err := c.AskPrompts(PromptContext{
		Config:      &Config{},
		BodyEditor:  true,
		CreateFiler: os.Create,
		EditorCmdBuilder: func(s string) (EditorRunner, error) {
			return BuildCommand(s)
		},
	})

	then.NotNil(t, err)
}

func TestGetBodyTxtWithEditorUnableToCreateTempFile(t *testing.T) {
	c := &PromptChange{}
	mockErr := errors.New("bad create file")
	err := c.AskPrompts(PromptContext{
		Config:     &Config{},
		BodyEditor: true,
		CreateFiler: func(filename string) (*os.File, error) {
			return nil, mockErr
		},
	})

	then.Err(t, mockErr, err)
}
