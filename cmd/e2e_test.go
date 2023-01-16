package cmd

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/miniscruff/changie/then"
)

func swapInReaderOutWriter(t *testing.T, inReader, outWriter *os.File) {
	rootCmd.SetOut(outWriter)
	rootCmd.SetIn(inReader)

	batchDryRunOut = outWriter
	mergeDryRunOut = outWriter

	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetIn(nil)
		batchDryRunOut = rootCmd.OutOrStdout()
		mergeDryRunOut = rootCmd.OutOrStdout()
	})
}

func testEcho(t *testing.T, reader io.Reader, args []string, expect string) {
	rootCmd.SetArgs(args)
	then.Nil(t, Execute(""))

	out := make([]byte, 1000)
	_, err := reader.Read(out)

	then.Nil(t, err)
	then.Contains(t, expect, string(out))
}

func testInit(t *testing.T) {
	rootCmd.SetArgs([]string{"init"})
	then.Nil(t, Execute(""))
	then.FileExists(t, ".changie.yaml")
}

func testNew(t *testing.T, w io.Writer, body string) {
	rootCmd.SetArgs([]string{"new"})
	then.DelayWrite(
		t, w,
		[]byte{106, 13},
		[]byte(body),
		[]byte{13},
	)
	then.Nil(t, Execute(""))
}

func testBatch(t *testing.T) {
	rootCmd.SetArgs([]string{"batch", "v0.1.0"})
	then.Nil(t, Execute(""))

	date := time.Now().Format("2006-01-02")
	changeContents := fmt.Sprintf(`## v0.1.0 - %s
### Changed
* older
* newer
`, date)

	then.FileContentsNoAfero(t, changeContents, ".changes", "v0.1.0.md")
}

func testMerge(t *testing.T) {
	rootCmd.SetArgs([]string{"merge"})
	then.Nil(t, Execute(""))

	date := time.Now().Format("2006-01-02")
	changeContents := fmt.Sprintf(`%s

## v0.1.0 - %s
### Changed
* older
* newer
`, defaultHeader, date)
	then.FileContentsNoAfero(t, changeContents, "CHANGELOG.md")
}

func TestFullRun(t *testing.T) {
	then.WithTempDir(t)
	inReader, inWriter, outReader, outWriter := then.WithStdInOut(t)
	swapInReaderOutWriter(t, inReader, outWriter)

	testInit(t)
	testEcho(t, outReader, []string{"latest"}, "0.0.0")
	testNew(t, inWriter, "older")
	time.Sleep(2 * time.Second) // let time pass for the next change
	testNew(t, inWriter, "newer")
	testBatch(t)
	testEcho(t, outReader, []string{"latest"}, "0.1.0")
	testEcho(t, outReader, []string{"next", "major"}, "1.0.0")
	testMerge(t)
}

func TestErrorNextBadInput(t *testing.T) {
	then.WithTempDir(t)
	testInit(t)

	rootCmd.SetArgs([]string{"next", "blah-blah-blah"})
	then.NotNil(t, Execute(""))
}

func TestErrorNextExactVersion(t *testing.T) {
	then.WithTempDir(t)
	testInit(t)

	rootCmd.SetArgs([]string{"next", "v1.2.3"})
	then.NotNil(t, Execute(""))
}
