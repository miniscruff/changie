package core

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/choose"
	"github.com/cqroot/prompt/constants"
	"github.com/cqroot/prompt/input"
	"github.com/cqroot/prompt/write"

	tea "github.com/charmbracelet/bubbletea"
)

// CustomType determines the possible custom choice types.
// Current values are: `string`, `block`, `int` and `enum`.
type CustomType string

const (
	CustomString CustomType = "string"
	CustomBlock  CustomType = "block"
	CustomInt    CustomType = "int"
	CustomEnum   CustomType = "enum"
)

var (
	errInvalidPromptType   = errors.New("invalid prompt type")
	errInvalidIntInput     = errors.New("invalid number")
	errIntTooLow           = errors.New("input below minimum")
	errIntTooHigh          = errors.New("input above maximum")
	errInputTooLong        = errors.New("input length too long")
	errInputTooShort       = errors.New("input length too short")
	errInvalidEnum         = errors.New("invalid enum")
	errInvalidCustomFormat = errors.New("invalid custom format, must be \"Key=Value\"")
	base10                 = 10
	bit64                  = 64
)

// Custom defines a custom choice that is asked when using 'changie new'.
// The result is an additional custom value in the change file for including in the change line.
//
// A simple one could be the issue number or authors github name.
// example: yaml
//   - key: Author
//     label: GitHub Name
//     type: string
//     minLength: 3
type Custom struct {
	// Value used as the key in the custom map for the change format.
	// This should only contain alpha numeric characters, usually starting with a capital.
	// example: yaml
	// key: Issue
	Key string `yaml:"key" required:"true"`

	// Specifies the type of choice which changes the prompt.
	//
	// | value | description | options
	// | -- | -- | -- |
	// string | Freeform text | [minLength](#custom-minlength) and [maxLength](#custom-maxlength)
	// block | Multiline text | [minLength](#custom-minlength) and [maxLength](#custom-maxlength)
	// int | Whole numbers | [minInt](#custom-minint) and [maxInt](#custom-maxint)
	// enum | Limited set of strings | [enumOptions](#custom-enumoptions) is used to specify values
	Type CustomType `yaml:"type" required:"true"`

	// If true, an empty value will not fail validation.
	// The optional check is handled before min so you can specify that the value is optional but if it
	// is used it must have a minimum length or value depending on type.
	//
	// When building templates that allow for optional values you can compare the custom choice to an
	// empty string to check for a value or empty.
	//
	// example: yaml
	// custom:
	// - key: TicketNumber
	//   type: int
	//   optional: true
	// changeFormat: >-
	// {{- if not (eq .Custom.TicketNumber "")}}
	// PROJ-{{.Custom.TicketNumber}}
	// {{- end}}
	// {{.Body}}
	Optional bool `yaml:"optional,omitempty" default:"false"`
	// Description used in the prompt when asking for the choice.
	// If empty key is used instead.
	// example: yaml
	// label: GitHub Username
	Label string `yaml:"label,omitempty" default:""`
	// If specified the input value must be greater than or equal to minInt.
	MinInt *int64 `yaml:"minInt,omitempty" default:"nil"`
	// If specified the input value must be less than or equal to maxInt.
	MaxInt *int64 `yaml:"maxInt,omitempty" default:"nil"`
	// If specified the string input must be at least this long
	MinLength *int64 `yaml:"minLength,omitempty" default:"nil"`
	// If specified string input must be no more than this long
	MaxLength *int64 `yaml:"maxLength,omitempty" default:"nil"`
	// When using the enum type, you must also specify what possible options to allow.
	// Users will be given a selection list to select the value they want.
	EnumOptions []string `yaml:"enumOptions,omitempty"`
}

func (c Custom) DisplayLabel() string {
	if c.Label == "" {
		return c.Key
	}

	return c.Label
}

func (c Custom) askString(stdinReader io.Reader) (string, error) {
	return prompt.New().Ask(c.DisplayLabel()).
		Input(
			"",
			input.WithHelp(true),
			input.WithValidateFunc(c.validateString),
			input.WithTeaProgramOpts(tea.WithInput(stdinReader)),
		)
}

func (c Custom) askBlock(stdinReader io.Reader) (string, error) {
	return prompt.New().Ask(c.DisplayLabel()).
		Write(
			"",
			write.WithHelp(true),
			write.WithValidateFunc(c.validateString),
			write.WithTeaProgramOpts(tea.WithInput(stdinReader)),
		)
}

