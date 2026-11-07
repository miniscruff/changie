package core

import (
	"testing"

	"charm.land/huh/v2"

	tea "charm.land/bubbletea/v2"

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

func TestErrorInvalidPromptTypeWhenValidating(t *testing.T) {
	custom := Custom{Type: "invalid-type"}
	err := custom.Validate("ignored")
	then.Err(t, errInvalidPromptType, err)
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

func TestValidEnums(t *testing.T) {
	custom := Custom{
		Type:        CustomEnums,
		EnumOptions: []string{"north", "south", "east", "west"},
	}
	err := custom.Validate("south, west")
	then.Nil(t, err)
}

func TestInvalidEnums(t *testing.T) {
	custom := Custom{
		Type:        CustomEnums,
		EnumOptions: []string{"north", "south", "east", "west"},
	}
	err := custom.Validate("south, north-south")
	then.Err(t, errInvalidEnum, err)
}

func TestOptionalEmptyEnums(t *testing.T) {
	custom := Custom{
		Type:        CustomEnums,
		Optional:    true,
		EnumOptions: []string{"north", "south", "east", "west"},
	}
	err := custom.Validate("")
	then.Nil(t, err)
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

// runCustomField runs a form with only the custom field, driving it with input.
func runCustomField(t *testing.T, custom Custom, input func(d *formDriver)) (*formDriver, string) {
	t.Helper()

	var value string

	field, err := custom.formField(&value)
	then.Nil(t, err)

	d := newFormDriver(t, huh.NewForm(huh.NewGroup(field)))
	input(d)

	return d, value
}

func TestCustomFormFieldString(t *testing.T) {
	custom := Custom{Key: "Author", Label: "GitHub Name", Type: CustomString}

	d, value := runCustomField(t, custom, func(d *formDriver) {
		then.Contains(t, "GitHub Name", d.view())
		d.typeText("miniscruff")
		d.enter()
	})

	then.True(t, d.completed())
	then.Equals(t, "miniscruff", value)
}

func TestCustomFormFieldStringLength(t *testing.T) {
	maxLength := int64(3)
	custom := Custom{Key: "Short", Type: CustomString, MaxLength: &maxLength}

	d, _ := runCustomField(t, custom, func(d *formDriver) {
		d.typeText("abcd")
		d.enter()
	})

	then.False(t, d.completed())
	then.Contains(t, errInputTooLong.Error(), d.view())
}

func TestCustomFormFieldBlock(t *testing.T) {
	custom := Custom{Key: "Notes", Type: CustomBlock}

	d, value := runCustomField(t, custom, func(d *formDriver) {
		d.typeText("line one")
		d.send(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt})
		d.typeText("line two")
		d.enter()
	})

	then.True(t, d.completed())
	then.Equals(t, "line one\nline two", value)
}

func TestCustomFormFieldInt(t *testing.T) {
	custom := Custom{Key: "Issue", Type: CustomInt}

	d, value := runCustomField(t, custom, func(d *formDriver) {
		d.typeText("15")
		d.enter()
	})

	then.True(t, d.completed())
	then.Equals(t, "15", value)
}

func TestCustomFormFieldIntInvalid(t *testing.T) {
	minInt := int64(10)
	custom := Custom{Key: "Issue", Type: CustomInt, MinInt: &minInt}

	d, _ := runCustomField(t, custom, func(d *formDriver) {
		d.typeText("abc")
		d.enter()
	})

	then.False(t, d.completed())
	then.Contains(t, errInvalidIntInput.Error(), d.view())

	d, _ = runCustomField(t, custom, func(d *formDriver) {
		d.typeText("5")
		d.enter()
	})

	then.False(t, d.completed())
	then.Contains(t, errIntTooLow.Error(), d.view())
}

func TestCustomFormFieldOptionalInt(t *testing.T) {
	custom := Custom{Key: "Issue", Type: CustomInt, Optional: true}

	d, value := runCustomField(t, custom, func(d *formDriver) {
		d.enter()
	})

	then.True(t, d.completed())
	then.Equals(t, "", value)
}

func TestCustomFormFieldEnum(t *testing.T) {
	custom := Custom{Key: "Platform", Type: CustomEnum, EnumOptions: []string{"linux", "mac", "windows"}}

	d, value := runCustomField(t, custom, func(d *formDriver) {
		d.down()
		d.enter()
	})

	then.True(t, d.completed())
	then.Equals(t, "mac", value)
}

func TestCustomFormFieldEnums(t *testing.T) {
	custom := Custom{Key: "Platforms", Type: CustomEnums, EnumOptions: []string{"linux", "mac", "windows"}}

	d, value := runCustomField(t, custom, func(d *formDriver) {
		d.press(' ')
		d.down()
		d.down()
		d.press(' ')
		d.enter()
	})

	then.True(t, d.completed())
	then.Equals(t, "linux, windows", value)
	then.Nil(t, custom.Validate(value))
}

func TestCustomFormFieldEnumsRequired(t *testing.T) {
	custom := Custom{Key: "Platforms", Type: CustomEnums, EnumOptions: []string{"linux", "mac"}}

	d, _ := runCustomField(t, custom, func(d *formDriver) {
		d.enter()
	})

	then.False(t, d.completed())
	then.Contains(t, errEnumsRequired.Error(), d.view())
}

func TestCustomFormFieldOptionalEnums(t *testing.T) {
	custom := Custom{Key: "Platforms", Type: CustomEnums, Optional: true, EnumOptions: []string{"linux", "mac"}}

	d, value := runCustomField(t, custom, func(d *formDriver) {
		d.enter()
	})

	then.True(t, d.completed())
	then.Equals(t, "", value)
}

func TestCustomFormFieldInvalidType(t *testing.T) {
	var value string

	_, err := Custom{Key: "Bad", Type: "bad"}.formField(&value)
	then.Err(t, errInvalidPromptType, err)
}
