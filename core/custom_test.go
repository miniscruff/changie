package core

import (
	"os"
	"testing"

	"github.com/miniscruff/changie/then"
)

type ValidationSpecs struct {
	Name   string
	Custom Custom
	Specs  []ValidationSpec
}

func (s *ValidationSpecs) Run(t *testing.T) {
	for _, spec := range s.Specs {
		t.Run(s.Name+spec.Name, func(t *testing.T) {
			spec.Run(t, s.Custom)
		})
	}
}

type ValidationSpec struct {
	Name  string
	Input string
	Error error
}

func (s *ValidationSpec) Run(t *testing.T, custom Custom) {
	err := custom.Validate(s.Input)

	if s.Error == nil {
		then.Nil(t, err)
	} else {
		then.Err(t, s.Error, err)
	}
}

func PtrInt64(value int64) *int64 {
	return &value
}

func TestDisplayLabelDefaultsToKey(t *testing.T) {
	custom := Custom{Key: "key"}
	then.Equals(t, "key", custom.DisplayLabel())
}

func TestDisplayLabelUsesLabelIfDefined(t *testing.T) {
	custom := Custom{Key: "key", Label: "Lab"}
	then.Equals(t, "Lab", custom.DisplayLabel())
}

func TestErrorInvalidPromptType(t *testing.T) {
	custom := Custom{Type: "invalid-type"}
	_, err := custom.AskPrompt(os.Stdin)
	then.Err(t, errInvalidPromptType, err)
}

func TestErrorCustomStringMultiChoosePrompt(t *testing.T) {
	custom := Custom{Type: CustomString}
	_, err := custom.AskMultiChoosePrompt(os.Stdin)
	then.Err(t, errInvalidPromptType, err)
}

func TestErrorCustomBlockMultiChoosePrompt(t *testing.T) {
	custom := Custom{Type: CustomBlock}
	_, err := custom.AskMultiChoosePrompt(os.Stdin)
	then.Err(t, errInvalidPromptType, err)
}

func TestErrorCustomIntMultiChoosePrompt(t *testing.T) {
	custom := Custom{Type: CustomInt}
	_, err := custom.AskMultiChoosePrompt(os.Stdin)
	then.Err(t, errInvalidPromptType, err)
}

func TestErrorInvalidMultiChoosePromptType(t *testing.T) {
	custom := Custom{Type: "invalid-type"}
	_, err := custom.AskMultiChoosePrompt(os.Stdin)
	then.Err(t, errInvalidPromptType, err)
}

func TestErrorInvalidPromptTypeWhenValidating(t *testing.T) {
	custom := Custom{Type: "invalid-type"}
	err := custom.Validate("ignored")
	then.Err(t, errInvalidPromptType, err)
}

func TestCreateCustomIntPrompt(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("15"),
		[]byte{13}, // 13=enter
	)

	custom := Custom{Type: CustomInt, Key: "name"}
	value, err := custom.AskPrompt(reader)

	then.Nil(t, err)
	then.Equals(t, value, "15")
}

func TestCanRunBlockPrompt(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte("I can write multiple lines"),
		[]byte{13},
		[]byte("In a block prompt"),
		[]byte{4}, // 4=EOT or ctrl+d
	)

	custom := Custom{Type: CustomBlock, Key: "block"}
	value, err := custom.AskPrompt(reader)

	then.Nil(t, err)
	then.Equals(t, value, "I can write multiple lines\nIn a block prompt")
}

func TestCanRunEnumPrompt(t *testing.T) {
	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{106, 13}, // 106 = down, 13 = enter
	)

	opts := []string{"a", "b", "c"}
	custom := Custom{Type: CustomEnum, EnumOptions: opts}

	value, err := custom.AskPrompt(reader)
	then.Nil(t, err)
	then.Equals(t, "b", value)
}