func (c Custom) askInt(stdinReader io.Reader) (string, error) {
	return prompt.New().Ask(c.DisplayLabel()).
		Input(
			"",
			input.WithHelp(true),
			input.WithInputMode(input.InputInteger),
			input.WithValidateFunc(c.validateInt),
			input.WithTeaProgramOpts(tea.WithInput(stdinReader)),
		)
}

// Scrolling theme for cqroot/prompt. Allows items to be scrolled through
// using the arrow keys.
//
// choices is a list of enum options to display.
// cursor is the index of the currently selected item.
func ThemeScroll(choices []choose.Choice, cursor int) string {
	s := strings.Builder{}
	s.WriteString("\n")

	numChoices := len(choices)
	index := cursor
	limit := min(10, numChoices)
	start := 0
	end := limit

	// Create a slice of items to display
	if numChoices > limit {
		if index < limit/2 {
			end = limit
		} else if index >= limit/2 && index <= numChoices-limit/2 {
			start = index - limit/2
			end = index + limit/2
			index = limit / 2
		} else {
			start = numChoices - limit
			end = numChoices
			index = limit - (numChoices - index)
		}
	}

	// Loop through the slice and display each item.
	// We will determine if the item is selected or not by calling isSelected
	// with the index of the item, but since we're only displaying a slice of
	// the items we need to offset the index.
	for i, choice := range choices[start:end] {
		// Check if the item is the one at the cursor's location
		isCursor := index == i

		if isCursor {
			s.WriteString(constants.DefaultSelectedItemStyle.Render(fmt.Sprintf("• %s", choice.Text)))
		} else {
			s.WriteString(constants.DefaultItemStyle.Render(fmt.Sprintf("  %s", choice.Text)))
		}

		s.WriteString("\n")
	}

	return s.String()
}

func (c Custom) askEnum(stdinReader io.Reader) (string, error) {
	return prompt.New().Ask(c.DisplayLabel()).
		Choose(
			c.EnumOptions,
			choose.WithHelp(true),
			choose.WithTeaProgramOpts(tea.WithInput(stdinReader)),
			choose.WithTheme(ThemeScroll),
		)
}

// CreatePrompt will create a promptui select or prompt from a custom choice
func (c Custom) AskPrompt(stdinReader io.Reader) (string, error) {
	switch c.Type {
	case CustomString:
		return c.askString(stdinReader)
	case CustomBlock:
		return c.askBlock(stdinReader)
	case CustomInt:
		return c.askInt(stdinReader)
	case CustomEnum:
		return c.askEnum(stdinReader)
	}

	return "", errInvalidPromptType
}

func (c Custom) Validate(input string) error {
	switch c.Type {
	case CustomString, CustomBlock:
		return c.validateString(input)
	case CustomInt:
		return c.validateInt(input)
	case CustomEnum:
		return c.validateEnum(input)
	}

	return errInvalidPromptType
}

func (c Custom) validateString(input string) error {
	length := int64(len(input))

	if c.Optional && length == 0 {
		return nil
	}

	if c.MinLength != nil && length < *c.MinLength {
		return fmt.Errorf("%w: length of %v < %v", errInputTooShort, length, *c.MinLength)
	}

	if c.MaxLength != nil && length > *c.MaxLength {
		return fmt.Errorf("%w: length of %v > %v", errInputTooLong, length, *c.MaxLength)
	}

	return nil
}

func (c Custom) validateInt(input string) error {
	if c.Optional && input == "" {
		return nil
	}

	value, err := strconv.ParseInt(input, base10, bit64)
	if err != nil {
		return errInvalidIntInput
	}

	if c.MinInt != nil && value < *c.MinInt {
		return fmt.Errorf("%w: %v < %v", errIntTooLow, value, *c.MinInt)
	}

	if c.MaxInt != nil && value > *c.MaxInt {
		return fmt.Errorf("%w: %v > %v", errIntTooHigh, value, *c.MinInt)
	}

	return nil
}

func (c Custom) validateEnum(input string) error {
	if slices.Contains(c.EnumOptions, input) {
		return nil
	}

	return fmt.Errorf("%w: %s", errInvalidEnum, input)
}

// CustomMapFromStrings will parse a CLI argument of strings into a key value map
// where each string is a key value pair separated by an equal sign.
// Eg: Issue=15 turns into {"Issue": "15"}
func CustomMapFromStrings(input []string) (map[string]string, error) {
	ret := make(map[string]string)

	for _, s := range input {
		key, value, found := strings.Cut(s, "=")
		if !found {
			// error
			return ret, fmt.Errorf("%w: %s", errInvalidCustomFormat, s)
		}

		ret[key] = value
	}

	return ret, nil
}
