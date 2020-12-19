package cmd

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"time"
)

var _ = Describe("Change", func() {
	It("should save an unreleased file", func() {
		config := Config{
			ChangesDir:    "Changes",
			UnreleasedDir: "Unrel",
		}
		change := Change{
			Kind: "kind",
			Body: "some body message",
		}

		mockTime := func() time.Time {
			return time.Date(2015, 5, 24, 3, 30, 10, 5, time.Local)
		}

		changesYaml := "kind: kind\nbody: some body message\n"

		writeCalled := false

		mockWf := newMockWriteFiler()
		mockWf.mockWriteFile = func(filepath string, bytes []byte, perm os.FileMode) error {
			writeCalled = true
			Expect(filepath).To(Equal("Changes/Unrel/kind-20150524-033010.yaml"))
			Expect(bytes).To(Equal([]byte(changesYaml)))
			return nil
		}

		err := change.SaveUnreleased(mockWf, mockTime, config)
		Expect(err).To(BeNil())
		Expect(writeCalled).To(Equal(true))
	})

	It("should load change from path", func() {
		mockRf := newMockReadFiler()
		mockRf.mockReadFile = func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal("some_file.yaml"))
			return []byte("kind: A\nbody: hey\n"), nil
		}

		change, err := LoadChange("some_file.yaml", mockRf)
		Expect(err).To(BeNil())
		Expect(change.Kind).To(Equal("A"))
		Expect(change.Body).To(Equal("hey"))
	})

	It("should return error from bad read", func() {
		mockErr := errors.New("bad file")

		mockRf := newMockReadFiler()
		mockRf.mockReadFile = func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal("some_file.yaml"))
			return []byte(""), mockErr
		}

		_, err := LoadChange("some_file.yaml", mockRf)
		Expect(err).To(Equal(mockErr))
	})

	It("should return error from bad file", func() {
		mockRf := newMockReadFiler()
		mockRf.mockReadFile = func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal("some_file.yaml"))
			return []byte("not a yaml file---"), nil
		}

		_, err := LoadChange("some_file.yaml", mockRf)
		Expect(err).NotTo(BeNil())
	})
})
