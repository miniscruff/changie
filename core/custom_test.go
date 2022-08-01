package core

import (
	"errors"
	"os"

	"github.com/manifoldco/promptui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/miniscruff/changie/testutils"
)

var _ = Describe("Custom", func() {
	It("display label defaults to Key", func() {
		Expect(Custom{Key: "key"}.DisplayLabel()).To(Equal("key"))
	})

	It("display label uses Label if defined", func() {
		Expect(Custom{Key: "key", Label: "Lab"}.DisplayLabel()).To(Equal("Lab"))
	})

	It("returns error on invalid prompt type", func() {
		_, err := Custom{
			Type: "invalid type",
		}.CreatePrompt(os.Stdin)
		Expect(errors.Is(err, errInvalidPromptType)).To(BeTrue())
	})

	It("returns error on invalid prompt type from .Validate()", func() {
		err := Custom{
			Type: "invalid type",
		}.Validate("blah")
		Expect(errors.Is(err, errInvalidPromptType)).To(BeTrue())
	})

	It("can create custom string prompt", func() {
		prompt, err := Custom{
			Type:  CustomString,
			Label: "a label",
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Label).To(Equal("a label"))
	})

	It("can create custom string prompt with min and max length", func() {
		var min int64 = 3
		var max int64 = 15
		longInput := "longer string than is allowed by rule"
		prompt, err := Custom{
			Type:      CustomString,
			Label:     "a label",
			MinLength: &min,
			MaxLength: &max,
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Validate("average")).To(BeNil())
		Expect(errors.Is(underPrompt.Validate(""), errInputTooShort)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("a"), errInputTooShort)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate(longInput), errInputTooLong)).To(BeTrue())
	})

	It("can create optional custom string prompt with min and max length", func() {
		var min int64 = 3
		var max int64 = 15
		longInput := "longer string than is allowed by rule"
		prompt, err := Custom{
			Type:      CustomString,
			Label:     "a label",
			Optional:  true,
			MinLength: &min,
			MaxLength: &max,
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Validate("average")).To(BeNil())
		Expect(underPrompt.Validate("")).To(BeNil())
		Expect(errors.Is(underPrompt.Validate("a"), errInputTooShort)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate(longInput), errInputTooLong)).To(BeTrue())
	})

	It("can create custom int prompt", func() {
		prompt, err := Custom{
			Key:  "name",
			Type: CustomInt,
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Validate("12")).To(BeNil())
		Expect(underPrompt.Label).To(Equal("name"))
	})

	It("can create custom int prompt with a min and max", func() {
		var min int64 = 5
		var max int64 = 10
		prompt, err := Custom{
			Type:   CustomInt,
			MinInt: &min,
			MaxInt: &max,
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Validate("7")).To(BeNil())
		Expect(errors.Is(underPrompt.Validate(""), errInvalidIntInput)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("not an int"), errInvalidIntInput)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("12.5"), errInvalidIntInput)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("3"), errIntTooLow)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("25"), errIntTooHigh)).To(BeTrue())
	})

	It("can create custom optional int prompt", func() {
		var min int64 = 5
		var max int64 = 10
		prompt, err := Custom{
			Type:     CustomInt,
			Optional: true,
			MinInt:   &min,
			MaxInt:   &max,
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Validate("7")).To(BeNil())
		Expect(underPrompt.Validate("")).To(BeNil())
		Expect(errors.Is(underPrompt.Validate("not an int"), errInvalidIntInput)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("12.5"), errInvalidIntInput)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("3"), errIntTooLow)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("25"), errIntTooHigh)).To(BeTrue())
	})

	It("can create custom enum prompt", func() {
		prompt, err := Custom{
			Type:        CustomEnum,
			EnumOptions: []string{"a", "b", "c"},
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*enumWrapper)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Select.Items).To(Equal([]string{"a", "b", "c"}))
	})

	It("can run enum prompt", func() {
		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		prompt, err := Custom{
			Type:        CustomEnum,
			EnumOptions: []string{"a", "b", "c"},
		}.CreatePrompt(stdinReader)
		Expect(err).To(BeNil())

		go func() {
			DelayWrite(stdinWriter, []byte{106, 13})
		}()

		out, err := prompt.Run()
		Expect(err).To(BeNil())
		Expect(out).To(Equal("b"))
	})

	It("validates strings", func() {
		var minLength int64 = 4
		var maxLength int64 = 5

		c := Custom{
			Type:      CustomString,
			MinLength: &minLength,
			MaxLength: &maxLength,
		}

		var err error

		err = c.Validate("bad")
		Expect(err).ToNot(BeNil())
		Expect(errors.Unwrap(err)).To(Equal(errInputTooShort))

		err = c.Validate("good")
		Expect(err).To(BeNil())

		err = c.Validate("bad-again")
		Expect(err).ToNot(BeNil())
		Expect(errors.Unwrap(err)).To(Equal(errInputTooLong))
	})

	It("validates ints", func() {
		var minInt int64 = 4
		var maxInt int64 = 5

		c := Custom{
			Type:   CustomInt,
			MinInt: &minInt,
			MaxInt: &maxInt,
		}

		var err error

		err = c.Validate("3")
		Expect(err).ToNot(BeNil())
		Expect(errors.Unwrap(err)).To(Equal(errIntTooLow))

		Expect(c.Validate("4")).To(BeNil())
		Expect(c.Validate("5")).To(BeNil())

		err = c.Validate("6")
		Expect(err).ToNot(BeNil())
		Expect(errors.Unwrap(err)).To(Equal(errIntTooHigh))
	})

	It("validates enums", func() {
		c := Custom{
			Type:        CustomEnum,
			EnumOptions: []string{"good"},
		}

		Expect(c.Validate("good")).To(BeNil())

		err := c.Validate("bad")
		Expect(err).ToNot(BeNil())
		Expect(errors.Unwrap(err)).To(Equal(errInvalidEnum))
	})
})
