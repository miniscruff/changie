package cmd

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"time"
)

var _ = Describe("New", func() {
	var (
		fs         *mockFs
		afs        afero.Afero
		testConfig Config
		inputSleep = 50 * time.Millisecond
	)
	BeforeEach(func() {
		fs = newMockFs()
		afs = afero.Afero{Fs: fs}
		testConfig = Config{
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
		testConfig.Save(afs.WriteFile)

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			time.Sleep(inputSleep)
			stdinWriter.Write([]byte{106, 13})

			time.Sleep(time.Second)
			stdinWriter.Write([]byte("a message"))
			stdinWriter.Write([]byte{13})
		}()

		err = newPipeline(afs, stdinReader)
		Expect(err).To(BeNil())

		futurePath := filepath.Join(testConfig.ChangesDir, testConfig.UnreleasedDir)
		fileInfos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(fileInfos).To(HaveLen(1))
		Expect(fileInfos[0].Name()).To(HaveSuffix(".yaml"))

		changeContent := "kind: removed\nbody: a message\n"
		changePath := filepath.Join(futurePath, fileInfos[0].Name())
		Expect(changePath).To(HaveContents(afs, changeContent))
	})

	It("creates new file on completion of prompts with custom options", func() {
		testConfig.CustomChoices = map[string]Custom{
			"emoji": Custom{
				Type:        CustomEnum,
				EnumOptions: []string{"rocket", "dog"},
			},
			"name": Custom{
				Type:  CustomString,
				Label: "your name",
			},
		}
		testConfig.Save(afs.WriteFile)

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			time.Sleep(inputSleep)
			stdinWriter.Write([]byte{106, 13})

			time.Sleep(inputSleep)
			stdinWriter.Write([]byte("body stuff"))
			stdinWriter.Write([]byte{13})

			time.Sleep(inputSleep)
			stdinWriter.Write([]byte{106, 13})

			time.Sleep(inputSleep)
			stdinWriter.Write([]byte("lord muffins"))
			stdinWriter.Write([]byte{13})
		}()

		err = newPipeline(afs, stdinReader)
		Expect(err).To(BeNil())

		futurePath := filepath.Join(testConfig.ChangesDir, testConfig.UnreleasedDir)
		fileInfos, err := afs.ReadDir(futurePath)
		Expect(err).To(BeNil())
		Expect(fileInfos).To(HaveLen(1))
		Expect(fileInfos[0].Name()).To(HaveSuffix(".yaml"))

		changeContent := `kind: removed
body: body stuff
custom:
  emoji: dog
  name: lord muffins
`
		changePath := filepath.Join(futurePath, fileInfos[0].Name())
		Expect(changePath).To(HaveContents(afs, changeContent))
	})

	It("returns error on bad kind", func() {
		testConfig.Save(afs.WriteFile)

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		stdinWriter.Write([]byte{3}) // 3 is ctrl-c

		err = newPipeline(afs, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad body", func() {
		testConfig.Save(afs.WriteFile)

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			stdinWriter.Write([]byte{13})
			time.Sleep(inputSleep)
			stdinWriter.Write([]byte("ctrl-c here"))
			stdinWriter.Write([]byte{3})
		}()

		err = newPipeline(afs, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad custom creation", func() {
		testConfig.CustomChoices = map[string]Custom{
			"name": Custom{
				Type: "bad type",
			},
		}
		testConfig.Save(afs.WriteFile)

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			time.Sleep(inputSleep)
			stdinWriter.Write([]byte{106, 13})

			time.Sleep(inputSleep)
			stdinWriter.Write([]byte("body stuff"))
			stdinWriter.Write([]byte{13})
		}()

		err = newPipeline(afs, stdinReader)
		Expect(errors.Is(err, invalidPromptType)).To(BeTrue())
	})

	It("returns error on bad input choice", func() {
		testConfig.CustomChoices = map[string]Custom{
			"name": Custom{
				Type: CustomString,
			},
		}
		testConfig.Save(afs.WriteFile)

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			time.Sleep(inputSleep)
			stdinWriter.Write([]byte{106, 13})

			time.Sleep(inputSleep)
			stdinWriter.Write([]byte("body stuff"))
			stdinWriter.Write([]byte{13})

			time.Sleep(inputSleep)
			stdinWriter.Write([]byte("jonny"))
			stdinWriter.Write([]byte{3})
		}()

		err = newPipeline(afs, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad config", func() {
		err := newPipeline(afs, os.Stdin)
		Expect(err).NotTo(BeNil())
	})
})