var validationSpecs = []ValidationSpecs{
	{
		Name: "String",
		Custom: Custom{
			Type:      CustomString,
			Key:       "string",
			MinLength: PtrInt64(3),
			MaxLength: PtrInt64(15),
		},
		Specs: []ValidationSpec{
			{
				Name:  "Empty",
				Input: "",
				Error: errInputTooShort,
			},
			{
				Name:  "TooShort",
				Input: "0",
				Error: errInputTooShort,
			},
			{
				Name:  "TooLong",
				Input: "01234567890123456789",
				Error: errInputTooLong,
			},
			{
				Name:  "Valid",
				Input: "normal",
			},
		},
	},
	{
		Name: "OptionalString",
		Custom: Custom{
			Type:      CustomString,
			Label:     "OptString",
			Optional:  true,
			MinLength: PtrInt64(3),
			MaxLength: PtrInt64(15),
		},
		Specs: []ValidationSpec{
			{
				Name: "Empty",
			},
			{
				Name:  "TooShort",
				Input: "0",
				Error: errInputTooShort,
			},
			{
				Name:  "TooLong",
				Input: "01234567890123456789",
				Error: errInputTooLong,
			},
			{
				Name:  "Valid",
				Input: "normal",
			},
		},
	},
	{
		Name: "Block",
		Custom: Custom{
			Type:      CustomBlock,
			Key:       "string",
			MinLength: PtrInt64(3),
			MaxLength: PtrInt64(15),
		},
		Specs: []ValidationSpec{
			{
				Name:  "Empty",
				Input: "",
				Error: errInputTooShort,
			},
			{
				Name:  "TooShort",
				Input: "0",
				Error: errInputTooShort,
			},
			{
				Name:  "TooLong",
				Input: "01234567890123456789",
				Error: errInputTooLong,
			},
			{
				Name:  "Valid",
				Input: "normal",
			},
		},
	},
	{
		Name: "OptionalBlock",
		Custom: Custom{
			Type:      CustomBlock,
			Label:     "OptString",
			Optional:  true,
			MinLength: PtrInt64(3),
			MaxLength: PtrInt64(15),
		},
		Specs: []ValidationSpec{
			{
				Name: "Empty",
			},
			{
				Name:  "TooShort",
				Input: "0",
				Error: errInputTooShort,
			},
			{
				Name:  "TooLong",
				Input: "01234567890123456789",
				Error: errInputTooLong,
			},
			{
				Name:  "Valid",
				Input: "normal",
			},
		},
	},
	{
		Name: "Int",
		Custom: Custom{
			Type:   CustomInt,
			Label:  "Int",
			MinInt: PtrInt64(10),
			MaxInt: PtrInt64(100),
		},
		Specs: []ValidationSpec{
			{
				Name:  "Empty",
				Input: "",
				Error: errInvalidIntInput,
			},
			{
				Name:  "Text",
				Input: "random text",
				Error: errInvalidIntInput,
			},
			{
				Name:  "Float Value",
				Input: "13.66",
				Error: errInvalidIntInput,
			},
			{
				Name:  "TooLow",
				Input: "5",
				Error: errIntTooLow,
			},
			{
				Name:  "TooHigh",
				Input: "500",
				Error: errIntTooHigh,
			},
			{
				Name:  "Valid",
				Input: "52",
			},
		},
	},
	{
		Name: "OptionalInt",
		Custom: Custom{
			Type:     CustomInt,
			Label:    "Int",
			Optional: true,
			MinInt:   PtrInt64(10),
			MaxInt:   PtrInt64(100),
		},
		Specs: []ValidationSpec{
			{
				Name:  "Empty",
				Input: "",
			},
			{
				Name:  "Text",
				Input: "random text",
				Error: errInvalidIntInput,
			},
			{
				Name:  "Float Value",
				Input: "13.66",
				Error: errInvalidIntInput,
			},
			{
				Name:  "TooLow",
				Input: "5",
				Error: errIntTooLow,
			},
			{
				Name:  "TooHigh",
				Input: "500",
				Error: errIntTooHigh,
			},
			{
				Name:  "Valid",
				Input: "52",
			},
		},
	},
}

func TestValidators(t *testing.T) {
	for _, specs := range validationSpecs {
		specs.Run(t)
	}
}

func TestValidEnum(t *testing.T) {
	custom := Custom{
		Type:        CustomEnum,
		EnumOptions: []string{"north", "south", "east", "west"},
	}
	err := custom.Validate("south")
	then.Nil(t, err)
}

func TestInvalidEnum(t *testing.T) {
	custom := Custom{
		Type:        CustomEnum,
		EnumOptions: []string{"north", "south", "east", "west"},
	}
	err := custom.Validate("north-south")
	then.Err(t, errInvalidEnum, err)
}

func TestCustomMapFromStrings(t *testing.T) {
	expected := map[string]string{
		"Issue": "15",
		"Tag":   "alpha",
		"Owner": "team-name",
	}

	inputs := []string{"Issue=15", "Tag=alpha", "Owner=team-name"}
	customValues, err := CustomMapFromStrings(inputs)
	then.Nil(t, err)
	then.MapEquals(t, expected, customValues)
}

func TestErrorBadMapFormat(t *testing.T) {
	inputs := []string{"Issue=15", "Tag=alpha", "Ownerteam-name"}
	_, err := CustomMapFromStrings(inputs)
	then.Err(t, errInvalidCustomFormat, err)
}
