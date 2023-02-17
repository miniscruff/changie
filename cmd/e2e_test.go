package cmd

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/miniscruff/changie/then"
	"github.com/spf13/cobra"
)

func swapInReaderOutWriter(t *testing.T, cmd *cobra.Command, inReader, outWriter *os.File) {
	cmd.SetOut(outWriter)
	cmd.SetIn(inReader)

	batchDryRunOut = outWriter
	mergeDryRunOut = outWriter

	t.Cleanup(func() {
		cmd.SetOut(nil)
		cmd.SetIn(nil)
		batchDryRunOut = cmd.OutOrStdout()
		mergeDryRunOut = cmd.OutOrStdout()
	})
}

func testEcho(t *testing.T, cmd *cobra.Command, reader io.Reader, args []string, expect string) {
	cmd.SetArgs(args)
	then.Nil(t, cmd.Execute())

	out := make([]byte, 1000)
	_, err := reader.Read(out)

	then.Nil(t, err)
	then.Contains(t, expect, string(out))
}

func testInit(t *testing.T, cmd *cobra.Command) {
	cmd.SetArgs([]string{"init"})
	then.Nil(t, cmd.Execute())
	then.FileExists(t, ".changie.yaml")
}

func testNew(t *testing.T, cmd *cobra.Command, w io.Writer, body string) {
	cmd.SetArgs([]string{"new"})
	then.DelayWrite(
		t, w,
		[]byte{106, 13},
		[]byte(body),
		[]byte{13},
	)
	then.Nil(t, cmd.Execute())
}

func testBatch(t *testing.T, cmd *cobra.Command) {
	cmd.SetArgs([]string{"batch", "v0.1.0"})
	then.Nil(t, cmd.Execute())

	date := time.Now().Format("2006-01-02")
	changeContents := fmt.Sprintf(`## v0.1.0 - %s
### Changed
* older
* newer
`, date)

	then.FileContentsNoAfero(t, changeContents, ".changes", "v0.1.0.md")
}

func testMerge(t *testing.T, cmd *cobra.Command) {
	cmd.SetArgs([]string{"merge"})
	then.Nil(t, cmd.Execute())

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
    cmd := RootCmd()

	then.WithTempDir(t)
	inReader, inWriter, outReader, outWriter := then.WithStdInOut(t)
	swapInReaderOutWriter(t, cmd, inReader, outWriter)

	testInit(t, cmd)
	testEcho(t, cmd, outReader, []string{"latest"}, "0.0.0")
	testNew(t, cmd, inWriter, "older")
	time.Sleep(2 * time.Second) // let time pass for the next change
	testNew(t, cmd, inWriter, "newer")
	testBatch(t, cmd)
	testEcho(t, cmd, outReader, []string{"latest"}, "0.1.0")
	testEcho(t, cmd, outReader, []string{"next", "major"}, "1.0.0")
	testMerge(t, cmd)
}

func TestErrorNextBadInput(t *testing.T) {
    cmd := RootCmd()

	then.WithTempDir(t)
	testInit(t, cmd)

	cmd.SetArgs([]string{"next", "blah-blah-blah"})
	then.NotNil(t, cmd.Execute())
}

func TestErrorNextExactVersion(t *testing.T) {
    cmd := RootCmd()

	then.WithTempDir(t)
	testInit(t, cmd)

	cmd.SetArgs([]string{"next", "v1.2.3"})
	then.NotNil(t, cmd.Execute())
}
