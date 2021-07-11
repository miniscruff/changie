package core

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/manifoldco/promptui"
)

// CustomType determines the possible custom choice types
type CustomType string

const (
	CustomString CustomType = "string"
	CustomInt    CustomType = "int"
	CustomEnum   CustomType = "enum"
)

var (
	ErrInvalidPromptType = errors.New("invalid prompt type")
	errInvalidIntInput   = errors.New("invalid number")
	errIntTooLow         = errors.New("input below minimum")
	errIntTooHigh        = errors.New("input above maximum")
	errInputTooLong      = errors.New("input length too long")
	errInputTooShort     = errors.New("input length too short")
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

// Custom contains the options for a custom choice for new changes
type Custom struct {
	Key         string
	Type        CustomType
	Label       string   `yaml:",omitempty"`
	MinInt      *int64   `yaml:"minInt,omitempty"`
	MaxInt      *int64   `yaml:"maxInt,omitempty"`
	MinLength   *int64   `yaml:"minLength,omitempty"`
	MaxLength   *int64   `yaml:"maxLength,omitempty"`
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
			value, err := strconv.ParseInt(input, 10, 64)
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

	return nil, ErrInvalidPromptType
}
