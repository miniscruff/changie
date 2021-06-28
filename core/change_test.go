package core

import (
	"errors"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Change", func() {
	mockTime := func() time.Time {
		return time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local)
	}

	orderedTimes := []time.Time{
		time.Date(2018, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2017, 5, 24, 3, 30, 10, 5, time.Local),
		time.Date(2016, 5, 24, 3, 30, 10, 5, time.Local),
	}

	It("should save an unreleased file", func() {
		config := Config{
			ChangesDir:    "Changes",
			UnreleasedDir: "Unrel",
		}
		change := Change{
			Body: "some body message",
			Time: mockTime(),
		}

		changesYaml := fmt.Sprintf(
			"body: some body message\ntime: %s\n",
			mockTime().Format(time.RFC3339Nano),
		)

		writeCalled := false

		mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
			writeCalled = true
			Expect(filepath).To(Equal("Changes/Unrel/20160524-033010.yaml"))
			Expect(string(bytes)).To(Equal(changesYaml))
			return nil
		}

		err := change.SaveUnreleased(mockWf, config)
		Expect(err).To(BeNil())
		Expect(writeCalled).To(Equal(true))
	})

	It("should save an unreleased file with optionals", func() {
		config := Config{
			ChangesDir:    "Changes",
			UnreleasedDir: "Unrel",
		}
		change := Change{
			Component: "comp",
			Kind:      "kind",
			Body:      "some body message",
			Time:      mockTime(),
		}

		changesYaml := fmt.Sprintf(
			"component: comp\nkind: kind\nbody: some body message\ntime: %s\n",
			mockTime().Format(time.RFC3339Nano),
		)

		writeCalled := false

		mockWf := func(filepath string, bytes []byte, perm os.FileMode) error {
			writeCalled = true
			Expect(filepath).To(Equal("Changes/Unrel/comp-kind-20160524-033010.yaml"))
			Expect(string(bytes)).To(Equal(changesYaml))
			return nil
		}

		err := change.SaveUnreleased(mockWf, config)
		Expect(err).To(BeNil())
		Expect(writeCalled).To(Equal(true))
	})

	It("should load change from path", func() {
		mockRf := func(filepath string) ([]byte, error) {
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

		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal("some_file.yaml"))
			return []byte(""), mockErr
		}

		_, err := LoadChange("some_file.yaml", mockRf)
		Expect(err).To(Equal(mockErr))
	})

	It("should return error from bad file", func() {
		mockRf := func(filepath string) ([]byte, error) {
			Expect(filepath).To(Equal("some_file.yaml"))
			return []byte("not a yaml file---"), nil
		}

		_, err := LoadChange("some_file.yaml", mockRf)
		Expect(err).NotTo(BeNil())
	})

	It("should sort by time", func() {
		changes := []Change{
			{Body: "third", Time: orderedTimes[2]},
			{Body: "second", Time: orderedTimes[1]},
			{Body: "first", Time: orderedTimes[0]},
		}
		config := Config{}
		SortByConfig(config).Sort(changes)

		Expect(changes[0].Body).To(Equal("first"))
		Expect(changes[1].Body).To(Equal("second"))
		Expect(changes[2].Body).To(Equal("third"))
	})

	It("should sort by kind then time", func() {
		config := Config{
			Kinds: []string{"A", "B"},
		}
		changes := []Change{
			{Body: "fourth", Kind: "B", Time: orderedTimes[1]},
			{Body: "second", Kind: "A", Time: orderedTimes[2]},
			{Body: "third", Kind: "B", Time: orderedTimes[0]},
			{Body: "first", Kind: "A", Time: orderedTimes[1]},
		}
		SortByConfig(config).Sort(changes)

		Expect(changes[0].Body).To(Equal("first"))
		Expect(changes[1].Body).To(Equal("second"))
		Expect(changes[2].Body).To(Equal("third"))
		Expect(changes[3].Body).To(Equal("fourth"))
	})

	It("should sort by component then kind", func() {
		config := Config{
			Kinds:      []string{"D", "E"},
			Components: []string{"A", "B", "C"},
		}
		changes := []Change{
			{Body: "fourth", Component: "B", Kind: "E"},
			{Body: "second", Component: "A", Kind: "E"},
			{Body: "third", Component: "B", Kind: "D"},
			{Body: "first", Component: "A", Kind: "D"},
		}
		SortByConfig(config).Sort(changes)

		Expect(changes[0].Body).To(Equal("first"))
		Expect(changes[1].Body).To(Equal("second"))
		Expect(changes[2].Body).To(Equal("third"))
		Expect(changes[3].Body).To(Equal("fourth"))
	})
})
