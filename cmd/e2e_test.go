package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
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

		tempDir, err = ioutil.TempDir("", "e2e-test")
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
	}

	delayWrite := func(writer io.Writer, data []byte) {
		time.Sleep(inputSleep)
		_, err := writer.Write(data)
		Expect(err).To(BeNil())
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

	testBatch := func() {
		rootCmd.SetArgs([]string{"batch", "v0.1.0"})
		Expect(Execute("")).To(Succeed())
	}

	testLatest := func() {
		rootCmd.SetArgs([]string{"latest"})
		Expect(Execute("")).To(Succeed())

		versionOut := make([]byte, 10)
		_, err := stdoutReader.Read(versionOut)
		Expect(err).To(BeNil())
		Expect(string(versionOut)).To(ContainSubstring("0.1.0"))
	}

	testMerge := func() {
		rootCmd.SetArgs([]string{"merge"})
		Expect(Execute("")).To(Succeed())

		date := time.Now().Format("2006-01-02")
		changeContents := fmt.Sprintf(`%s

## v0.1.0 - %s

### Changed
* first
* second`, defaultHeader, date)
		changeFile, err := ioutil.ReadFile("CHANGELOG.md")
		Expect(err).To(BeNil())
		Expect(string(changeFile)).To(Equal(changeContents))
	}

	It("should all work", func() {
		testInit()
		testNew("first")
		time.Sleep(2 * time.Second) // let time pass for the next change
		testNew("second")
		testBatch()
		testLatest()
		testMerge()
	})

	It("should fail to find latest when no changes", func() {
		rootCmd.SetArgs([]string{"latest"})
		err := Execute("")
		Expect(err).NotTo(BeNil())
	})
})
