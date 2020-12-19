package cmd

import (
	"errors"
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

		mockWf := newMockWriteFiler()
		mockWf.mockWriteFile = func(filepath string, bytes []byte, perm os.FileMode) error {
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
		mockRf := newMockReadFiler()
		mockRf.mockReadFile = func(filepath string) ([]byte, error) {
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

		mockRf := newMockReadFiler()
		mockRf.mockReadFile = func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal(configPath))
			return []byte(""), mockErr
		}

		_, err := LoadConfig(mockRf)
		Expect(err).To(Equal(mockErr))
	})

	It("should return error from bad file", func() {
		mockRf := newMockReadFiler()
		mockRf.mockReadFile = func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal(configPath))
			return []byte("not a yaml file---"), nil
		}

		_, err := LoadConfig(mockRf)
		Expect(err).NotTo(BeNil())
	})
})
