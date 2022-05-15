package core

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/miniscruff/changie/testutils"
)

var _ = Describe("Change", func() {
	mockTime := func() time.Time {
		return time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local)
	}

	orderedTimes := []time.Time{
		time.Date(2018, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2017, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local),
	}

	It("should write a change", func() {
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

		Expect(builder.String()).To(Equal(changesYaml))
		Expect(err).To(BeNil())
	})

	It("should load change from path", func() {
		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal("some_file.yaml"))
			return []byte("kind: A\nbody: hey\n"), nil
		}

		change, err := LoadChange("some_file.yaml", mockRf)
		Expect(err).To(BeNil())
		Expect(change.Kind).To(Equal("A"))
		Expect(change.Body).To(Equal("hey"))
	})

	It("should return error from bad read", func() {
		mockErr := errors.New("bad file")

		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal("some_file.yaml"))
			return []byte(""), mockErr
		}

		_, err := LoadChange("some_file.yaml", mockRf)
		Expect(err).To(Equal(mockErr))
	})

	It("should return error from bad file", func() {
		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal("some_file.yaml"))
			return []byte("not a yaml file---"), nil
		}

		_, err := LoadChange("some_file.yaml", mockRf)
		Expect(err).NotTo(BeNil())
	})

	It("should sort by time", func() {
		changes := []Change{
			{Body: "third", Time: orderedTimes[2]},
			{Body: "second", Time: orderedTimes[1]},
			{Body: "first", Time: orderedTimes[0]},
		}
		config := Config{}
		SortByConfig(config).Sort(changes)

		Expect(changes[0].Body).To(Equal("third"))
		Expect(changes[1].Body).To(Equal("second"))
		Expect(changes[2].Body).To(Equal("first"))
	})

	It("should sort by kind then time", func() {
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

		Expect(changes[0].Body).To(Equal("first"))
		Expect(changes[1].Body).To(Equal("second"))
		Expect(changes[2].Body).To(Equal("third"))
		Expect(changes[3].Body).To(Equal("fourth"))
	})

	It("should sort by component then kind", func() {
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

		Expect(changes[0].Body).To(Equal("first"))
		Expect(changes[1].Body).To(Equal("second"))
		Expect(changes[2].Body).To(Equal("third"))
		Expect(changes[3].Body).To(Equal("fourth"))
	})
})

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

	It("for only body", func() {
		config := Config{}
		go func() {
			DelayWrite(stdinWriter, []byte("body stuff"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		c := &Change{}
		Expect(AskPrompts(c, config, stdinReader)).To(Succeed())
		Expect(c.Component).To(BeEmpty())
		Expect(c.Kind).To(BeEmpty())
		Expect(c.Custom).To(BeEmpty())
		Expect(c.Body).To(Equal("body stuff"))
	})

	It("for component, kind and body", func() {
		config := Config{
			Components: []string{"cli", "tests", "utils"},
			Kinds: []KindConfig{
				{Label: "added"},
				{Label: "changed"},
				{Label: "removed"},
			},
		}
		go func() {
			DelayWrite(stdinWriter, []byte{106, 13})
			DelayWrite(stdinWriter, []byte{106, 106, 13})
			DelayWrite(stdinWriter, []byte("body here"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		c := &Change{}
		Expect(AskPrompts(c, config, stdinReader)).To(Succeed())
		Expect(c.Component).To(Equal("tests"))
		Expect(c.Kind).To(Equal("removed"))
		Expect(c.Custom).To(BeEmpty())
		Expect(c.Body).To(Equal("body here"))
	})

	It("for body and custom", func() {
		config := Config{
			CustomChoices: []Custom{
				{Key: "check", Type: CustomString},
			},
		}
		go func() {
			DelayWrite(stdinWriter, []byte("body again"))
			DelayWrite(stdinWriter, []byte{13})
			DelayWrite(stdinWriter, []byte("custom check value"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		c := &Change{}
		Expect(AskPrompts(c, config, stdinReader)).To(Succeed())
		Expect(c.Component).To(BeEmpty())
		Expect(c.Kind).To(BeEmpty())
		Expect(c.Custom["check"]).To(Equal("custom check value"))
		Expect(c.Body).To(Equal("body again"))
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
		Expect(AskPrompts(c, config, stdinReader)).To(Succeed())
		Expect(c.Component).To(BeEmpty())
		Expect(c.Kind).To(Equal("added"))
		Expect(c.Custom).NotTo(HaveKey("skipped"))
		Expect(c.Custom).To(HaveKeyWithValue("break", "breaking value"))
		Expect(c.Body).To(BeEmpty())
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
		Expect(AskPrompts(c, config, stdinReader)).NotTo(Succeed())
		Expect(submitFailed).To(BeTrue())
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
		Expect(AskPrompts(c, config, stdinReader)).NotTo(Succeed())
	})

	It("gets error for invalid component", func() {
		config := Config{
			Components: []string{"a", "b"},
		}
		go func() {
			DelayWrite(stdinWriter, []byte{3})
		}()

		c := &Change{}
		Expect(AskPrompts(c, config, stdinReader)).NotTo(Succeed())
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
		Expect(AskPrompts(c, config, stdinReader)).NotTo(Succeed())
	})

	It("gets error for invalid body", func() {
		config := Config{}
		go func() {
			DelayWrite(stdinWriter, []byte{3})
		}()

		c := &Change{}
		Expect(AskPrompts(c, config, stdinReader)).NotTo(Succeed())
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
		Expect(AskPrompts(c, config, stdinReader)).NotTo(Succeed())
	})

	It("panics when trying to find kind from bad label", func() {
		config := Config{
			Kinds: []KindConfig{
				{Label: "a"},
			},
		}
		badKindFromLabel := func() {
			_ = kindFromLabel(config, "not a")
			Fail("should panic before this")
		}
		defer GinkgoRecover()
		Expect(badKindFromLabel).To(Panic())
	})
})
