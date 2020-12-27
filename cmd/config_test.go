package cmd

import (
	"errors"
	"github.com/manifoldco/promptui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("Config", func() {
	It("should save the config", func() {
		config := Config{
			ChangesDir:    "Changes",
			UnreleasedDir: "Unrel",
			HeaderPath:    "header.tpl.md",
			ChangelogPath: "CHANGELOG.md",
		}

		configYaml := `changesDir: Changes
unreleasedDir: Unrel
headerPath: header.tpl.md
changelogPath: CHANGELOG.md
versionExt: ""
versionFormat: ""
kindFormat: ""
changeFormat: ""
kinds: []
`

		writeCalled := false

		mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
			writeCalled = true
			Expect(filepath).To(Equal(configPath))
			Expect(string(bytes)).To(Equal(configYaml))
			return nil
		}

		err := config.Save(mockWf)
		Expect(err).To(BeNil())
		Expect(writeCalled).To(Equal(true))
	})

	It("should save the config with custom choices", func() {
		config := Config{
			ChangesDir:    "Changes",
			UnreleasedDir: "Unrel",
			HeaderPath:    "header.tpl.md",
			ChangelogPath: "CHANGELOG.md",
			CustomChoices: map[string]Custom{
				"first": Custom{
					Type:  CustomString,
					Label: "First name",
				},
			},
		}

		configYaml := `changesDir: Changes
unreleasedDir: Unrel
headerPath: header.tpl.md
changelogPath: CHANGELOG.md
versionExt: ""
versionFormat: ""
kindFormat: ""
changeFormat: ""
kinds: []
custom:
  first:
    type: string
    label: First name
`

		writeCalled := false

		mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
			writeCalled = true
			Expect(filepath).To(Equal(configPath))
			Expect(string(bytes)).To(Equal(configYaml))
			return nil
		}

		err := config.Save(mockWf)
		Expect(err).To(BeNil())
		Expect(writeCalled).To(Equal(true))
	})

	It("should load change from path", func() {
		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal(configPath))
			return []byte("changesDir: C\nheaderPath: header.rst\n"), nil
		}

		config, err := LoadConfig(mockRf)
		Expect(err).To(BeNil())
		Expect(config.ChangesDir).To(Equal("C"))
		Expect(config.HeaderPath).To(Equal("header.rst"))
	})

	It("should return error from bad read", func() {
		mockErr := errors.New("bad file")

		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal(configPath))
			return []byte(""), mockErr
		}

		_, err := LoadConfig(mockRf)
		Expect(err).To(Equal(mockErr))
	})

	It("should return error from bad file", func() {
		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal(configPath))
			return []byte("not a yaml file---"), nil
		}

		_, err := LoadConfig(mockRf)
		Expect(err).NotTo(BeNil())
	})
})

var _ = Describe("Custom", func() {
	It("returns error on invalid prompt type", func() {
		_, err := Custom{
			Type: "invalid type",
		}.CreatePrompt("name", os.Stdin)
		Expect(errors.Is(err, invalidPromptType)).To(BeTrue())
	})

	It("can create custom string prompt", func() {
		prompt, err := Custom{
			Type:  CustomString,
			Label: "a label",
		}.CreatePrompt("name", os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Label).To(Equal("a label"))
	})

	It("can create custom int prompt", func() {
		prompt, err := Custom{
			Type: CustomInt,
		}.CreatePrompt("name", os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Validate("12")).To(BeNil())
		Expect(underPrompt.Validate("not an int")).NotTo(BeNil())
		Expect(underPrompt.Validate("12.5")).NotTo(BeNil())
		Expect(underPrompt.Label).To(Equal("name"))
	})

	It("can create custom int prompt with a min and max", func() {
		var min int64 = 5
		var max int64 = 10
		prompt, err := Custom{
			Type:   CustomInt,
			MinInt: &min,
			MaxInt: &max,
		}.CreatePrompt("name", os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Validate("7")).To(BeNil())
		Expect(errors.Is(underPrompt.Validate("3"), intTooLow)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("25"), intTooHigh)).To(BeTrue())
	})

	It("can create custom enum prompt", func() {
		prompt, err := Custom{
			Type:        CustomEnum,
			EnumOptions: []string{"a", "b", "c"},
		}.CreatePrompt("name", os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*EnumWrapper)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Select.Items).To(Equal([]string{"a", "b", "c"}))
	})
})
