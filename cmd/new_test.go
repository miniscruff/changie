package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/miniscruff/changie/core"
	. "github.com/miniscruff/changie/testutils"
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
			Kinds: []core.KindConfig{
				{Label: "added"},
				{Label: "removed"},
				{Label: "other"},
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

	It("creates new file with just a body", func() {
		testConfig.Kinds = []core.KindConfig{}
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

	It("returns error on bad config", func() {
		err := newPipeline(afs, mockTime, os.Stdin)
		Expect(err).NotTo(BeNil())
	})
})
