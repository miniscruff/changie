package core

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/manifoldco/promptui"
)

// CustomType determines the possible custom choice types.
// Current values are: `string`, `int` and `enum`.
type CustomType string

const (
	CustomString CustomType = "string"
	CustomInt    CustomType = "int"
	CustomEnum   CustomType = "enum"
)

var (
	errInvalidPromptType = errors.New("invalid prompt type")
	errInvalidIntInput   = errors.New("invalid number")
	errIntTooLow         = errors.New("input below minimum")
	errIntTooHigh        = errors.New("input above maximum")
	errInputTooLong      = errors.New("input length too long")
	errInputTooShort     = errors.New("input length too short")
	base10               = 10
	bit64                = 64
)

type enumWrapper struct {
	*promptui.Select
}

func (e *enumWrapper) Run() (string, error) {
	_, value, err := e.Select.Run()
	return value, err
}

// Prompt is a small wrapper around the promptui Run method
type Prompt interface {
	Run() (string, error)
}

// Custom defines a custom choice that is asked when using 'changie new'.
// The result is an additional custom value in the change file for including in the change line.
//
// A simple one could be the issue number or authors github name.
// example: yaml
// - key: Author
//   label: GitHub Name
//   type: string
//   minLength: 3
type Custom struct {
	// Value used as the key in the custom map for the change format.
	// This should only contain alpha numeric characters, usually starting with a capital.
	// example: yaml
	// key: Issue
	Key string `yaml:"" required:"true"`

	// Specifies the type of choice which changes the prompt.
	//
	// | value | description | options
	// | -- | -- | -- |
	// string | Freeform text option   | [minLength](#custom-minlength) and [maxLength](#custom-maxlength) to limit value
	// int    | Whole numbers          | [minInt](#custom-minint) and [maxInt](#custom-maxint) values to limit value
	// enum   | Limited set of strings | [enumOptions](#custom-enumoptions) is used to specify possible values
	Type CustomType `yaml:"" required:"true"`

	// If true, an empty value will not fail validation.
	// The optional check is handled before min so you can specify that the value is optional but if it is used it must have a minimum length or value depending on type.
	//
	// When building templates that allow for optional values you can compare the custom choice to an empty string to check for a value or empty.
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
	Optional bool `yaml:",omitempty" default:"false"`
	// Description used in the prompt when asking for the choice.
	// If empty key is used instead.
	// example: yaml
	// label: GitHub Username
	Label string `yaml:",omitempty" default:""`
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

func (c Custom) createStringPrompt(stdinReader io.ReadCloser) (Prompt, error) {
	return &promptui.Prompt{
		Label: c.DisplayLabel(),
		Stdin: stdinReader,
		Validate: func(input string) error {
			length := int64(len(input))

			if c.Optional && length == 0 {
				return nil
			}
			if c.MinLength != nil && length < *c.MinLength {
				return fmt.Errorf("%w: length of %v < %v", errInputTooShort, length, c.MinLength)
			}
			if c.MaxLength != nil && length > *c.MaxLength {
				return fmt.Errorf("%w: length of %v > %v", errInputTooLong, length, c.MaxLength)
			}
			return nil
		},
	}, nil
}

func (c Custom) createIntPrompt(stdinReader io.ReadCloser) (Prompt, error) {
	return &promptui.Prompt{
		Label: c.DisplayLabel(),
		Stdin: stdinReader,
		Validate: func(input string) error {
			if c.Optional && input == "" {
				return nil
			}

			value, err := strconv.ParseInt(input, base10, bit64)
			if err != nil {
				return errInvalidIntInput
			}
			if c.MinInt != nil && value < *c.MinInt {
				return fmt.Errorf("%w: %v < %v", errIntTooLow, value, c.MinInt)
			}
			if c.MaxInt != nil && value > *c.MaxInt {
				return fmt.Errorf("%w: %v > %v", errIntTooHigh, value, c.MinInt)
			}
			return nil
		},
	}, nil
}

func (c Custom) createEnumPrompt(stdinReader io.ReadCloser) (Prompt, error) {
	return &enumWrapper{
		Select: &promptui.Select{
			Label: c.DisplayLabel(),
			Stdin: stdinReader,
			Items: c.EnumOptions,
		}}, nil
}

// CreatePrompt will create a promptui select or prompt from a custom choice
func (c Custom) CreatePrompt(stdinReader io.ReadCloser) (Prompt, error) {
	switch c.Type {
	case CustomString:
		return c.createStringPrompt(stdinReader)
	case CustomInt:
		return c.createIntPrompt(stdinReader)
	case CustomEnum:
		return c.createEnumPrompt(stdinReader)
	}

	return nil, errInvalidPromptType
}
