package core

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/miniscruff/changie/testutils"
	"github.com/miniscruff/changie/then"
)

var (
	orderedTimes = []time.Time{
		time.Date(2018, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2017, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local),
	}
)

func mockTime() time.Time {
	return time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local)
}

func TestWriteChange(t *testing.T) {
	change := Change{
		Body: "some body message",
		Time: mockTime(),
	}

	changesYaml := fmt.Sprintf(
		"body: some body message\ntime: %s\n",
		mockTime().Format(time.RFC3339Nano),
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
	reader, writer := then.WithStdIn(t)

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
	reader, writer := then.WithStdIn(t)

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
	reader, _ := then.WithStdIn(t)

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
	reader, writer := then.WithStdIn(t)

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
	reader, writer := then.WithStdIn(t)

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

/*
func TestSkippedBodyGlobalChoicesKindWithAdditional(t *testing.T) {
	writer, reader := then.WithStdIn(t)

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
*/

var _ = Describe("Change ask prompts", func() {
	var (
		stdinReader *os.File
		stdinWriter *os.File
	)

	BeforeEach(func() {
		var pipeErr error
		stdinReader, stdinWriter, pipeErr = os.Pipe()
		Expect(pipeErr).To(BeNil())
	})

	AfterEach(func() {
		stdinReader.Close()
		stdinWriter.Close()
	})

	It("for skipped body, skipped global choices and kind choices", func() {
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
		go func() {
			DelayWrite(stdinWriter, []byte{13})
			DelayWrite(stdinWriter, []byte("breaking value"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		c := &Change{}
		Expect(c.AskPrompts(config, stdinReader)).To(Succeed())
		Expect(c.Component).To(BeEmpty())
		Expect(c.Kind).To(Equal("added"))
		Expect(c.Custom).NotTo(HaveKey("skipped"))
		Expect(c.Custom).To(HaveKeyWithValue("break", "breaking value"))
		Expect(c.Body).To(BeEmpty())
	})

	It("for body and post process", func() {
		config := Config{
			Post: []PostProcessConfig{
				{Key: "Post", Value: "Body again {{.Body}}"},
			},
		}

		c := &Change{
			Body: "our body",
		}
		Expect(c.AskPrompts(config, stdinReader)).To(Succeed())
		Expect(c.Component).To(BeEmpty())
		Expect(c.Kind).To(BeEmpty())
		Expect(c.Body).To(Equal("our body"))
		Expect(c.Custom["Post"]).To(Equal("Body again our body"))
	})

	It("for body and post process skipping global post", func() {
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

		Expect(c.AskPrompts(config, stdinReader)).To(Succeed())
		Expect(c.Component).To(BeEmpty())
		Expect(c.Kind).To(Equal("added"))
		Expect(c.Body).To(Equal("our body"))
		Expect(c.Custom["GlobalPost"]).To(BeEmpty())
		Expect(c.Custom["Post"]).To(Equal("our body+30"))
	})

	It("gets error for invalid body", func() {
		var min int64 = 5
		submitFailed := false
		config := Config{
			Body: BodyConfig{
				MinLength: &min,
			},
		}
		go func() {
			DelayWrite(stdinWriter, []byte("abc"))
			DelayWrite(stdinWriter, []byte{13})
			// we need to ctrl-c out of the prompt or it will stick
			DelayWrite(stdinWriter, []byte{3})
			// use this boolean to check the control-c was required
			submitFailed = true
		}()

		c := &Change{}
		Expect(c.AskPrompts(config, stdinReader)).NotTo(Succeed())
		Expect(submitFailed).To(BeTrue())
	})

	It("gets error for invalid post", func() {
		config := Config{
			Post: []PostProcessConfig{
				{Key: "Post", Value: "invalid {{++...thing}}"},
			},
		}

		c := &Change{
			Body: "our body",
		}
		Expect(c.AskPrompts(config, stdinReader)).NotTo(Succeed())
	})

	It("gets error for invalid custom type", func() {
		config := Config{
			CustomChoices: []Custom{
				{Key: "check", Type: "bad type", Label: "a"},
			},
		}
		go func() {
			DelayWrite(stdinWriter, []byte("body again"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		c := &Change{}
		Expect(c.AskPrompts(config, stdinReader)).NotTo(Succeed())
	})

	It("gets error for invalid component", func() {
		config := Config{
			Components: []string{"a", "b"},
		}
		go func() {
			DelayWrite(stdinWriter, []byte{3})
		}()

		c := &Change{}
		Expect(c.AskPrompts(config, stdinReader)).NotTo(Succeed())
	})

	It("gets error for invalid kind", func() {
		config := Config{
			Kinds: []KindConfig{
				{Label: "a"},
				{Label: "b"},
			},
		}
		go func() {
			DelayWrite(stdinWriter, []byte{3})
		}()

		c := &Change{}
		Expect(c.AskPrompts(config, stdinReader)).NotTo(Succeed())
	})

	It("gets error for invalid body", func() {
		config := Config{}
		go func() {
			DelayWrite(stdinWriter, []byte{3})
		}()

		c := &Change{}
		Expect(c.AskPrompts(config, stdinReader)).NotTo(Succeed())
	})

	It("gets error for invalid custom value", func() {
		config := Config{
			CustomChoices: []Custom{
				{Key: "check", Type: CustomString, Label: "a"},
			},
		}
		go func() {
			DelayWrite(stdinWriter, []byte("body again"))
			DelayWrite(stdinWriter, []byte{13})
			DelayWrite(stdinWriter, []byte("control C out"))
			DelayWrite(stdinWriter, []byte{3})
		}()

		c := &Change{}
		Expect(c.AskPrompts(config, stdinReader)).NotTo(Succeed())
	})

	It("errors when trying to find kind from bad label", func() {
		config := Config{
			Kinds: []KindConfig{
				{Label: "kind"},
			},
		}

		c := &Change{
			Kind: "not kind",
			Body: "body",
		}

		err := c.AskPrompts(config, stdinReader)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError(errInvalidKind))
	})

	It("errors when trying to find component from bad input", func() {
		config := Config{
			Components: []string{"a", "b", "c"},
		}

		c := &Change{
			Component: "d",
			Body:      "body",
		}

		err := c.AskPrompts(config, stdinReader)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError(errInvalidComponent))
	})

	It("doesn't prompt for component if it's already set", func() {
		config := Config{
			Components: []string{"a", "b"},
		}

		c := &Change{
			Component: "a",
		}

		go func() {
			DelayWrite(stdinWriter, []byte("body"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		Expect(c.AskPrompts(config, stdinReader)).To(Succeed())
		Expect(c.Component).To(Equal("a"))
		Expect(c.Body).To(Equal("body"))
	})

	It("doesn't prompt for kind if it's already set", func() {
		config := Config{
			Kinds: []KindConfig{
				{Label: "kind"},
			},
		}

		c := &Change{
			Kind: "kind",
		}

		go func() {
			DelayWrite(stdinWriter, []byte("body"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		Expect(c.AskPrompts(config, stdinReader)).To(Succeed())
		Expect(c.Kind).To(Equal("kind"))
		Expect(c.Body).To(Equal("body"))
	})

	It("doesn't prompt for body if it's already set", func() {
		config := Config{
			Kinds: []KindConfig{
				{Label: "kind"},
			},
		}

		c := &Change{
			Body: "body",
		}

		go func() {
			DelayWrite(stdinWriter, []byte("kind"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		Expect(c.AskPrompts(config, stdinReader)).To(Succeed())
		Expect(c.Kind).To(Equal("kind"))
		Expect(c.Body).To(Equal("body"))
	})

	It("errors when component is given when no configuration is used", func() {
		config := Config{}

		c := &Change{
			Component: "we shouldn't have a component",
			Body:      "body",
		}

		err := c.AskPrompts(config, stdinReader)
		Expect(err).To(MatchError(errComponentProvidedWhenNotConfigured))
	})

	It("errors when kind is given when no configured is used", func() {
		config := Config{}

		c := &Change{
			Kind: "we shouldn't have a kind",
			Body: "body",
		}

		err := c.AskPrompts(config, stdinReader)
		Expect(err).To(MatchError(errKindProvidedWhenNotConfigured))
	})

	It("errors when custom is given when not present", func() {
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

		err := c.AskPrompts(config, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("errors when custom is given that does not pass validation", func() {
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

		err := c.AskPrompts(config, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("errors when body is given for a kind that shouldn't have one", func() {
		config := Config{
			Kinds: []KindConfig{
				{Label: "kind", SkipBody: true},
			},
		}

		c := &Change{
			Body: "body",
			Kind: "kind",
		}

		err := c.AskPrompts(config, stdinReader)
		Expect(err).To(MatchError(errKindDoesNotAcceptBody))
	})

	It("validates body without prompt", func() {
		var min int64 = 10

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

		err := c.AskPrompts(config, stdinReader)
		Expect(err).To(MatchError(errInputTooShort))
	})
})
