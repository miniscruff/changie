package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/miniscruff/changie/core"
	. "github.com/miniscruff/changie/test_utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("New", func() {
	mockTime := func() time.Time {
		return time.Date(2021, 5, 22, 13, 30, 10, 5, time.UTC)
	}

	var (
		fs         *MockFS
		afs        afero.Afero
		testConfig core.Config
	)
	BeforeEach(func() {
		fs = NewMockFS()
		afs = afero.Afero{Fs: fs}
		testConfig = core.Config{
			ChangesDir:    "news",
			UnreleasedDir: "future",
			HeaderPath:    "header.rst",
			ChangelogPath: "news.md",
			VersionExt:    "md",
			VersionFormat: "## {{.Version}}",
			KindFormat:    "### {{.Kind}}",
			ChangeFormat:  "* {{.Body}}",
			Kinds: []string{
				"added", "removed", "other",
			},
		}
	})

	It("creates new file on completion of prompts", func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			DelayWrite(stdinWriter, []byte{106, 13})
			DelayWrite(stdinWriter, []byte("a message"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(err).To(BeNil())

		futurePath := filepath.Join(testConfig.ChangesDir, testConfig.UnreleasedDir)
		fileInfos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(fileInfos).To(HaveLen(1))
		Expect(fileInfos[0].Name()).To(HaveSuffix(".yaml"))

		changeContent := fmt.Sprintf(
			"kind: removed\nbody: a message\ntime: %s\n",
			mockTime().Format(time.RFC3339Nano),
		)
		changePath := filepath.Join(futurePath, fileInfos[0].Name())
		Expect(changePath).To(HaveContents(afs, changeContent))
	})

	It("creates new file on completion of prompts with custom options", func() {
		testConfig.CustomChoices = []core.Custom{
			{
				Key:  "Issue",
				Type: core.CustomInt,
			},
			{
				Key:         "Emoji",
				Type:        core.CustomEnum,
				EnumOptions: []string{"rocket", "dog"},
			},
		}
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			DelayWrite(stdinWriter, []byte{106, 13})
			DelayWrite(stdinWriter, []byte("body stuff"))
			DelayWrite(stdinWriter, []byte{13})
			DelayWrite(stdinWriter, []byte("15"))
			DelayWrite(stdinWriter, []byte{13})
			DelayWrite(stdinWriter, []byte{106, 13})
		}()

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(err).To(BeNil())

		futurePath := filepath.Join(testConfig.ChangesDir, testConfig.UnreleasedDir)
		fileInfos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(fileInfos).To(HaveLen(1))
		Expect(fileInfos[0].Name()).To(HaveSuffix(".yaml"))

		changeContent := fmt.Sprintf(`kind: removed
body: body stuff
time: %s
custom:
  Emoji: dog
  Issue: "15"
`, mockTime().Format(time.RFC3339Nano))
		changePath := filepath.Join(futurePath, fileInfos[0].Name())
		Expect(changePath).To(HaveContents(afs, changeContent))
	})

	It("creates new file on completion of prompts with components", func() {
		testConfig.Components = []string{"tools", "compiler", "linker"}
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			DelayWrite(stdinWriter, []byte{106, 106, 13})
			DelayWrite(stdinWriter, []byte{13})
			DelayWrite(stdinWriter, []byte("body stuff"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(err).To(BeNil())

		futurePath := filepath.Join(testConfig.ChangesDir, testConfig.UnreleasedDir)
		fileInfos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(fileInfos).To(HaveLen(1))
		Expect(fileInfos[0].Name()).To(HaveSuffix(".yaml"))

		changeContent := fmt.Sprintf(`component: linker
kind: added
body: body stuff
time: %s
`, mockTime().Format(time.RFC3339Nano))
		changePath := filepath.Join(futurePath, fileInfos[0].Name())
		Expect(changePath).To(HaveContents(afs, changeContent))
	})

	It("creates new file with just a body", func() {
		testConfig.Kinds = []string{}
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			DelayWrite(stdinWriter, []byte("body stuff"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(err).To(BeNil())

		futurePath := filepath.Join(testConfig.ChangesDir, testConfig.UnreleasedDir)
		fileInfos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(fileInfos).To(HaveLen(1))
		Expect(fileInfos[0].Name()).To(HaveSuffix(".yaml"))

		changeContent := fmt.Sprintf(
			"body: body stuff\ntime: %s\n",
			mockTime().Format(time.RFC3339Nano),
		)
		changePath := filepath.Join(futurePath, fileInfos[0].Name())
		Expect(changePath).To(HaveContents(afs, changeContent))
	})

	It("returns error on bad component", func() {
		testConfig.Components = []string{"A", "B"}
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		DelayWrite(stdinWriter, []byte{3}) // 3 is ctrl-c

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad kind", func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		DelayWrite(stdinWriter, []byte{3}) // 3 is ctrl-c

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad body", func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			DelayWrite(stdinWriter, []byte{13})
			DelayWrite(stdinWriter, []byte("ctrl-c here"))
			DelayWrite(stdinWriter, []byte{3})
		}()

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad custom creation", func() {
		testConfig.CustomChoices = []core.Custom{
			{
				Key:  "name",
				Type: "bad type",
			},
		}
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			DelayWrite(stdinWriter, []byte{106, 13})
			DelayWrite(stdinWriter, []byte("body stuff"))
			DelayWrite(stdinWriter, []byte{13})
		}()

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(errors.Is(err, core.ErrInvalidPromptType)).To(BeTrue())
	})

	It("returns error on bad input choice", func() {
		testConfig.CustomChoices = []core.Custom{
			{
				Key:  "name",
				Type: core.CustomString,
			},
		}
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			DelayWrite(stdinWriter, []byte{106, 13})
			DelayWrite(stdinWriter, []byte("body stuff"))
			DelayWrite(stdinWriter, []byte{13})
			DelayWrite(stdinWriter, []byte("jonny"))
			DelayWrite(stdinWriter, []byte{3})
		}()

		err = newPipeline(afs, mockTime, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad config", func() {
		err := newPipeline(afs, mockTime, os.Stdin)
		Expect(err).NotTo(BeNil())
	})
})
