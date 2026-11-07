package core

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"charm.land/huh/v2"
)

// CustomType determines the possible custom choice types.
// Current values are: `string`, `block`, `int`, `enum` and `enums`.
type CustomType string

const (
	CustomString CustomType = "string"
	CustomBlock  CustomType = "block"
	CustomInt    CustomType = "int"
	CustomEnum   CustomType = "enum"
	CustomEnums  CustomType = "enums"
)

// enumsSeparator joins and splits the selected values of an "enums" custom
// choice, as the selections are stored as a single string.
const enumsSeparator = ", "

var (
	errInvalidPromptType   = errors.New("invalid prompt type")
	errInvalidIntInput     = errors.New("invalid number")
	errIntTooLow           = errors.New("input below minimum")
	errIntTooHigh          = errors.New("input above maximum")
	errInputTooLong        = errors.New("input length too long")
	errInputTooShort       = errors.New("input length too short")
	errInvalidEnum         = errors.New("invalid enum")
	errEnumsRequired       = errors.New("at least one option is required")
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
	// enums | Multiple values from a limited set of strings | [enumOptions](#custom-enumoptions) is used to specify values
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
	// When using the enum or enums type, you must also specify what possible options to allow.
	// Users will be given a selection list to select the value, or values, they want.
	EnumOptions []string `yaml:"enumOptions,omitempty"`
}

func (c Custom) DisplayLabel() string {
	if c.Label == "" {
		return c.Key
	}

	return c.Label
}

// formField creates a huh field for the custom choice, storing the answer in
// value. Multiple enum selections are stored joined.
func (c Custom) formField(value *string) (huh.Field, error) {
	switch c.Type {
	case CustomString:
		return huh.NewInput().
			Key(c.Key).
			Title(c.DisplayLabel()).
			Validate(c.validateString).
			Value(value), nil
	case CustomBlock:
		return huh.NewText().
			Key(c.Key).
			Title(c.DisplayLabel()).
			Validate(c.validateString).
			Value(value), nil
	case CustomInt:
		return huh.NewInput().
			Key(c.Key).
			Title(c.DisplayLabel()).
			Validate(c.validateInt).
			Value(value), nil
	case CustomEnum:
		return huh.NewSelect[string]().
			Key(c.Key).
			Title(c.DisplayLabel()).
			Options(huh.NewOptions(c.EnumOptions...)...).
			Height(min(len(c.EnumOptions)+1, maxOptionsHeight)).
			Value(value), nil
	case CustomEnums:
		return huh.NewMultiSelect[string]().
			Key(c.Key).
			Title(c.DisplayLabel()).
			Options(huh.NewOptions(c.EnumOptions...)...).
			Height(min(len(c.EnumOptions)+1, maxOptionsHeight)).
			Validate(func(values []string) error {
				if len(values) == 0 && !c.Optional {
					return errEnumsRequired
				}

				return nil
			}).
			Accessor(joinedAccessor{value: value}), nil
	}

	return nil, fmt.Errorf("%w: %s", errInvalidPromptType, c.Type)
}

// joinedAccessor stores multiple selected values as a single joined string.
type joinedAccessor struct {
	value *string
}

func (a joinedAccessor) Get() []string {
	if *a.value == "" {
		return []string{}
	}

	return strings.Split(*a.value, enumsSeparator)
}

func (a joinedAccessor) Set(values []string) {
	*a.value = strings.Join(values, enumsSeparator)
}

func (c Custom) Validate(input string) error {
	switch c.Type {
	case CustomString, CustomBlock:
		return c.validateString(input)
	case CustomInt:
		return c.validateInt(input)
	case CustomEnum:
		return c.validateEnum(input)
	case CustomEnums:
		return c.validateEnums(input)
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

func (c Custom) validateEnums(input string) error {
	if c.Optional && input == "" {
		return nil
	}

	for _, value := range strings.Split(input, enumsSeparator) {
		if !slices.Contains(c.EnumOptions, value) {
			return fmt.Errorf("%w: %s", errInvalidEnum, value)
		}
	}

	return nil
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
