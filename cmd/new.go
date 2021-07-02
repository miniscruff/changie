package cmd

import (
	"io"
	"os"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/miniscruff/changie/core"
	"github.com/miniscruff/changie/shared"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new change file",
	Long: `Creates a new change file.
Change files are processed when batching a new release.
Each version is merged together for the overall project changelog.`,
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	afs := afero.Afero{Fs: fs}

	return newPipeline(afs, time.Now, os.Stdin)
}

func newPipeline(afs afero.Afero, tn shared.TimeNow, stdinReader io.ReadCloser) error {
	config, err := core.LoadConfig(afs.ReadFile)
	if err != nil {
		return err
	}

	var change core.Change

	err = core.AskPrompts(&change, config, stdinReader)
	if err != nil {
		return err
	}

	change.Time = tn()

	return change.SaveUnreleased(afs.WriteFile, config)
}
