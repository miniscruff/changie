package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("end to end", func() {
	var (
		stdinReader  *os.File
		stdinWriter  *os.File
		oldStdin     *os.File
		stdoutReader *os.File
		stdoutWriter *os.File
		oldStdout    *os.File
		inputSleep   = 50 * time.Millisecond
		startDir     string
		tempDir      string
	)

	BeforeEach(func() {
		var err error

		startDir, err = os.Getwd()
		Expect(err).To(BeNil())

		tempDir, err = os.MkdirTemp("", "e2e-test-"+CurrentGinkgoTestDescription().TestText)
		Expect(err).To(BeNil())

		err = os.Chdir(tempDir)
		Expect(err).To(BeNil())

		stdinReader, stdinWriter, err = os.Pipe()
		Expect(err).To(BeNil())

		oldStdin = os.Stdin
		os.Stdin = stdinReader

		stdoutReader, stdoutWriter, err = os.Pipe()
		Expect(err).To(BeNil())

		oldStdout = os.Stdout
		os.Stdout = stdoutWriter

		rootCmd.SetOut(stdoutWriter)
		rootCmd.SetIn(stdinReader)
	})

	AfterEach(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout

		err := os.Chdir(startDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(tempDir)
		Expect(err).To(BeNil())
	})

	testInit := func() {
		rootCmd.SetArgs([]string{"init"})
		Expect(Execute("")).To(Succeed())

		_, err := os.ReadFile(".changie.yaml")
		Expect(err).To(BeNil())
	}

	delayWrite := func(writer io.Writer, data []byte) {
		time.Sleep(inputSleep)
		_, err := writer.Write(data)
		Expect(err).To(BeNil())
	}

	testEcho := func(args []string, expect string) {
		rootCmd.SetArgs(args)
		Expect(Execute("")).To(Succeed())

		out := make([]byte, 1000)
		_, err := stdoutReader.Read(out)
		Expect(err).To(BeNil())
		Expect(string(out)).To(ContainSubstring(expect))
	}

	testNew := func(body string) {
		rootCmd.SetArgs([]string{"new"})
		go func() {
			delayWrite(stdinWriter, []byte{106, 13})
			delayWrite(stdinWriter, []byte(body))
			delayWrite(stdinWriter, []byte{13})
		}()
		Expect(Execute("")).To(Succeed())
	}

	testNewDry := func(body string) {
		go func() {
			delayWrite(stdinWriter, []byte{106, 13})
			delayWrite(stdinWriter, []byte(body))
			delayWrite(stdinWriter, []byte{13})
		}()

		date := time.Now().Format("2006-01-02")
		contents := fmt.Sprintf(`kind: Changed
body: show contents
time: %s`, date)
		testEcho([]string{"new", "--dry-run"}, contents)
		Expect(newCmd.Flags().Set("dry-run", "false")).To(Succeed())
	}

	testBatch := func() {
		rootCmd.SetArgs([]string{"batch", "v0.1.0"})
		Expect(Execute("")).To(Succeed())

		date := time.Now().Format("2006-01-02")
		changeContents := fmt.Sprintf(`## v0.1.0 - %s
### Changed
* newer
* older`, date)
		changeFile, err := os.ReadFile(".changes/v0.1.0.md")
		Expect(err).To(BeNil())
		Expect(string(changeFile)).To(Equal(changeContents))
	}

	testBatchDry := func() {
		date := time.Now().Format("2006-01-02")
		batchContents := fmt.Sprintf(`## v0.1.0 - %s
### Changed
* newer
* older`, date)
		testEcho([]string{"batch", "v0.1.0", "--dry-run"}, batchContents)
		Expect(batchCmd.Flags().Set("dry-run", "false")).To(Succeed())
	}

	testMerge := func() {
		rootCmd.SetArgs([]string{"merge"})
		Expect(Execute("")).To(Succeed())

		date := time.Now().Format("2006-01-02")
		changeContents := fmt.Sprintf(`%s

## v0.1.0 - %s
### Changed
* newer
* older`, defaultHeader, date)
		changeFile, err := os.ReadFile("CHANGELOG.md")
		Expect(err).To(BeNil())
		fmt.Println(string(changeFile))
		Expect(string(changeFile)).To(Equal(changeContents))
	}

	testGen := func() {
		docsPath := filepath.Join(tempDir, "website", "content", "cli")
		Expect(os.MkdirAll(docsPath, os.ModePerm)).To(Succeed())
		rootCmd.SetArgs([]string{"gen"})
		Expect(Execute("")).To(Succeed())

		// test a few command files exist
		Expect(filepath.Join(docsPath, "changie.md")).To(BeARegularFile())
		Expect(filepath.Join(docsPath, "changie_init.md")).To(BeARegularFile())
		Expect(filepath.Join(docsPath, "changie_batch.md")).To(BeARegularFile())
		Expect(filepath.Join(docsPath, "changie_merge.md")).To(BeARegularFile())
	}

	It("should all work", func() {
		testInit()
		testEcho([]string{"latest"}, "0.0.0")
		testNew("older")
		time.Sleep(2 * time.Second) // let time pass for the next change
		testNew("newer")
		testNewDry("show contents")
		testBatchDry()
		testBatch()
		testEcho([]string{"latest"}, "0.1.0")
		testEcho([]string{"next", "major"}, "1.0.0")
		testMerge()
		testGen()
	})

	It("should fail to batch on bad version", func() {
		rootCmd.SetArgs([]string{"batch", "not-a-semantic-version-$$$"})
		err := Execute("")
		Expect(err).NotTo(BeNil())
	})

	It("should fail on new with no config", func() {
		rootCmd.SetArgs([]string{"new"})
		err := Execute("")
		Expect(err).NotTo(BeNil())
	})

	It("should fail to find latest if you do not init", func() {
		rootCmd.SetArgs([]string{"latest"})
		err := Execute("")
		Expect(err).NotTo(BeNil())
	})

	It("should fail to find next if you do not init", func() {
		rootCmd.SetArgs([]string{"next", "patch"})
		err := Execute("")
		Expect(err).NotTo(BeNil())
	})

	It("should fail to echo next if bad input", func() {
		testInit()

		rootCmd.SetArgs([]string{"next", "blah-blah-blah"})
		err := Execute("")
		Expect(err).NotTo(BeNil())
	})
})
