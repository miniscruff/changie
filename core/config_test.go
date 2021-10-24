package core

import (
	"errors"
	"os"

	"github.com/manifoldco/promptui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	It("should save the config", func() {
		config := Config{
			ChangesDir:        "Changes",
			UnreleasedDir:     "Unrel",
			HeaderPath:        "header.tpl.md",
			VersionHeaderPath: "header.md",
			ChangelogPath:     "CHANGELOG.md",
		}

		configYaml := `changesDir: Changes
unreleasedDir: Unrel
headerPath: header.tpl.md
versionHeaderPath: header.md
changelogPath: CHANGELOG.md
versionExt: ""
versionFormat: ""
changeFormat: ""
`

		writeCalled := false

		mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
			writeCalled = true
			Expect(filepath).To(Equal(ConfigPaths[0]))
			Expect(string(bytes)).To(Equal(configYaml))
			return nil
		}

		err := config.Save(mockWf)
		Expect(err).To(BeNil())
		Expect(writeCalled).To(Equal(true))
	})

	It("should save the config with custom choices and optionals", func() {
		config := Config{
			ChangesDir:      "Changes",
			UnreleasedDir:   "Unrel",
			HeaderPath:      "header.tpl.md",
			ChangelogPath:   "CHANGELOG.md",
			VersionFormat:   "vers",
			ComponentFormat: "comp",
			KindFormat:      "kind",
			ChangeFormat:    "chng",
			Components:      []string{"A", "D", "G"},
			Kinds: []KindConfig{
				{Label: "B"},
				{Label: "C"},
				{Label: "E"},
			},
			CustomChoices: []Custom{
				{
					Key:   "first",
					Type:  CustomString,
					Label: "First name",
				},
			},
		}

		configYaml := `changesDir: Changes
unreleasedDir: Unrel
headerPath: header.tpl.md
versionHeaderPath: ""
changelogPath: CHANGELOG.md
versionExt: ""
versionFormat: vers
componentFormat: comp
kindFormat: kind
changeFormat: chng
components:
- A
- D
- G
kinds:
- label: B
- label: C
- label: E
custom:
- key: first
  type: string
  label: First name
`

		writeCalled := false

		mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
			writeCalled = true
			Expect(filepath).To(Equal(ConfigPaths[0]))
			Expect(string(bytes)).To(Equal(configYaml))
			return nil
		}

		err := config.Save(mockWf)
		Expect(err).To(BeNil())
		Expect(writeCalled).To(Equal(true))
	})

	It("should load config from path", func() {
		mockRf := func(filepath string) ([]byte, error) {
			if filepath == ConfigPaths[0] {
				return []byte("changesDir: C\nheaderPath: header.rst\n"), nil
			}
			return nil, os.ErrNotExist
		}

		config, err := LoadConfig(mockRf)
		Expect(err).To(BeNil())
		Expect(config.ChangesDir).To(Equal("C"))
		Expect(config.HeaderPath).To(Equal("header.rst"))
	})

	It("should load config from alternate path", func() {
		mockRf := func(filepath string) ([]byte, error) {
			if filepath == ConfigPaths[1] {
				return []byte("changesDir: C\nheaderPath: header.rst\n"), nil
			}
			return nil, os.ErrNotExist
		}

		config, err := LoadConfig(mockRf)
		Expect(err).To(BeNil())
		Expect(config.ChangesDir).To(Equal("C"))
		Expect(config.HeaderPath).To(Equal("header.rst"))
	})

	It("should load config from custom path", func() {
		mockRf := func(filepath string) ([]byte, error) {
			if filepath == "./custom/changie.yaml" {
				return []byte("changesDir: C\nheaderPath: header.rst\n"), nil
			}
			return nil, os.ErrNotExist
		}

		os.Setenv("CHANGIE_CONFIG_PATH", "./custom/changie.yaml")
		config, err := LoadConfig(mockRf)
		Expect(err).To(BeNil())
		Expect(config.ChangesDir).To(Equal("C"))
		Expect(config.HeaderPath).To(Equal("header.rst"))
		os.Setenv("CHANGIE_CONFIG_PATH", "")
	})

	It("should return error from bad read", func() {
		mockErr := errors.New("bad file")

		mockRf := func(filepath string) ([]byte, error) {
			return []byte(""), mockErr
		}

		_, err := LoadConfig(mockRf)
		Expect(err).To(Equal(mockErr))
	})

	It("should return error from bad file", func() {
		mockRf := func(filepath string) ([]byte, error) {
			return []byte("not a yaml file---"), nil
		}

		_, err := LoadConfig(mockRf)
		Expect(err).NotTo(BeNil())
	})
})

var _ = Describe("Body Config", func() {
	It("can create prompt when config empty", func() {
		p := BodyConfig{}.CreatePrompt(os.Stdin)
		underPrompt, ok := p.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Label).To(Equal("Body"))
		Expect(underPrompt.Validate("anything")).To(BeNil())
	})

	It("can create prompt with min and max", func() {
		var max int64 = 10
		longInput := "jas dklfjaklsd fjklasjd flkasjdfkl sd"
		p := BodyConfig{MaxLength: &max}.CreatePrompt(os.Stdin)
		underPrompt, ok := p.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Label).To(Equal("Body"))
		Expect(errors.Is(underPrompt.Validate(longInput), errInputTooLong)).To(BeTrue())
	})
})
