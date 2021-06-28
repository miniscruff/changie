package core

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/manifoldco/promptui"
)

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

// CustomType determines the possible custom choice types
type CustomType string

// Custom contains the options for a custom choice for new changes
type Custom struct {
	Key         string
	Type        CustomType
	Label       string   `yaml:",omitempty"`
	MinInt      *int64   `yaml:"minInt,omitempty"`
	MaxInt      *int64   `yaml:"maxInt,omitempty"`
	EnumOptions []string `yaml:"enumOptions,omitempty"`
}

// CreatePrompt will create a promptui select or prompt from a custom choice
func (c Custom) CreatePrompt(stdinReader io.ReadCloser) (Prompt, error) {
	label := c.Key
	if c.Label != "" {
		label = c.Label
	}

	switch c.Type {
	case CustomString:
		return &promptui.Prompt{
			Label: label,
			Stdin: stdinReader,
		}, nil
	case CustomInt:
		return &promptui.Prompt{
			Label: label,
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
	case CustomEnum:
		return &enumWrapper{
			Select: &promptui.Select{
				Label: label,
				Stdin: stdinReader,
				Items: c.EnumOptions,
			}}, nil
	}

	return nil, ErrInvalidPromptType
}
