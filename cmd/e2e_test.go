package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"time"
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

	const dirName = "test-dir"

	BeforeEach(func() {
		var err error

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

		startDir, err = os.Getwd()
		Expect(err).To(BeNil())

		tempDir, err = ioutil.TempDir("", "e2e-test")
		Expect(err).To(BeNil())

		err = os.Chdir(tempDir)
		Expect(err).To(BeNil())
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

	testNew := func(body string) {
		rootCmd.SetArgs([]string{"new"})
		go func() {
			time.Sleep(inputSleep)
			stdinWriter.Write([]byte{106, 13})

			time.Sleep(time.Second)
			stdinWriter.Write([]byte(body))
			stdinWriter.Write([]byte{13})
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

		changeContents := `# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.1.0 - 2020-12-26

### Changed
* first
* second`
		changeFile, err := ioutil.ReadFile("CHANGELOG.md")
		Expect(err).To(BeNil())
		Expect(string(changeFile)).To(Equal(changeContents))
	}

	It("should all work", func() {
		testInit()
		testNew("first")
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
