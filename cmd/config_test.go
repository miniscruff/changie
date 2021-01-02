package cmd

import (
	"errors"
	"os"

	"github.com/manifoldco/promptui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
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
			CustomChoices: []Custom{
				{
					Key:   "first",
					Type:  customString,
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
- key: first
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
		}.CreatePrompt(os.Stdin)
		Expect(errors.Is(err, errInvalidPromptType)).To(BeTrue())
	})

	It("can create custom string prompt", func() {
		prompt, err := Custom{
			Type:  customString,
			Label: "a label",
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Label).To(Equal("a label"))
	})

	It("can create custom int prompt", func() {
		prompt, err := Custom{
			Key:  "name",
			Type: customInt,
		}.CreatePrompt(os.Stdin)
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
			Type:   customInt,
			MinInt: &min,
			MaxInt: &max,
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*promptui.Prompt)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Validate("7")).To(BeNil())
		Expect(errors.Is(underPrompt.Validate("3"), errIntTooLow)).To(BeTrue())
		Expect(errors.Is(underPrompt.Validate("25"), errIntTooHigh)).To(BeTrue())
	})

	It("can create custom enum prompt", func() {
		prompt, err := Custom{
			Type:        customEnum,
			EnumOptions: []string{"a", "b", "c"},
		}.CreatePrompt(os.Stdin)
		Expect(err).To(BeNil())

		underPrompt, ok := prompt.(*enumWrapper)
		Expect(ok).To(BeTrue())
		Expect(underPrompt.Select.Items).To(Equal([]string{"a", "b", "c"}))
	})
})

var _ = Describe("Replacement", func() {
	var (
		fs  afero.Fs
		afs afero.Afero
	)
	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		afs = afero.Afero{Fs: fs}
	})

	It("can find and replace contents in a file", func() {
		startData := `first line
second line
third line
ignore me`
		endData := `first line
replaced here
third line
ignore me`

		err := afs.WriteFile("file.txt", []byte(startData), os.ModeTemporary)
		Expect(err).To(BeNil())

		rep := Replacement{
			Path:    "file.txt",
			Find:    "second line",
			Replace: "replaced here",
		}
		err = rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{})
		Expect(err).To(BeNil())
		Expect("file.txt").To(HaveContents(afs, endData))
	})

	It("can find and replace with template", func() {
		startData := `{
  "name": "demo-project",
  "version": "1.0.0",
}`
		endData := `{
  "name": "demo-project",
  "version": "1.1.0",
}`

		err := afs.WriteFile("file.txt", []byte(startData), os.ModeTemporary)
		Expect(err).To(BeNil())

		rep := Replacement{
			Path:    "file.txt",
			Find:    `  "version": ".*",`,
			Replace: `  "version": "{{.VersionNoPrefix}}",`,
		}
		err = rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{
			VersionNoPrefix: "1.1.0",
		})
		Expect(err).To(BeNil())
		Expect("file.txt").To(HaveContents(afs, endData))
	})

	It("returns error on bad file read", func() {
		rep := Replacement{
			Path: "does not exist",
		}
		err := rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{})
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad template parsing", func() {
		rep := Replacement{
			Replace: "{{.bad..}}",
		}
		err := rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{})
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad template execute", func() {
		rep := Replacement{
			Replace: "{{.bad..}}",
		}
		err := rep.Execute(afs.ReadFile, afs.WriteFile, ReplaceData{})
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad write file", func() {
		startData := "some data"

		err := afs.WriteFile("file.txt", []byte(startData), os.ModeTemporary)
		Expect(err).To(BeNil())

		mockError := errors.New("bad write")
		badWrite := func(path string, data []byte, mode os.FileMode) error {
			return mockError
		}

		rep := Replacement{
			Path:    "file.txt",
			Find:    "some",
			Replace: "none",
		}
		err = rep.Execute(afs.ReadFile, badWrite, ReplaceData{})
		Expect(err).To(Equal(mockError))
	})
})
