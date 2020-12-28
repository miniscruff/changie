package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

func delayWrite(writer io.Writer, data []byte) {
	time.Sleep(50 * time.Millisecond)

	_, err := writer.Write(data)
	Expect(err).To(BeNil())
}

var _ = Describe("New", func() {
	var (
		fs         *mockFs
		afs        afero.Afero
		testConfig Config
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
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			delayWrite(stdinWriter, []byte{106, 13})
			delayWrite(stdinWriter, []byte("a message"))
			delayWrite(stdinWriter, []byte{13})
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
		testConfig.CustomChoices = []Custom{
			{
				Key:  "Issue",
				Type: customInt,
			},
			{
				Key:         "Emoji",
				Type:        customEnum,
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
			delayWrite(stdinWriter, []byte{106, 13})
			delayWrite(stdinWriter, []byte("body stuff"))
			delayWrite(stdinWriter, []byte{13})
			delayWrite(stdinWriter, []byte("15"))
			delayWrite(stdinWriter, []byte{13})
			delayWrite(stdinWriter, []byte{106, 13})
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
  Emoji: dog
  Issue: "15"
`
		changePath := filepath.Join(futurePath, fileInfos[0].Name())
		Expect(changePath).To(HaveContents(afs, changeContent))
	})

	It("returns error on bad kind", func() {
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		delayWrite(stdinWriter, []byte{3}) // 3 is ctrl-c

		err = newPipeline(afs, stdinReader)
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
			delayWrite(stdinWriter, []byte{13})
			delayWrite(stdinWriter, []byte("ctrl-c here"))
			delayWrite(stdinWriter, []byte{3})
		}()

		err = newPipeline(afs, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad custom creation", func() {
		testConfig.CustomChoices = []Custom{
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
			delayWrite(stdinWriter, []byte{106, 13})
			delayWrite(stdinWriter, []byte("body stuff"))
			delayWrite(stdinWriter, []byte{13})
		}()

		err = newPipeline(afs, stdinReader)
		Expect(errors.Is(err, errInvalidPromptType)).To(BeTrue())
	})

	It("returns error on bad input choice", func() {
		testConfig.CustomChoices = []Custom{
			{
				Key:  "name",
				Type: customString,
			},
		}
		err := testConfig.Save(afs.WriteFile)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err := os.Pipe()
		Expect(err).To(BeNil())

		defer stdinReader.Close()
		defer stdinWriter.Close()

		go func() {
			delayWrite(stdinWriter, []byte{106, 13})
			delayWrite(stdinWriter, []byte("body stuff"))
			delayWrite(stdinWriter, []byte{13})
			delayWrite(stdinWriter, []byte("jonny"))
			delayWrite(stdinWriter, []byte{3})
		}()

		err = newPipeline(afs, stdinReader)
		Expect(err).NotTo(BeNil())
	})

	It("returns error on bad config", func() {
		err := newPipeline(afs, os.Stdin)
		Expect(err).NotTo(BeNil())
	})
})
