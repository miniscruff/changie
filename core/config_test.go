package core

import (
	"errors"
	"os"

	. "github.com/miniscruff/changie/test_utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
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
			Expect(filepath).To(Equal(ConfigPath))
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
			Kinds:           []string{"B", "C", "E"},
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
- B
- C
- E
custom:
- key: first
  type: string
  label: First name
`

		writeCalled := false

		mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
			writeCalled = true
			Expect(filepath).To(Equal(ConfigPath))
			Expect(string(bytes)).To(Equal(configYaml))
			return nil
		}

		err := config.Save(mockWf)
		Expect(err).To(BeNil())
		Expect(writeCalled).To(Equal(true))
	})

	It("should load change from path", func() {
		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal(ConfigPath))
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
			Expect(filepath).To(Equal(ConfigPath))
			return []byte(""), mockErr
		}

		_, err := LoadConfig(mockRf)
		Expect(err).To(Equal(mockErr))
	})

	It("should return error from bad file", func() {
		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal(ConfigPath))
			return []byte("not a yaml file---"), nil
		}

		_, err := LoadConfig(mockRf)
		Expect(err).NotTo(BeNil())
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
