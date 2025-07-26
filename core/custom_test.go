package core

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cqroot/prompt/choose"

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

func TestThemeScrollWithFewerThan10Choices(t *testing.T) {
	choices := []choose.Choice{
		{Text: "Option 1"},
		{Text: "Option 2"},
		{Text: "Option 3"},
	}

	result := ThemeScroll(choices, 0) // cursor at index 0

	// Should show all 3 options with the second one selected
	then.True(t, strings.Contains(result, "[•] Option 1"))
	then.True(t, strings.Contains(result, "[ ] Option 2"))
	then.True(t, strings.Contains(result, "[ ] Option 3"))
	then.False(t, strings.Contains(result, "Option 4"))
}

func TestThemeScrollWithExactly10Choices(t *testing.T) {
	choices := make([]choose.Choice, 10)
	for i := range 10 {
		choices[i] = choose.Choice{Text: fmt.Sprintf("Option %d", i+1)}
	}

	result := ThemeScroll(choices, 5) // cursor at middle

	// Should show all 10 options with the 6th one selected
	then.True(t, strings.Contains(result, "[ ] Option 1"))
	then.True(t, strings.Contains(result, "[•] Option 6"))
	then.True(t, strings.Contains(result, "[ ] Option 10"))
}

func TestThemeScrollWithMoreThan10ChoicesCursorAtStart(t *testing.T) {
	choices := make([]choose.Choice, 15)
	for i := 0; i < 15; i++ {
		choices[i] = choose.Choice{Text: fmt.Sprintf("Option %d", i+1)}
	}

	result := ThemeScroll(choices, 2) // cursor at index 2 (less than limit/2)

	// Should show first 10 options with the 3rd one selected
	then.True(t, strings.Contains(result, "[ ] Option 1"))
	then.True(t, strings.Contains(result, "[•] Option 3"))
	then.True(t, strings.Contains(result, "[ ] Option 10"))
	then.False(t, strings.Contains(result, "Option 11"))
}

func TestThemeScrollWithMoreThan10ChoicesCursorInMiddle(t *testing.T) {
	choices := make([]choose.Choice, 15)
	for i := 0; i < 15; i++ {
		choices[i] = choose.Choice{Text: fmt.Sprintf("Option %d", i+1)}
	}

	result := ThemeScroll(choices, 7) // cursor at index 7 (in middle range)

	// Should show options 3-12 with the 8th one selected (index 7 becomes index 5 in display)
	then.True(t, strings.Contains(result, "[ ] Option 3"))
	then.True(t, strings.Contains(result, "[•] Option 8"))
	then.True(t, strings.Contains(result, "[ ] Option 12"))
	// Verify the range is correct by checking that we have exactly 10 options
	lines := strings.Split(strings.TrimSpace(result), "\n")
	then.Equals(t, 10, len(lines))
}

func TestThemeScrollWithMoreThan10ChoicesCursorAtEnd(t *testing.T) {
	choices := make([]choose.Choice, 15)
	for i := 0; i < 15; i++ {
		choices[i] = choose.Choice{Text: fmt.Sprintf("Option %d", i+1)}
	}

	result := ThemeScroll(choices, 13) // cursor at index 13 (near end)

	// Should show last 10 options with the 14th one selected
	then.True(t, strings.Contains(result, "[ ] Option 6"))
	then.True(t, strings.Contains(result, "[•] Option 14"))
	then.True(t, strings.Contains(result, "[ ] Option 15"))
	// Verify the range is correct by checking that we have exactly 10 options
	lines := strings.Split(strings.TrimSpace(result), "\n")
	then.Equals(t, 10, len(lines))
}

func TestThemeScrollWithMoreThan10ChoicesCursorAtVeryEnd(t *testing.T) {
	choices := make([]choose.Choice, 15)
	for i := 0; i < 15; i++ {
		choices[i] = choose.Choice{Text: fmt.Sprintf("Option %d", i+1)}
	}

	result := ThemeScroll(choices, 14) // cursor at last index

	// Should show last 10 options with the 15th one selected
	then.True(t, strings.Contains(result, "[ ] Option 6"))
	then.True(t, strings.Contains(result, "[ ] Option 14"))
	then.True(t, strings.Contains(result, "[•] Option 15"))
	then.False(t, strings.Contains(result, "Option 2"))
	then.False(t, strings.Contains(result, "Option 5"))
}

func TestThemeScrollWithEmptyChoices(t *testing.T) {
	choices := []choose.Choice{}

	result := ThemeScroll(choices, 0)

	// Should return just a newline
	then.Equals(t, "\n", result)
}

func TestThemeScrollWithSingleChoice(t *testing.T) {
	choices := []choose.Choice{
		{Text: "Only Option"},
	}

	result := ThemeScroll(choices, 0)

	// Should show the single option as selected
	then.True(t, strings.EqualFold(strings.TrimSpace(result), "[•] Only Option"))
}

func TestThemeScrollCursorOutOfBounds(t *testing.T) {
	choices := []choose.Choice{
		{Text: "Option 1"},
		{Text: "Option 2"},
		{Text: "Option 3"},
	}

	// Test with cursor beyond array bounds
	result := ThemeScroll(choices, 5)

	// Should handle gracefully and show all options
	then.True(t, strings.Contains(result, "[ ] Option 1"))
	then.True(t, strings.Contains(result, "[ ] Option 2"))
	then.True(t, strings.Contains(result, "[ ] Option 3"))
}

func TestThemeScrollLimitCalculation(t *testing.T) {
	// Test that limit is correctly calculated as min(10, numChoices)
	// Test with 5 choices
	choices5 := make([]choose.Choice, 5)
	for i := 0; i < 5; i++ {
		choices5[i] = choose.Choice{Text: fmt.Sprintf("Option %d", i+1)}
	}

	result5 := ThemeScroll(choices5, 0)
	lines5 := strings.Split(strings.TrimSpace(result5), "\n")
	// Should have 5 lines (one for each choice)
	then.Equals(t, 5, len(lines5))

	// Test with 15 choices
	choices15 := make([]choose.Choice, 15)
	for i := 0; i < 15; i++ {
		choices15[i] = choose.Choice{Text: fmt.Sprintf("Option %d", i+1)}
	}

	result15 := ThemeScroll(choices15, 0)
	lines15 := strings.Split(strings.TrimSpace(result15), "\n")
	// Should have 10 lines (limited by min(10, 15))
	then.Equals(t, 10, len(lines15))
}
